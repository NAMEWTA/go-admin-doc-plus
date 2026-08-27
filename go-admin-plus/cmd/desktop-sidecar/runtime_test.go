package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	desktopplatform "go-admin/internal/platform/desktop"
)

func TestDatabaseStartupFailureAlwaysClosesBeforeRestoreAndReportsRecoveryFault(t *testing.T) {
	var order []string
	err := databaseStartupFailure(func() error {
		order = append(order, "close")
		return nil
	}, func() error {
		order = append(order, "restore")
		return nil
	}, "desktop database migration failed")
	if err.Error() != "desktop database migration failed" || fmt.Sprint(order) != "[close restore]" {
		t.Fatalf("successful recovery = %v order=%v", err, order)
	}
	order = nil
	err = databaseStartupFailure(nil, func() error {
		order = append(order, "restore")
		return errors.New("injected restore fault")
	}, "desktop database open failed")
	if err.Error() != "desktop database recovery failed" || fmt.Sprint(order) != "[restore]" {
		t.Fatalf("failed recovery = %v order=%v", err, order)
	}
	order = nil
	err = databaseStartupFailure(func() error {
		order = append(order, "close")
		return errors.New("injected close fault")
	}, func() error {
		order = append(order, "restore")
		return nil
	}, "desktop database migration failed")
	if err.Error() != "desktop database recovery failed" || fmt.Sprint(order) != "[close]" {
		t.Fatalf("close fault recovery = %v order=%v", err, order)
	}
}

func TestControlBoundaryRejectsOriginAndMissingSecret(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	runtime, err := newSidecarRuntime(desktopplatform.LaunchMaterial{
		DataDirectory: root + "/data", LogDirectory: root + "/logs", LoopbackPort: 0,
		ReadinessNonce: "abcdefghijklmnopqrstuvwxyzABCDEFGH123456789",
		ControlToken:   "ZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZ",
	}, cancel)
	if err != nil {
		t.Fatal(err)
	}
	if err := runtime.kernel.Start(ctx); err != nil {
		t.Fatal(err)
	}
	port := runtime.port()
	if port == 0 {
		t.Fatal("sidecar did not bind an operating-system-assigned port")
	}
	defer func() {
		shutdown, stop := context.WithTimeout(context.Background(), 3*time.Second)
		defer stop()
		if err := runtime.kernel.Drain(shutdown); err != nil {
			t.Fatal(err)
		}
	}()
	client := &http.Client{Timeout: time.Second}
	request, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1:"+desktopplatformPort(port)+"/api/iam/session/current", nil)
	response, err := client.Do(request)
	if err != nil || response.StatusCode != http.StatusForbidden {
		t.Fatalf("missing control status=%v err=%v", responseStatus(response), err)
	}
	_ = response.Body.Close()
	request, _ = http.NewRequest(http.MethodGet, "http://127.0.0.1:"+desktopplatformPort(port)+"/api/iam/session/current", nil)
	request.Header.Set(desktopplatform.ControlHeader, "ZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZ")
	request.Header.Set("Origin", "tauri://localhost")
	response, err = client.Do(request)
	if err != nil || response.StatusCode != http.StatusForbidden {
		t.Fatalf("origin status=%v err=%v", responseStatus(response), err)
	}
	_ = response.Body.Close()
	ready, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1:"+desktopplatformPort(port)+"/__desktop/ready", nil)
	ready.Header.Set(desktopplatform.ReadinessHeader, "abcdefghijklmnopqrstuvwxyzABCDEFGH123456789")
	response, err = client.Do(ready)
	if err != nil || response.StatusCode != http.StatusOK {
		t.Fatalf("readiness status=%v err=%v", responseStatus(response), err)
	}
	_ = response.Body.Close()
	ready, _ = http.NewRequest(http.MethodGet, "http://127.0.0.1:"+desktopplatformPort(port)+"/__desktop/ready", nil)
	ready.Header.Set(desktopplatform.ReadinessHeader, "abcdefghijklmnopqrstuvwxyzABCDEFGH123456789")
	response, err = client.Do(ready)
	if err != nil || response.StatusCode != http.StatusForbidden {
		t.Fatalf("replayed readiness status=%v err=%v", responseStatus(response), err)
	}
	_ = response.Body.Close()
}

func TestSidecarRuntimeHoldsInstanceLockUntilDrain(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	material := desktopplatform.LaunchMaterial{
		DataDirectory: root + "/data", LogDirectory: root + "/logs-one", LoopbackPort: 0,
		ReadinessNonce: "abcdefghijklmnopqrstuvwxyzABCDEFGH123456789",
		ControlToken:   "ZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZ",
	}
	first, err := newSidecarRuntime(material, cancel)
	if err != nil || first.kernel.Start(ctx) != nil {
		t.Fatalf("start first runtime: %v", err)
	}
	secondMaterial := material
	secondMaterial.LogDirectory = root + "/logs-two"
	secondMaterial.ReadinessNonce = "YYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYY"
	second, err := newSidecarRuntime(secondMaterial, cancel)
	if err != nil {
		t.Fatal(err)
	}
	if err := second.kernel.Start(ctx); err == nil {
		t.Fatal("second runtime acquired the same data-root lock")
	}
	shutdown, stop := context.WithTimeout(context.Background(), 3*time.Second)
	defer stop()
	if err := first.kernel.Drain(shutdown); err != nil {
		t.Fatal(err)
	}
	restarted, err := newSidecarRuntime(secondMaterial, cancel)
	if err != nil || restarted.kernel.Start(ctx) != nil {
		t.Fatalf("instance lock was not released after drain: %v", err)
	}
	if err := restarted.kernel.Drain(shutdown); err != nil {
		t.Fatal(err)
	}
}

func desktopplatformPort(port uint16) string { return fmt.Sprintf("%d", port) }
func responseStatus(response *http.Response) int {
	if response == nil {
		return 0
	}
	return response.StatusCode
}
