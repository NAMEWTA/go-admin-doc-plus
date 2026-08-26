package health_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"go-admin/internal/application"
	"go-admin/internal/application/health"
)

func TestLiveReadyAndCapabilitiesHaveDistinctPublicSemantics(t *testing.T) {
	var state atomic.Value
	state.Store(application.StateStarting)
	dependencyErr := atomic.Bool{}

	handler, err := health.New(
		func() application.State { return state.Load().(application.State) },
		health.Capabilities{
			HostProfile:   "server",
			Version:       "test-version",
			Desktop:       false,
			Offline:       false,
			NativeDialogs: false,
		},
		health.Checker{Name: "database", Check: func(context.Context) error {
			if dependencyErr.Load() {
				return errors.New("postgres://operator:secret@example.invalid/private")
			}
			return nil
		}},
	)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	server := httptest.NewServer(handler.Wrap(http.NotFoundHandler()))
	t.Cleanup(server.Close)

	assertStatus(t, server.URL+"/health/live", http.StatusOK, `{"status":"live"}`)
	assertStatus(t, server.URL+"/health/ready", http.StatusServiceUnavailable, `{"status":"not_ready"}`)

	state.Store(application.StateReady)
	assertStatus(t, server.URL+"/health/ready", http.StatusOK, `{"status":"ready"}`)

	dependencyErr.Store(true)
	response := assertStatus(t, server.URL+"/health/ready", http.StatusServiceUnavailable, `{"status":"not_ready"}`)
	if strings.Contains(response, "secret") || strings.Contains(response, "postgres://") {
		t.Fatalf("readiness leaked dependency details: %s", response)
	}
	assertStatus(t, server.URL+"/health/live", http.StatusOK, `{"status":"live"}`)

	response = assertStatus(t, server.URL+"/api/v1/runtime/capabilities", http.StatusOK, "")
	var capabilities health.Capabilities
	if err := json.Unmarshal([]byte(response), &capabilities); err != nil {
		t.Fatalf("decode capabilities: %v", err)
	}
	if capabilities.HostProfile != "server" || capabilities.Version != "test-version" || capabilities.Desktop || capabilities.Offline || capabilities.NativeDialogs {
		t.Fatalf("capabilities = %#v", capabilities)
	}
}

func TestHealthRejectsInvalidConfigurationAndMethods(t *testing.T) {
	if _, err := health.New(nil, health.Capabilities{}); err == nil {
		t.Fatal("New accepted a nil application state function")
	}
	if _, err := health.New(
		func() application.State { return application.StateReady },
		health.Capabilities{HostProfile: "server"},
		health.Checker{Name: "", Check: func(context.Context) error { return nil }},
	); err == nil {
		t.Fatal("New accepted an unnamed readiness checker")
	}

	handler, err := health.New(
		func() application.State { return application.StateReady },
		health.Capabilities{HostProfile: "server"},
	)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	request := httptest.NewRequest(http.MethodPost, "/health/live", nil)
	response := httptest.NewRecorder()
	handler.Wrap(http.NotFoundHandler()).ServeHTTP(response, request)
	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST /health/live status = %d, want 405", response.Code)
	}
}

func assertStatus(t *testing.T, url string, wantStatus int, wantBody string) string {
	t.Helper()
	response, err := http.Get(url)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	defer response.Body.Close()
	var decoded any
	if err := json.NewDecoder(response.Body).Decode(&decoded); err != nil {
		t.Fatalf("decode %s: %v", url, err)
	}
	body, err := json.Marshal(decoded)
	if err != nil {
		t.Fatalf("encode response: %v", err)
	}
	if response.StatusCode != wantStatus {
		t.Fatalf("GET %s status = %d, want %d; body = %s", url, response.StatusCode, wantStatus, body)
	}
	if wantBody != "" && string(body) != wantBody {
		t.Fatalf("GET %s body = %s, want %s", url, body, wantBody)
	}
	if response.Header.Get("Cache-Control") != "no-store" {
		t.Fatalf("GET %s Cache-Control = %q, want no-store", url, response.Header.Get("Cache-Control"))
	}
	return string(body)
}
