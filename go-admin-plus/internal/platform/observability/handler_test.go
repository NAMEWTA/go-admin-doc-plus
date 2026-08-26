package observability_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"go-admin/internal/app/kernel"
	"go-admin/internal/platform/observability"
)

type statusSource struct {
	mu       sync.RWMutex
	snapshot kernel.Snapshot
}

func (source *statusSource) Snapshot() kernel.Snapshot {
	source.mu.RLock()
	defer source.mu.RUnlock()
	return source.snapshot
}

func (source *statusSource) set(snapshot kernel.Snapshot) {
	source.mu.Lock()
	source.snapshot = snapshot
	source.mu.Unlock()
}

func TestHandlerDistinguishesLifecycleStates(t *testing.T) {
	source := &statusSource{snapshot: kernel.Snapshot{State: kernel.StateStarting}}
	handler, err := observability.New(source, observability.Capabilities{
		Profile:       "server-sqlite",
		Version:       "test-version",
		Database:      "sqlite",
		Desktop:       false,
		Offline:       false,
		NativeDialogs: false,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	server := httptest.NewServer(handler.Wrap(http.NotFoundHandler()))
	defer server.Close()

	assertEndpoint(t, server.URL+observability.LivePath, http.StatusOK, `"status":"live"`)
	assertEndpoint(t, server.URL+observability.ReadyPath, http.StatusServiceUnavailable, `"status":"not_ready"`)

	source.set(kernel.Snapshot{State: kernel.StateReady})
	assertEndpoint(t, server.URL+observability.ReadyPath, http.StatusOK, `"status":"ready"`)
	metrics := assertEndpoint(t, server.URL+observability.MetricsPath, http.StatusOK, "go_admin_runtime_ready 1")
	if !strings.Contains(metrics, `go_admin_runtime_state{state="ready"} 1`) {
		t.Fatalf("metrics body = %q", metrics)
	}
	capabilities := assertEndpoint(t, server.URL+observability.CapabilitiesPath, http.StatusOK, `"profile":"server-sqlite"`)
	if strings.Contains(capabilities, "secret") || strings.Contains(capabilities, "path") {
		t.Fatalf("capabilities leaked configuration shape: %s", capabilities)
	}

	source.set(kernel.Snapshot{State: kernel.StateDraining})
	assertEndpoint(t, server.URL+observability.ReadyPath, http.StatusServiceUnavailable, `"status":"not_ready"`)
	status := assertEndpoint(t, server.URL+observability.StatusPath, http.StatusOK, `"state":"draining"`)
	var decoded map[string]any
	if err := json.Unmarshal([]byte(status), &decoded); err != nil {
		t.Fatalf("decode status: %v", err)
	}
	if decoded["ready"] != false {
		t.Fatalf("status ready = %#v", decoded["ready"])
	}

	source.set(kernel.Snapshot{State: kernel.StateFailed, Failure: kernel.FailureDependency})
	assertEndpoint(t, server.URL+observability.LivePath, http.StatusOK, `"status":"live"`)
	assertEndpoint(t, server.URL+observability.ReadyPath, http.StatusServiceUnavailable, `"status":"not_ready"`)

	source.set(kernel.Snapshot{State: kernel.StateStopped})
	assertEndpoint(t, server.URL+observability.ReadyPath, http.StatusServiceUnavailable, `"status":"not_ready"`)
}

func TestDependencyProbeFailureIsRedactedEverywhere(t *testing.T) {
	source := &statusSource{snapshot: kernel.Snapshot{State: kernel.StateReady}}
	secret := "postgres://admin:never-return-this@example.invalid/product"
	handler, err := observability.New(source, observability.Capabilities{
		Profile:  "server-postgres",
		Version:  "test-version",
		Database: "postgres",
	}, observability.Probe{
		Name: "database",
		Check: func(context.Context) error {
			return errors.New(secret)
		},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	server := httptest.NewServer(handler.Wrap(http.NotFoundHandler()))
	defer server.Close()

	checks := []struct {
		path       string
		wantStatus int
		wantBody   string
	}{
		{path: observability.LivePath, wantStatus: http.StatusOK, wantBody: `"status":"live"`},
		{path: observability.ReadyPath, wantStatus: http.StatusServiceUnavailable, wantBody: `"status":"not_ready"`},
		{path: observability.MetricsPath, wantStatus: http.StatusOK, wantBody: "go_admin_runtime_ready 0"},
		{path: observability.CapabilitiesPath, wantStatus: http.StatusOK, wantBody: `"database":"postgres"`},
		{path: observability.StatusPath, wantStatus: http.StatusOK, wantBody: `"ready":false`},
	}
	for _, check := range checks {
		body := assertEndpoint(t, server.URL+check.path, check.wantStatus, check.wantBody)
		if strings.Contains(body, secret) || strings.Contains(body, "never-return-this") {
			t.Fatalf("GET %s leaked probe failure: %s", check.path, body)
		}
	}

	request := httptest.NewRequest(http.MethodPost, observability.StatusPath, nil)
	response := httptest.NewRecorder()
	handler.Wrap(nil).ServeHTTP(response, request)
	if response.Code != http.StatusMethodNotAllowed || response.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("POST status response = %d headers=%v", response.Code, response.Header())
	}
}

func assertEndpoint(t *testing.T, endpoint string, wantStatus int, wantBody string) string {
	t.Helper()
	response, err := http.Get(endpoint)
	if err != nil {
		t.Fatalf("GET %s: %v", endpoint, err)
	}
	defer response.Body.Close()
	if response.StatusCode != wantStatus {
		t.Fatalf("GET %s status = %d, want %d", endpoint, response.StatusCode, wantStatus)
	}
	if response.Header.Get("Cache-Control") != "no-store" {
		t.Fatalf("GET %s Cache-Control = %q", endpoint, response.Header.Get("Cache-Control"))
	}
	contents, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read GET %s: %v", endpoint, err)
	}
	body := string(contents)
	if !strings.Contains(body, wantBody) {
		t.Fatalf("GET %s body = %q, want %q", endpoint, body, wantBody)
	}
	return body
}
