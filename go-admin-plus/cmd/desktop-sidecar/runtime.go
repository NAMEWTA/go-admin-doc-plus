package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go-admin/internal/app/kernel"
	"go-admin/internal/modules/audit"
	auditmigration "go-admin/internal/modules/audit/migrations/0011-audit"
	"go-admin/internal/modules/demo"
	productsmigration "go-admin/internal/modules/demo/migrations/0010-products"
	"go-admin/internal/modules/iam/administration"
	"go-admin/internal/modules/iam/authorization"
	sessionmigration "go-admin/internal/modules/iam/migrations/0010-session-schema"
	administrationmigration "go-admin/internal/modules/iam/migrations/0020-administration-schema"
	"go-admin/internal/modules/iam/session"
	"go-admin/internal/platform/config"
	"go-admin/internal/platform/database"
	desktopplatform "go-admin/internal/platform/desktop"
	"go-admin/internal/platform/migrations"
	reliablemigration "go-admin/internal/platform/migrations/reliable-runtime"
)

const (
	requestTimeout  = 15 * time.Second
	shutdownTimeout = 5 * time.Second
)

type sidecarRuntime struct {
	material desktopplatform.LaunchMaterial
	kernel   *kernel.Kernel
	database *database.Database
	backup   desktopplatform.DatabaseBackup
	instance *desktopplatform.InstanceLock
	listener net.Listener
	server   *http.Server
	sessions *session.Service
	serveErr chan error
	stopOnce sync.Once
	stop     context.CancelFunc
}

func newSidecarRuntime(material desktopplatform.LaunchMaterial, stop context.CancelFunc) (*sidecarRuntime, error) {
	if err := material.Validate(); err != nil || stop == nil {
		return nil, errors.New("desktop sidecar dependencies are invalid")
	}
	runtime := &sidecarRuntime{material: material, serveErr: make(chan error, 1), stop: stop}
	owner, err := kernel.New(
		kernel.Lifecycle{Name: "desktop-instance", Start: runtime.startInstance, Drain: noopLifecycle, Stop: runtime.stopInstance},
		kernel.Lifecycle{Name: "desktop-database", Start: runtime.startDatabase, Drain: noopLifecycle, Stop: runtime.stopDatabase},
		kernel.Lifecycle{Name: "desktop-http", Start: runtime.startHTTP, Drain: runtime.drainHTTP, Stop: runtime.stopHTTP},
	)
	if err != nil {
		return nil, err
	}
	runtime.kernel = owner
	return runtime, nil
}

func (runtime *sidecarRuntime) startInstance(context.Context) error {
	if err := secureDirectories(runtime.material.DataDirectory, runtime.material.LogDirectory); err != nil {
		return err
	}
	lock, err := desktopplatform.AcquireSecureInstanceLock(runtime.material.DataDirectory)
	if err != nil {
		return errors.New("desktop instance lock failed")
	}
	runtime.instance = lock
	return nil
}

func (runtime *sidecarRuntime) stopInstance(context.Context) error {
	if runtime.instance == nil {
		return nil
	}
	err := runtime.instance.Close()
	runtime.instance = nil
	return err
}

func (runtime *sidecarRuntime) startDatabase(ctx context.Context) error {
	databasePath := filepath.Join(runtime.material.DataDirectory, "go-admin-plus.db")
	backup, err := desktopplatform.BackupDatabase(databasePath, filepath.Join(runtime.material.DataDirectory, "backups"), time.Now())
	if err != nil {
		return errors.New("desktop database backup failed")
	}
	runtime.backup = backup
	process := database.NewProcess()
	db, err := process.Open(ctx, database.Config{Profile: config.ProfileDesktopSQLite, SQLitePath: databasePath})
	if err != nil {
		return databaseStartupFailure(nil, runtime.backup.Restore, "desktop database open failed")
	}
	runtime.database = db
	runner, err := migrations.NewRunner(
		reliablemigration.Provider{}, sessionmigration.Provider{}, administrationmigration.Provider{},
		productsmigration.Provider{}, auditmigration.Provider{},
	)
	if err == nil {
		_, err = runner.Up(ctx, db)
	}
	if err == nil {
		var registry *authorization.CapabilityRegistry
		registry, err = authorization.NewCapabilityRegistry(db)
		if err == nil {
			err = demo.RegisterCapabilities(ctx, registry)
		}
	}
	if err != nil {
		return runtime.recoverDatabase()
	}
	return nil
}

func (runtime *sidecarRuntime) recoverDatabase() error {
	if runtime.database == nil {
		return errors.New("desktop database recovery failed")
	}
	database := runtime.database
	runtime.database = nil
	return databaseStartupFailure(database.Close, runtime.backup.Restore, "desktop database migration failed")
}

func databaseStartupFailure(closeDatabase, restoreDatabase func() error, cause string) error {
	var closeErr error
	if closeDatabase != nil {
		closeErr = closeDatabase()
	}
	if closeErr != nil {
		return errors.New("desktop database recovery failed")
	}
	restoreErr := restoreDatabase()
	if restoreErr != nil {
		return errors.New("desktop database recovery failed")
	}
	return errors.New(cause)
}

func (runtime *sidecarRuntime) stopDatabase(context.Context) error {
	if runtime.database == nil {
		return nil
	}
	err := runtime.database.Close()
	runtime.database = nil
	return err
}

func (runtime *sidecarRuntime) startHTTP(ctx context.Context) error {
	handler, err := runtime.buildHandler()
	if err != nil {
		return err
	}
	var listenerConfig net.ListenConfig
	listener, err := listenerConfig.Listen(ctx, "tcp4", "127.0.0.1:0")
	if err != nil {
		return errors.New("desktop loopback listener failed")
	}
	if !listener.Addr().(*net.TCPAddr).IP.IsLoopback() {
		_ = listener.Close()
		return errors.New("desktop loopback listener is unsafe")
	}
	runtime.listener = listener
	runtime.server = &http.Server{
		Handler: handler, ReadHeaderTimeout: requestTimeout, ReadTimeout: requestTimeout,
		WriteTimeout: requestTimeout, IdleTimeout: requestTimeout,
	}
	go func() { runtime.serveErr <- runtime.server.Serve(listener) }()
	return nil
}

func (runtime *sidecarRuntime) port() uint16 {
	if runtime.listener == nil {
		return 0
	}
	return uint16(runtime.listener.Addr().(*net.TCPAddr).Port)
}

func (runtime *sidecarRuntime) buildHandler() (http.Handler, error) {
	loginFacts, err := audit.NewSessionLoginFactAdapter(runtime.database)
	if err != nil {
		return nil, errors.New("desktop login audit adapter failed")
	}
	sessions, err := session.NewService(runtime.database, config.DefaultSessionPolicy(), session.WithLoginFactPort(loginFacts))
	if err != nil {
		return nil, errors.New("desktop session service failed")
	}
	runtime.sessions = sessions
	authorizer, err := demo.NewIAMAuthorizationAdapter(runtime.database)
	if err != nil {
		return nil, errors.New("desktop authorization adapter failed")
	}
	demoService, err := demo.NewService(runtime.database, authorizer)
	if err != nil {
		return nil, errors.New("desktop demo service failed")
	}
	sessionAdapter, err := demo.NewIAMSessionRequestAdapter(sessions)
	if err != nil {
		return nil, errors.New("desktop session adapter failed")
	}
	trace := func(*http.Request) string { return newTraceID() }
	sessionHandler, err := session.NewHTTPHandler(sessions, trace)
	if err != nil {
		return nil, errors.New("desktop session transport failed")
	}
	demoHandler, err := demo.NewHTTPHandler(demoService, sessionAdapter, trace)
	if err != nil {
		return nil, errors.New("desktop demo transport failed")
	}
	adminService, err := administration.NewService(runtime.database)
	if err != nil {
		return nil, errors.New("desktop administration service failed")
	}
	adminHandler, err := administration.NewHTTPHandler(adminService, sessions, trace)
	if err != nil {
		return nil, errors.New("desktop administration transport failed")
	}

	api := http.NewServeMux()
	api.Handle("/iam/session/", sessionHandler)
	api.Handle("/iam/administration/", adminHandler)
	api.Handle("/demo/", demoHandler)

	readiness, err := desktopplatform.NewNonceGate(runtime.material.ReadinessNonce)
	if err != nil {
		return nil, err
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /__desktop/ready", func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Cache-Control", "no-store")
		if request.Header.Get("Origin") != "" || !isLoopback(request.RemoteAddr) || !readiness.Consume(request.Header.Get(desktopplatform.ReadinessHeader)) {
			writer.WriteHeader(http.StatusForbidden)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"state":"ready"}`))
	})
	mux.Handle("/api/", runtime.requireControl(http.StripPrefix("/api", api)))
	mux.Handle("POST /__desktop/shutdown", runtime.requireControl(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusNoContent)
		runtime.stopOnce.Do(runtime.stop)
	})))
	runtime.registerNativeE2EControl(mux)
	return mux, nil
}

func (runtime *sidecarRuntime) requireControl(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Cache-Control", "no-store")
		if request.Header.Get("Origin") != "" || !isLoopback(request.RemoteAddr) ||
			!desktopplatform.MatchesControl(runtime.material.ControlToken, request.Header.Get(desktopplatform.ControlHeader)) {
			writer.WriteHeader(http.StatusForbidden)
			return
		}
		next.ServeHTTP(writer, request)
	})
}

func isLoopback(remote string) bool {
	host, _, err := net.SplitHostPort(remote)
	return err == nil && net.ParseIP(host) != nil && net.ParseIP(host).IsLoopback()
}

func (runtime *sidecarRuntime) drainHTTP(ctx context.Context) error {
	if runtime.server == nil {
		return nil
	}
	return runtime.server.Shutdown(ctx)
}
func (runtime *sidecarRuntime) stopHTTP(context.Context) error {
	if runtime.server == nil {
		return nil
	}
	err := runtime.server.Close()
	select {
	case serveErr := <-runtime.serveErr:
		if !errors.Is(serveErr, http.ErrServerClosed) {
			err = errors.Join(err, serveErr)
		}
	default:
	}
	return err
}
func noopLifecycle(context.Context) error { return nil }

func secureDirectories(dataDirectory, logDirectory string) error {
	for _, directory := range []string{dataDirectory, logDirectory} {
		if !filepath.IsAbs(directory) {
			return errors.New("desktop runtime directory is invalid")
		}
		if _, err := desktopplatform.EnsurePrivateDirectory(filepath.Clean(directory)); err != nil {
			return errors.New("desktop runtime directory cannot be secured")
		}
	}
	dataDirectory, logDirectory = filepath.Clean(dataDirectory), filepath.Clean(logDirectory)
	dataToLog, dataErr := filepath.Rel(dataDirectory, logDirectory)
	logToData, logErr := filepath.Rel(logDirectory, dataDirectory)
	if dataErr != nil || logErr != nil || dataToLog == "." || dataToLog != ".." && !strings.HasPrefix(dataToLog, ".."+string(filepath.Separator)) ||
		logToData != ".." && !strings.HasPrefix(logToData, ".."+string(filepath.Separator)) {
		return errors.New("desktop runtime directories conflict")
	}
	return nil
}

func newTraceID() string {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		return "0000000000000000"
	}
	return hex.EncodeToString(value)
}
