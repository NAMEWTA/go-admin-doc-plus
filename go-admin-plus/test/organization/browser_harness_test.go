package organization_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"go-admin/internal/modules/iam/account"
	"go-admin/internal/modules/iam/administration"
	"go-admin/internal/modules/iam/authorization"
	sessionmigration "go-admin/internal/modules/iam/migrations/0010-session-schema"
	administrationmigration "go-admin/internal/modules/iam/migrations/0020-administration-schema"
	"go-admin/internal/modules/iam/session"
	"go-admin/internal/modules/organization"
	organizationmigration "go-admin/internal/modules/organization/migrations"
	"go-admin/internal/platform/config"
	"go-admin/internal/platform/database"
	"go-admin/internal/platform/migrations"
)

const (
	organizationHarnessServe   = "GO_ADMIN_ORGANIZATION_E2E_SERVE"
	organizationHarnessProfile = "GO_ADMIN_ORGANIZATION_E2E_PROFILE"
	organizationHarnessReady   = "GO_ADMIN_ORGANIZATION_E2E_READY_FILE"
	organizationHarnessStatic  = "GO_ADMIN_ORGANIZATION_E2E_STATIC_DIR"
	postgresDSNEnvironment     = "GO_ADMIN_TEST_POSTGRES_DISPOSABLE_DSN"
	organizationAdminID        = "account-organization-admin-001"
)

// TestOrganizationBrowserHarnessServer is compiled by source gates and executed only by the
// required-E2E matrix. It mounts production session and organization handlers on one HTTPS origin.
func TestOrganizationBrowserHarnessServer(t *testing.T) {
	if os.Getenv(organizationHarnessServe) != "1" {
		t.Skip(organizationHarnessServe + " is not enabled")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	db := openOrganizationBrowserDB(t, ctx, os.Getenv(organizationHarnessProfile))
	runner, err := migrations.NewRunner(sessionmigration.Provider{}, administrationmigration.Provider{}, organizationmigration.Provider{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := runner.Up(ctx, db); err != nil {
		t.Fatal("organization browser migrations failed")
	}
	seedOrganizationBrowserIdentity(t, ctx, db)
	policy, err := config.NewSessionPolicy(2*time.Minute, 8*time.Minute, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	sessions, err := session.NewService(db, policy)
	if err != nil {
		t.Fatal(err)
	}
	service, err := organization.NewService(db, authorization.NewService(db))
	if err != nil {
		t.Fatal(err)
	}
	sessionHandler, err := session.NewHTTPHandler(sessions, func(*http.Request) string { return "0123456789abcdef" })
	if err != nil {
		t.Fatal(err)
	}
	organizationHandler, err := organization.NewHTTPHandler(service, sessions, func(*http.Request) string { return "0123456789abcdef" })
	if err != nil {
		t.Fatal(err)
	}
	administrationService, err := administration.NewService(db)
	if err != nil {
		t.Fatal(err)
	}
	administrationHandler, err := administration.NewHTTPHandler(administrationService, sessions, func(*http.Request) string { return "0123456789abcdef" })
	if err != nil {
		t.Fatal(err)
	}

	var cookieAttributes struct {
		sync.Mutex
		secure, httpOnly, strict bool
	}
	recordCookie := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			recorder := &headerCapture{ResponseWriter: w}
			next.ServeHTTP(recorder, r)
			value := recorder.Header().Get("Set-Cookie")
			if value != "" {
				cookieAttributes.Lock()
				cookieAttributes.secure = strings.Contains(value, "Secure")
				cookieAttributes.httpOnly = strings.Contains(value, "HttpOnly")
				cookieAttributes.strict = strings.Contains(value, "SameSite=Strict")
				cookieAttributes.Unlock()
			}
		})
	}
	shutdown := make(chan struct{})
	var once sync.Once
	mux := http.NewServeMux()
	mux.Handle("/api/iam/session/", http.StripPrefix("/api", recordCookie(sessionHandler)))
	mux.Handle("/api/iam/administration/", http.StripPrefix("/api", recordCookie(administrationHandler)))
	mux.Handle("/api/organization/", http.StripPrefix("/api", recordCookie(organizationHandler)))
	mux.HandleFunc("/__test/snapshot", func(w http.ResponseWriter, r *http.Request) {
		var departments, positions int
		if err := db.Bun().QueryRowContext(r.Context(), `SELECT COUNT(*) FROM organization_departments`).Scan(&departments); err != nil {
			w.WriteHeader(500)
			return
		}
		if err := db.Bun().QueryRowContext(r.Context(), `SELECT COUNT(*) FROM organization_positions`).Scan(&positions); err != nil {
			w.WriteHeader(500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]int{"departments": departments, "positions": positions})
	})
	mux.HandleFunc("/__test/cookie-attributes", func(w http.ResponseWriter, _ *http.Request) {
		cookieAttributes.Lock()
		defer cookieAttributes.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]bool{"secure": cookieAttributes.secure, "httpOnly": cookieAttributes.httpOnly, "strict": cookieAttributes.strict})
	})
	mux.HandleFunc("/__test/revoke-position-read", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if _, err := db.Bun().ExecContext(r.Context(), `DELETE FROM iam_role_permissions WHERE role_id = ? AND permission_code = ?`, "role-organization-admin", organization.PermissionPositionsRead); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/__test/shutdown", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		once.Do(func() { close(shutdown) })
	})
	staticRoot, readyFile := os.Getenv(organizationHarnessStatic), os.Getenv(organizationHarnessReady)
	if info, err := os.Stat(staticRoot); err != nil || !info.IsDir() || readyFile == "" {
		t.Fatal("organization browser harness paths are invalid")
	}
	mux.Handle("/", http.FileServer(http.Dir(staticRoot)))
	server := httptest.NewUnstartedServer(mux)
	server.StartTLS()
	defer server.Close()
	if err := os.MkdirAll(filepath.Dir(readyFile), 0o700); err != nil {
		t.Fatal("ready directory failed")
	}
	if err := os.WriteFile(readyFile, []byte(server.URL), 0o600); err != nil {
		t.Fatal("ready file failed")
	}
	select {
	case <-shutdown:
	case <-ctx.Done():
		t.Fatal("organization browser harness timed out")
	}
}

type headerCapture struct{ http.ResponseWriter }

func openOrganizationBrowserDB(t *testing.T, ctx context.Context, profile string) *database.Database {
	t.Helper()
	if profile == "sqlite" {
		if os.Getenv(postgresDSNEnvironment) != "" {
			t.Fatal("SQLite harness received PostgreSQL material")
		}
		db, err := database.NewProcess().Open(ctx, database.Config{Profile: config.ProfileServerSQLite, SQLitePath: filepath.Join(t.TempDir(), "organization.sqlite3")})
		if err != nil {
			t.Fatal("SQLite harness open failed")
		}
		t.Cleanup(func() { _ = db.Close() })
		return db
	}
	if profile != "postgres" {
		t.Fatal("organization browser harness profile is invalid")
	}
	dsn := os.Getenv(postgresDSNEnvironment)
	if dsn == "" {
		t.Fatal("PostgreSQL harness material is required")
	}
	admin, err := database.NewProcess().Open(ctx, database.Config{Profile: config.ProfileServerPostgres, PostgresDSN: dsn})
	if err != nil {
		t.Fatal("PostgreSQL harness administrator failed")
	}
	schema := "t09_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err := admin.SQL().ExecContext(ctx, `CREATE SCHEMA `+schema); err != nil {
		_ = admin.Close()
		t.Fatal("PostgreSQL harness schema failed")
	}
	var db *database.Database
	t.Cleanup(func() {
		if db != nil {
			_ = db.Close()
		}
		cleanup, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		if _, err := admin.SQL().ExecContext(cleanup, `DROP SCHEMA IF EXISTS `+schema+` CASCADE`); err != nil {
			t.Error("PostgreSQL harness cleanup failed")
		}
		_ = admin.Close()
	})
	db, err = database.NewProcess().Open(ctx, database.Config{Profile: config.ProfileServerPostgres, PostgresDSN: organizationPostgresDSN(t, dsn, schema)})
	if err != nil {
		t.Fatal("isolated PostgreSQL harness failed")
	}
	return db
}

func organizationPostgresDSN(t *testing.T, dsn, schema string) string {
	t.Helper()
	if !strings.HasPrefix(schema, "t09_") || strings.Trim(schema, "abcdefghijklmnopqrstuvwxyz0123456789_") != "" {
		t.Fatal("isolated schema name is invalid")
	}
	parsed, err := url.Parse(dsn)
	if err != nil {
		t.Fatal("PostgreSQL DSN is invalid")
	}
	query := parsed.Query()
	query.Set("search_path", schema)
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

func seedOrganizationBrowserIdentity(t *testing.T, ctx context.Context, db *database.Database) {
	t.Helper()
	hash, err := account.HashPassword("organization administrator password")
	if err != nil {
		t.Fatal(err)
	}
	if err := db.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error {
		return account.NewRepository(db.Dialect()).Create(ctx, tx, account.Credential{Profile: account.Profile{ID: organizationAdminID, Username: "organization-admin", DisplayName: "Organization Administrator", Email: "organization-admin@example.test"}, PasswordHash: hash}, time.Now().UTC())
	}); err != nil {
		t.Fatal("organization account seed failed")
	}
	now := time.Now().UTC()
	if _, err := db.Bun().ExecContext(ctx, `INSERT INTO iam_roles(id, role_key, name, data_scope, enabled, protected, created_at, updated_at) VALUES (?, ?, ?, 'all', TRUE, FALSE, ?, ?)`, "role-organization-admin", "organization-admin", "Organization Administrator", now, now); err != nil {
		t.Fatal("organization role seed failed")
	}
	permissions := []string{organization.PermissionDepartmentsRead, organization.PermissionDepartmentsWrite, organization.PermissionDepartmentsDelete, organization.PermissionPositionsRead, organization.PermissionPositionsWrite, organization.PermissionPositionsDelete}
	for _, permission := range permissions {
		if _, err := db.Bun().ExecContext(ctx, `INSERT INTO iam_permissions(code, name) VALUES (?, ?)`, permission, permission); err != nil {
			t.Fatal("organization permission seed failed")
		}
		if _, err := db.Bun().ExecContext(ctx, `INSERT INTO iam_role_permissions(role_id, permission_code) VALUES (?, ?)`, "role-organization-admin", permission); err != nil {
			t.Fatal("organization role permission seed failed")
		}
	}
	if _, err := db.Bun().ExecContext(ctx, `INSERT INTO iam_role_permissions(role_id, permission_code) VALUES (?, ?)`, "role-organization-admin", authorization.PermissionManifestRead); err != nil {
		t.Fatal("organization manifest permission seed failed")
	}
	for _, menu := range []struct{ id, key, label, path, permission string }{
		{"menu-organization-departments", "organization-departments", "Departments", "/organization/departments", organization.PermissionDepartmentsRead},
		{"menu-organization-positions", "organization-positions", "Positions", "/organization/positions", organization.PermissionPositionsRead},
	} {
		if _, err := db.Bun().ExecContext(ctx, `INSERT INTO iam_menus(id, menu_key, label, path, permission_code, sort_order, protected, created_at, updated_at) VALUES (?, ?, ?, ?, ?, 100, FALSE, ?, ?)`, menu.id, menu.key, menu.label, menu.path, menu.permission, now, now); err != nil {
			t.Fatal("organization menu seed failed")
		}
		if _, err := db.Bun().ExecContext(ctx, `INSERT INTO iam_role_menus(role_id, menu_id) VALUES (?, ?)`, "role-organization-admin", menu.id); err != nil {
			t.Fatal("organization role menu seed failed")
		}
	}
	if _, err := db.Bun().ExecContext(ctx, `INSERT INTO iam_account_roles(account_id, role_id) VALUES (?, ?)`, organizationAdminID, "role-organization-admin"); err != nil {
		t.Fatal("organization account role seed failed")
	}
}

func TestOrganizationPostgresDSNUsesIsolatedSchemaWithoutExposingMaterial(t *testing.T) {
	const secret = "postgres://user:private-password@private.example/database?sslmode=disable"
	value := organizationPostgresDSN(t, secret, "t09_contract")
	parsed, err := url.Parse(value)
	if err != nil || parsed.Query().Get("search_path") != "t09_contract" || parsed.Query().Get("sslmode") != "disable" {
		t.Fatal("organization PostgreSQL DSN lost isolation settings")
	}
	diagnostic := fmt.Sprintf("%v", database.Config{Profile: config.ProfileServerPostgres, PostgresDSN: secret})
	if strings.Contains(diagnostic, "private-password") || strings.Contains(diagnostic, "private.example") {
		t.Fatal("database diagnostics exposed PostgreSQL material")
	}
}
