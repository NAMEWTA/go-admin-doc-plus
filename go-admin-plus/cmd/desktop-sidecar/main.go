package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	desktopplatform "go-admin/internal/platform/desktop"
)

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "desktop sidecar failed")
		os.Exit(1)
	}
}

func run() error {
	material, err := desktopplatform.ReadLaunchMaterial(os.Stdin)
	if err != nil {
		return err
	}
	parent, cancelSignals := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancelSignals()
	runCtx, stop := context.WithCancel(parent)
	runtime, err := newSidecarRuntime(material, stop)
	if err != nil {
		return err
	}
	if err := runtime.kernel.Start(runCtx); err != nil {
		return err
	}
	status := struct {
		State string `json:"state"`
		Port  uint16 `json:"port"`
	}{State: "listening", Port: runtime.port()}
	if err := json.NewEncoder(os.Stdout).Encode(status); err != nil {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		return errors.Join(err, runtime.kernel.Drain(shutdownCtx))
	}
	select {
	case <-runCtx.Done():
	case serveErr := <-runtime.serveErr:
		if serveErr != nil && !errors.Is(serveErr, context.Canceled) {
			err = errors.New("desktop sidecar server stopped")
		}
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	return errors.Join(err, runtime.kernel.Drain(shutdownCtx))
}
