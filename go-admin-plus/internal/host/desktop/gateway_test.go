package desktop

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGatewayRequiresLaunchTokenAndExactDesktopOrigin(t *testing.T) {
	gateway, err := NewGateway("launch-secret")
	if err != nil {
		t.Fatalf("NewGateway: %v", err)
	}
	handler := gateway.Wrap(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusNoContent)
	}))

	tests := []struct {
		name   string
		origin string
		token  string
		status int
	}{
		{name: "mac", origin: "wails://wails", token: "launch-secret", status: http.StatusNoContent},
		{name: "windows", origin: "http://wails.localhost", token: "launch-secret", status: http.StatusNoContent},
		{name: "missing token", origin: "wails://wails", status: http.StatusUnauthorized},
		{name: "wrong token", origin: "wails://wails", token: "wrong", status: http.StatusUnauthorized},
		{name: "web origin", origin: "https://example.test", token: "launch-secret", status: http.StatusForbidden},
		{name: "missing origin", token: "launch-secret", status: http.StatusForbidden},
		{name: "non-loopback client", origin: "wails://wails", token: "launch-secret", status: http.StatusForbidden},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "http://127.0.0.1/api/v1/test", nil)
			request.RemoteAddr = "127.0.0.1:54321"
			if test.name == "non-loopback client" {
				request.RemoteAddr = "192.0.2.10:54321"
			}
			request.Header.Set("Origin", test.origin)
			request.Header.Set(LaunchTokenHeader, test.token)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != test.status {
				t.Fatalf("status = %d, want %d; body %s", response.Code, test.status, response.Body.String())
			}
			if test.status == http.StatusNoContent && response.Header().Get("Access-Control-Allow-Origin") != test.origin {
				t.Fatalf("allow origin = %q", response.Header().Get("Access-Control-Allow-Origin"))
			}
		})
	}
}

func TestGatewayAnswersValidPreflightWithoutExposingWildcardCORS(t *testing.T) {
	gateway, err := NewGateway("launch-secret")
	if err != nil {
		t.Fatalf("NewGateway: %v", err)
	}
	request := httptest.NewRequest(http.MethodOptions, "http://127.0.0.1/api/v1/login", nil)
	request.RemoteAddr = "[::1]:54321"
	request.Header.Set("Origin", "wails://wails")
	request.Header.Set("Access-Control-Request-Method", http.MethodPost)
	request.Header.Set("Access-Control-Request-Headers", "content-type, x-go-admin-launch-token")
	response := httptest.NewRecorder()

	gateway.Wrap(http.NotFoundHandler()).ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", response.Code)
	}
	if got := response.Header().Get("Access-Control-Allow-Origin"); got != "wails://wails" || got == "*" {
		t.Fatalf("allow origin = %q", got)
	}
	if got := response.Header().Get("Access-Control-Allow-Headers"); got != "Authorization, Content-Type, X-Go-Admin-Launch-Token" {
		t.Fatalf("allow headers = %q", got)
	}
}

func TestGatewayRejectsEmptyLaunchToken(t *testing.T) {
	if _, err := NewGateway(""); err == nil {
		t.Fatal("NewGateway unexpectedly accepted an empty token")
	}
}

func TestGatewayRejectsMalformedRemoteAddressWithoutCallingApplication(t *testing.T) {
	gateway, err := NewGateway("launch-secret")
	if err != nil {
		t.Fatalf("NewGateway: %v", err)
	}
	called := false
	request := httptest.NewRequest(http.MethodGet, "http://127.0.0.1/api/v1/test", nil)
	request.RemoteAddr = "not-an-address"
	request.Header.Set("Origin", "wails://wails")
	request.Header.Set(LaunchTokenHeader, "launch-secret")
	response := httptest.NewRecorder()
	gateway.Wrap(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		called = true
	})).ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", response.Code)
	}
	if called {
		t.Fatal("application handler was called for malformed remote address")
	}
}
