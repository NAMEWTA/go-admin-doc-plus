package desktop

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

var (
	ErrMissingAssets  = errors.New("desktop Admin assets are missing")
	ErrAlreadyRun     = errors.New("desktop host has already run")
	ErrUnsafeListener = errors.New("desktop listener is not loopback-only")
)

const (
	defaultRequestTimeout  = 10 * time.Second
	defaultShutdownTimeout = 5 * time.Second
)

type Application interface {
	Handler() http.Handler
	Start(context.Context) error
	Stop(context.Context) error
}

type Runtime struct {
	Application Application
	Close       func(context.Context) error
}

type Builder func(context.Context) (Runtime, error)

type WindowRunner func(context.Context, *Bridge, fs.FS) error

type Config struct {
	Assets          fs.FS
	Version         string
	RequestTimeout  time.Duration
	ShutdownTimeout time.Duration
}

type Host struct {
	config Config
	build  Builder
	window WindowRunner
	listen func(context.Context) (net.Listener, error)
	run    atomic.Bool
}

func New(config Config, build Builder, window WindowRunner) (*Host, error) {
	if config.Assets == nil {
		return nil, ErrMissingAssets
	}
	if info, err := fs.Stat(config.Assets, "index.html"); err != nil || info.IsDir() {
		return nil, ErrMissingAssets
	}
	if build == nil {
		return nil, errors.New("desktop runtime builder is required")
	}
	if window == nil {
		return nil, errors.New("desktop window runner is required")
	}
	if config.RequestTimeout < 0 || config.ShutdownTimeout < 0 {
		return nil, errors.New("desktop timeouts cannot be negative")
	}
	if config.RequestTimeout == 0 {
		config.RequestTimeout = defaultRequestTimeout
	}
	if config.ShutdownTimeout == 0 {
		config.ShutdownTimeout = defaultShutdownTimeout
	}
	config.Version = strings.TrimSpace(config.Version)
	if config.Version == "" {
		config.Version = "dev"
	}
	return &Host{
		config: config,
		build:  build,
		window: window,
		listen: listenLoopback,
	}, nil
}

func (host *Host) Run(ctx context.Context) (runErr error) {
	if ctx == nil {
		return errors.New("desktop host context is required")
	}
	if !host.run.CompareAndSwap(false, true) {
		return ErrAlreadyRun
	}
	runtime, err := host.build(ctx)
	if err != nil {
		return fmt.Errorf("build desktop runtime: %w", err)
	}
	if runtime.Application == nil {
		return errors.Join(errors.New("desktop runtime application is required"), closeRuntime(host.config.ShutdownTimeout, runtime.Close))
	}
	defer func() {
		runErr = errors.Join(runErr, closeRuntime(host.config.ShutdownTimeout, runtime.Close))
	}()

	listener, err := host.listen(ctx)
	if err != nil {
		return fmt.Errorf("listen on desktop loopback: %w", err)
	}
	listenerOwned := true
	defer func() {
		if listenerOwned {
			runErr = errors.Join(runErr, listener.Close())
		}
	}()
	if err := validateLoopbackListener(listener); err != nil {
		return err
	}
	token, err := generateLaunchToken()
	if err != nil {
		return fmt.Errorf("generate desktop launch token: %w", err)
	}
	gateway, err := NewGateway(token)
	if err != nil {
		return err
	}
	if err := runtime.Application.Start(ctx); err != nil {
		return fmt.Errorf("start desktop application: %w", err)
	}
	applicationStarted := true

	httpServer := &http.Server{
		Handler:           gateway.Wrap(runtime.Application.Handler()),
		ReadHeaderTimeout: host.config.RequestTimeout,
		IdleTimeout:       host.config.RequestTimeout,
	}
	serveErrors := make(chan error, 1)
	go func() { serveErrors <- httpServer.Serve(listener) }()
	bridge := newBridge("http://"+listener.Addr().String(), token, host.config.Version, host.config.RequestTimeout)
	windowErr := host.window(ctx, bridge, host.config.Assets)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), host.config.ShutdownTimeout)
	defer cancel()
	serverStopErr := httpServer.Shutdown(shutdownCtx)
	listenerOwned = false
	var serveErr error
	select {
	case serveErr = <-serveErrors:
		if errors.Is(serveErr, http.ErrServerClosed) {
			serveErr = nil
		}
	case <-shutdownCtx.Done():
		serveErr = shutdownCtx.Err()
	}
	var applicationStopErr error
	if applicationStarted {
		applicationStopErr = runtime.Application.Stop(shutdownCtx)
	}
	return errors.Join(wrap("run desktop window", windowErr), wrap("serve desktop HTTP", serveErr), serverStopErr, applicationStopErr)
}

func validateLoopbackListener(listener net.Listener) error {
	if listener == nil || listener.Addr() == nil {
		return ErrUnsafeListener
	}
	host, port, err := net.SplitHostPort(listener.Addr().String())
	if err != nil || port == "" {
		return ErrUnsafeListener
	}
	ip := net.ParseIP(host)
	if ip == nil || !ip.IsLoopback() {
		return ErrUnsafeListener
	}
	return nil
}

func listenLoopback(ctx context.Context) (net.Listener, error) {
	var config net.ListenConfig
	return config.Listen(ctx, "tcp4", "127.0.0.1:0")
}

func generateLaunchToken() (string, error) {
	value := make([]byte, 32)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}

func closeRuntime(timeout time.Duration, close func(context.Context) error) error {
	if close == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return wrap("close desktop runtime", close(ctx))
}

func wrap(operation string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", operation, err)
}
