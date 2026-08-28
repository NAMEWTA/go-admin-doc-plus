package demo_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/demo"
)

type requestAuthenticator struct{}

func (requestAuthenticator) CookieName() string { return "__Host-test" }

func (requestAuthenticator) AuthorizeRequest(_ context.Context, token, csrf string, mutation bool) (demo.RequestIdentity, error) {
	if token != "test-session" {
		return demo.RequestIdentity{}, demo.ErrAuthentication
	}
	if mutation && csrf != strings.Repeat("c", 43) {
		return demo.RequestIdentity{}, demo.ErrCSRF
	}
	return demo.RequestIdentity{ActorID: "http-actor", CSRF: strings.Repeat("c", 43)}, nil
}

func TestHTTPContractCSRFPermissionAndStableProblems(t *testing.T) {
	db := openDatabase(t, filepath.Join(t.TempDir(), "http.sqlite3"))
	service := newService(t, db, demo.ScopeAll, nil)
	handler, err := demo.NewHTTPHandler(service, requestAuthenticator{}, func(*http.Request) string { return "0123456789abcdef" })
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, "/demo/products", bytes.NewBufferString(`{"sku":"HTTP-01","name":"HTTP product","description":"","priceCents":100,"status":"active"}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-CSRF-Token", strings.Repeat("c", 43))
	request.AddCookie(&http.Cookie{Name: "__Host-test", Value: "test-session"})
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != 201 || response.Header().Get("X-CSRF-Token") != strings.Repeat("c", 43) {
		t.Fatalf("create status=%d csrf=%v", response.Code, response.Header().Get("X-CSRF-Token") != "")
	}

	missingCSRF := httptest.NewRequest(http.MethodPost, "/demo/products", bytes.NewBufferString(`{"sku":"HTTP-02","name":"HTTP product","description":"","priceCents":100,"status":"active"}`))
	missingCSRF.Header.Set("Content-Type", "application/json")
	missingCSRF.AddCookie(&http.Cookie{Name: "__Host-test", Value: "test-session"})
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, missingCSRF)
	if response.Code != 403 || !strings.Contains(response.Body.String(), `"code":"CSRF_REJECTED"`) {
		t.Fatalf("csrf response=%d body=%s", response.Code, response.Body.String())
	}
	var count int
	if err := db.Bun().QueryRowContext(context.Background(), `SELECT COUNT(*) FROM demo_products`).Scan(&count); err != nil || count != 1 {
		t.Fatalf("csrf changed state count=%d err=%v", count, err)
	}

	denied := newService(t, db, demo.ScopeAll, map[string]bool{demo.PermissionProductsRead: true})
	deniedHandler, err := demo.NewHTTPHandler(denied, requestAuthenticator{}, func(*http.Request) string { return "0123456789abcdef" })
	if err != nil {
		t.Fatal(err)
	}
	list := httptest.NewRequest(http.MethodGet, "/demo/products", nil)
	list.AddCookie(&http.Cookie{Name: "__Host-test", Value: "test-session"})
	response = httptest.NewRecorder()
	deniedHandler.ServeHTTP(response, list)
	if response.Code != 403 || !strings.Contains(response.Body.String(), `"code":"PERMISSION_DENIED"`) {
		t.Fatalf("denied response=%d body=%s", response.Code, response.Body.String())
	}
}
