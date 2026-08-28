//go:build desktop_native_e2e

package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/app/product"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/iam/session"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/platform/config"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/platform/database"
)

func TestNativeE2EControlRequestIsExactAndBounded(t *testing.T) {
	for _, action := range []string{"scope-self", "scope-all", "permissions-off", "permissions-on", "session-revoke"} {
		request := httptest.NewRequest(http.MethodPost, "/__desktop/test-control", strings.NewReader(`{"action":"`+action+`"}`))
		request.Header.Set("Content-Type", "application/json")
		got, err := decodeNativeE2EAction(request)
		if err != nil || got != action {
			t.Fatalf("action %q = %q, %v", action, got, err)
		}
	}
	for _, body := range []string{
		`{"action":"unknown"}`,
		`{"action":"scope-all","extra":true}`,
		`{"action":"scope-all"}{"action":"scope-self"}`,
		`{"action":"` + strings.Repeat("a", 130) + `"}`,
	} {
		request := httptest.NewRequest(http.MethodPost, "/__desktop/test-control", strings.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		if _, err := decodeNativeE2EAction(request); err == nil {
			t.Fatalf("accepted invalid body %q", body)
		}
	}
	request := httptest.NewRequest(http.MethodPost, "/__desktop/test-control", strings.NewReader(`{"action":"scope-all"}`))
	if _, err := decodeNativeE2EAction(request); err == nil {
		t.Fatal("accepted request without exact JSON content type")
	}
}

func TestNativeE2EScopeControlEnforcesAndRestoresOwnership(t *testing.T) {
	ctx := context.Background()
	process := database.NewProcess()
	db, err := process.Open(ctx, database.Config{
		Profile: config.ProfileDesktopSQLite, SQLitePath: filepath.Join(t.TempDir(), "desktop.db"),
	})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	runner, err := product.NewMigrationRunner()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := runner.Up(ctx, db); err != nil {
		t.Fatal(err)
	}
	runtime := &sidecarRuntime{database: db, sessions: &session.Service{}}
	for _, action := range []string{"scope-self", "scope-self", "scope-all"} {
		if err := runtime.applyNativeE2EAction(ctx, action); err != nil {
			t.Fatalf("%s: %v", action, err)
		}
	}
	var scope string
	if err := db.Bun().NewRaw(`SELECT data_scope FROM iam_roles WHERE id = ?`, "role-system-admin").Scan(ctx, &scope); err != nil || scope != "all" {
		t.Fatalf("scope = %q, %v", scope, err)
	}
	var sentinels int
	if err := db.Bun().NewRaw(`SELECT COUNT(*) FROM demo_products WHERE sku = ?`, "E2E-FOREIGN").Scan(ctx, &sentinels); err != nil || sentinels != 1 {
		t.Fatalf("sentinels = %d, %v", sentinels, err)
	}
}
