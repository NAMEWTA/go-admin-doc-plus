package demo_test

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

	"go-admin/internal/modules/demo"
	productsmigration "go-admin/internal/modules/demo/migrations/0010-products"
	"go-admin/internal/modules/iam/account"
	"go-admin/internal/modules/iam/administration"
	"go-admin/internal/modules/iam/authorization"
	sessionmigration "go-admin/internal/modules/iam/migrations/0010-session-schema"
	administrationmigration "go-admin/internal/modules/iam/migrations/0020-administration-schema"
	"go-admin/internal/modules/iam/session"
	"go-admin/internal/platform/config"
	"go-admin/internal/platform/database"
	"go-admin/internal/platform/migrations"
)

const (
	demoServe           = "GO_ADMIN_DEMO_E2E_SERVE"
	demoProfile         = "GO_ADMIN_DEMO_E2E_PROFILE"
	demoReady           = "GO_ADMIN_DEMO_E2E_READY_FILE"
	demoStatic          = "GO_ADMIN_DEMO_E2E_STATIC_DIR"
	postgresEnvironment = "GO_ADMIN_TEST_POSTGRES_DISPOSABLE_DSN"
	demoAdminID         = "account-demo-admin"
	demoForeignID       = "account-demo-foreign"
)

type demoBrowserLoginFactNoop struct{}

func (demoBrowserLoginFactNoop) RecordLoginFact(context.Context, database.Tx, session.LoginFact) error {
	return nil
}

// TestDemoBrowserHarnessServer is compiled by source gates and run only by the Lead matrix.
func TestDemoBrowserHarnessServer(t *testing.T) {
	if os.Getenv(demoServe) != "1" {
		t.Skip(demoServe + " is not enabled")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	profile := os.Getenv(demoProfile)
	opener, cleanup, expectedSchema := demoDatabaseOpener(t, ctx, profile)
	defer cleanup()
	current := &switchingHandler{}
	type liveRuntime struct {
		db       *database.Database
		sessions *session.Service
	}
	var liveMu sync.RWMutex
	var live liveRuntime
	open := func() *database.Database {
		db := opener()
		if profile == "postgres" {
			var schema string
			if err := db.Bun().QueryRowContext(ctx, `SELECT current_schema()`).Scan(&schema); err != nil || schema != expectedSchema {
				t.Fatalf("demo harness PostgreSQL schema is not isolated current=%q err=%v", schema, err)
			}
		}
		runner, err := migrations.NewRunner(sessionmigration.Provider{}, administrationmigration.Provider{}, productsmigration.Provider{})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := runner.Up(ctx, db); err != nil {
			t.Fatal("demo harness migration failed")
		}
		seedDemoBrowserAccounts(t, ctx, db)
		registry, err := authorization.NewCapabilityRegistry(db)
		if err != nil {
			t.Fatal(err)
		}
		if err := demo.RegisterCapabilities(ctx, registry); err != nil {
			t.Fatal("demo capability registration failed")
		}
		seedForeignDemoProduct(t, ctx, db)
		policy, err := config.NewSessionPolicy(2*time.Minute, 8*time.Minute, time.Minute)
		if err != nil {
			t.Fatal(err)
		}
		sessions, err := session.NewService(db, policy, session.WithLoginFactPort(demoBrowserLoginFactNoop{}))
		if err != nil {
			t.Fatal(err)
		}
		authorizer, err := demo.NewIAMAuthorizationAdapter(db)
		if err != nil {
			t.Fatal(err)
		}
		service, err := demo.NewService(db, authorizer)
		if err != nil {
			t.Fatal(err)
		}
		authenticator, err := demo.NewIAMSessionRequestAdapter(sessions)
		if err != nil {
			t.Fatal(err)
		}
		demoHandler, err := demo.NewHTTPHandler(service, authenticator, func(*http.Request) string { return "0123456789abcdef" })
		if err != nil {
			t.Fatal(err)
		}
		sessionHandler, err := session.NewHTTPHandler(sessions, func(*http.Request) string { return "0123456789abcdef" })
		if err != nil {
			t.Fatal(err)
		}
		adminService, err := administration.NewService(db)
		if err != nil {
			t.Fatal(err)
		}
		adminHandler, err := administration.NewHTTPHandler(adminService, sessions, func(*http.Request) string { return "0123456789abcdef" })
		if err != nil {
			t.Fatal(err)
		}
		api := http.NewServeMux()
		api.Handle("/demo/", demoHandler)
		api.Handle("/iam/session/", sessionHandler)
		api.Handle("/iam/administration/", adminHandler)
		current.set(http.StripPrefix("/api", api))
		liveMu.Lock()
		live = liveRuntime{db: db, sessions: sessions}
		liveMu.Unlock()
		return db
	}
	db := open()
	defer func() { _ = db.Close() }()
	mux := http.NewServeMux()
	mux.Handle("/api/", current)
	mux.HandleFunc("/__test/restart", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(405)
			return
		}
		if err := db.Close(); err != nil {
			w.WriteHeader(500)
			return
		}
		db = open()
		w.WriteHeader(204)
	})
	mux.HandleFunc("/__test/permissions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(405)
			return
		}
		var body struct {
			Enabled bool `json:"enabled"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024)).Decode(&body) != nil {
			w.WriteHeader(400)
			return
		}
		liveMu.RLock()
		runtime := live
		liveMu.RUnlock()
		if body.Enabled {
			registry, err := authorization.NewCapabilityRegistry(runtime.db)
			if err != nil || demo.RegisterCapabilities(r.Context(), registry) != nil {
				w.WriteHeader(500)
				return
			}
		} else if _, err := runtime.db.Bun().ExecContext(r.Context(), `DELETE FROM iam_role_permissions WHERE role_id = ? AND permission_code IN (?, ?)`, "role-system-admin", demo.PermissionProductsWrite, demo.PermissionProductsDelete); err != nil {
			w.WriteHeader(500)
			return
		}
		w.WriteHeader(204)
	})
	mux.HandleFunc("/__test/scope", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(405)
			return
		}
		var body struct {
			Scope string `json:"scope"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024)).Decode(&body) != nil || (body.Scope != "self" && body.Scope != "all") {
			w.WriteHeader(400)
			return
		}
		liveMu.RLock()
		runtime := live
		liveMu.RUnlock()
		if _, err := runtime.db.Bun().ExecContext(r.Context(), `UPDATE iam_roles SET data_scope = ? WHERE id = ?`, body.Scope, "role-system-admin"); err != nil {
			w.WriteHeader(500)
			return
		}
		w.WriteHeader(204)
	})
	mux.HandleFunc("/__test/revoke-session", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(405)
			return
		}
		liveMu.RLock()
		runtime := live
		liveMu.RUnlock()
		if err := runtime.sessions.RevokeAccount(r.Context(), demoAdminID); err != nil {
			w.WriteHeader(500)
			return
		}
		w.WriteHeader(204)
	})
	shutdown := make(chan struct{})
	var once sync.Once
	mux.HandleFunc("/__test/shutdown", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(405)
			return
		}
		w.WriteHeader(204)
		once.Do(func() { close(shutdown) })
	})
	staticRoot, ready := os.Getenv(demoStatic), os.Getenv(demoReady)
	if info, err := os.Stat(staticRoot); err != nil || !info.IsDir() || ready == "" {
		t.Fatal("demo harness paths are invalid")
	}
	files := http.FileServer(http.Dir(staticRoot))
	mux.Handle("/", files)
	server := httptest.NewUnstartedServer(mux)
	server.StartTLS()
	defer server.Close()
	if err := os.MkdirAll(filepath.Dir(ready), 0o700); err != nil {
		t.Fatal("ready directory failed")
	}
	if err := os.WriteFile(ready, []byte(server.URL), 0o600); err != nil {
		t.Fatal("ready file failed")
	}
	select {
	case <-shutdown:
	case <-ctx.Done():
		t.Fatal("demo browser harness timed out")
	}
}

type switchingHandler struct {
	mu      sync.RWMutex
	handler http.Handler
}

func (value *switchingHandler) set(handler http.Handler) {
	value.mu.Lock()
	defer value.mu.Unlock()
	value.handler = handler
}
func (value *switchingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	value.mu.RLock()
	handler := value.handler
	value.mu.RUnlock()
	handler.ServeHTTP(w, r)
}

func seedDemoBrowserAccounts(t *testing.T, ctx context.Context, db *database.Database) {
	t.Helper()
	repository := account.NewRepository(db.Dialect())
	for _, value := range []struct{ id, username, password string }{{demoAdminID, "admin", "administrator password"}, {demoForeignID, "foreign", "foreign password value"}} {
		var count int
		if err := db.Bun().QueryRowContext(ctx, `SELECT COUNT(*) FROM iam_accounts WHERE id = ?`, value.id).Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count == 0 {
			hash, err := account.HashPassword(value.password)
			if err != nil {
				t.Fatal(err)
			}
			err = db.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error {
				return repository.Create(ctx, tx, account.Credential{Profile: account.Profile{ID: value.id, Username: value.username, DisplayName: strings.ToUpper(value.username[:1]) + value.username[1:], Email: value.username + "@example.test"}, PasswordHash: hash}, time.Now().UTC())
			})
			if err != nil {
				t.Fatal("demo browser account seed failed")
			}
		}
	}
	if _, err := db.Bun().ExecContext(ctx, `INSERT INTO iam_account_roles(account_id, role_id) VALUES (?, ?) ON CONFLICT(account_id, role_id) DO NOTHING`, demoAdminID, "role-system-admin"); err != nil {
		t.Fatal("demo browser role seed failed")
	}
}

func seedForeignDemoProduct(t *testing.T, ctx context.Context, db *database.Database) {
	t.Helper()
	now := time.Now().UTC()
	name := "Foreign product"
	nameKey := demoFixtureNameKey(name)
	if _, err := db.Bun().ExecContext(ctx, `INSERT INTO demo_products(id, owner_account_id, sku, name, name_key, description, price_cents, status, revision, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON CONFLICT(id) DO NOTHING`, "00000000-0000-4000-8000-000000000099", demoForeignID, "FOREIGN-01", name, nameKey, "scope probe", 1, "active", 1, now, now); err != nil {
		t.Fatal("foreign demo seed failed")
	}
}

func demoFixtureNameKey(value string) string {
	var result strings.Builder
	for _, value := range []byte(strings.ToLower(value)) {
		_, _ = fmt.Fprintf(&result, "%02x.", value)
	}
	return result.String()
}

func demoDatabaseOpener(t *testing.T, ctx context.Context, profile string) (func() *database.Database, func(), string) {
	t.Helper()
	if profile == "sqlite" {
		if os.Getenv(postgresEnvironment) != "" {
			t.Fatal("SQLite demo harness received PostgreSQL material")
		}
		path := filepath.Join(t.TempDir(), "demo.sqlite3")
		return func() *database.Database {
			db, err := database.NewProcess().Open(ctx, database.Config{Profile: config.ProfileServerSQLite, SQLitePath: path})
			if err != nil {
				t.Fatal("SQLite demo harness open failed")
			}
			return db
		}, func() {}, ""
	}
	if profile != "postgres" {
		t.Fatal("demo harness profile is invalid")
	}
	dsn := os.Getenv(postgresEnvironment)
	if dsn == "" {
		t.Fatal("PostgreSQL demo harness material is required")
	}
	admin, err := database.NewProcess().Open(ctx, database.Config{Profile: config.ProfileServerPostgres, PostgresDSN: dsn})
	if err != nil {
		t.Fatal("PostgreSQL demo harness administrator failed")
	}
	schema := fmt.Sprintf("t14_demo_%d", time.Now().UnixNano())
	if _, err := admin.SQL().ExecContext(ctx, `CREATE SCHEMA `+schema); err != nil {
		_ = admin.Close()
		t.Fatal("PostgreSQL demo harness schema failed")
	}
	isolated := demoSearchPath(t, dsn, schema)
	open := func() *database.Database {
		db, err := database.NewProcess().Open(ctx, database.Config{Profile: config.ProfileServerPostgres, PostgresDSN: isolated})
		if err != nil {
			t.Fatal("isolated PostgreSQL demo harness failed")
		}
		return db
	}
	cleanup := func() {
		cleanupContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if _, err := admin.SQL().ExecContext(cleanupContext, `DROP SCHEMA IF EXISTS `+schema+` CASCADE`); err != nil {
			t.Error("PostgreSQL demo cleanup failed")
		}
		_ = admin.Close()
	}
	return open, cleanup, schema
}

func demoSearchPath(t *testing.T, dsn, schema string) string {
	t.Helper()
	if !strings.HasPrefix(schema, "t14_") {
		t.Fatal("isolated demo schema name is invalid")
	}
	parsed, err := url.Parse(dsn)
	if err != nil {
		t.Fatal("PostgreSQL demo material is invalid")
	}
	query := parsed.Query()
	query.Set("search_path", schema)
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

func TestDemoSearchPathIsURLSafe(t *testing.T) {
	value := demoSearchPath(t, "postgres://localhost/example?sslmode=disable", "t14_demo_contract")
	if !strings.Contains(value, "search_path=t14_demo_contract") || !strings.Contains(value, "sslmode=disable") {
		t.Fatal("isolated search path was not preserved")
	}
}
