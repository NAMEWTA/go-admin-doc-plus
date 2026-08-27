package settings_test

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

	"go-admin/internal/modules/iam/account"
	"go-admin/internal/modules/iam/administration"
	"go-admin/internal/modules/iam/authorization"
	sessionmigration "go-admin/internal/modules/iam/migrations/0010-session-schema"
	administrationmigration "go-admin/internal/modules/iam/migrations/0020-administration-schema"
	"go-admin/internal/modules/iam/session"
	"go-admin/internal/modules/settings"
	settingsmigration "go-admin/internal/modules/settings/migrations/0010-settings"
	"go-admin/internal/platform/config"
	"go-admin/internal/platform/database"
	"go-admin/internal/platform/migrations"
)

const (
	settingsServe   = "GO_ADMIN_SETTINGS_E2E_SERVE"
	settingsProfile = "GO_ADMIN_SETTINGS_E2E_PROFILE"
	settingsReady   = "GO_ADMIN_SETTINGS_E2E_READY_FILE"
	settingsStatic  = "GO_ADMIN_SETTINGS_E2E_STATIC_DIR"
	postgresDSN     = "GO_ADMIN_TEST_POSTGRES_DISPOSABLE_DSN"
	adminID         = "account-settings-admin"
)

type loginFactNoop struct{}

func (loginFactNoop) RecordLoginFact(context.Context, database.Tx, session.LoginFact) error {
	return nil
}

func TestSettingsBrowserHarnessServer(t *testing.T) {
	if os.Getenv(settingsServe) != "1" {
		t.Skip(settingsServe + " is not enabled")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	profile := os.Getenv(settingsProfile)
	opener, cleanup, expectedSchema := settingsDatabaseOpener(t, ctx, profile)
	defer cleanup()
	current := &switchingHandler{}
	type runtime struct {
		db       *database.Database
		sessions *session.Service
	}
	var live runtime
	var liveMu sync.RWMutex
	open := func() *database.Database {
		db := opener()
		if profile == "postgres" {
			var schema string
			if err := db.Bun().QueryRowContext(ctx, "SELECT current_schema()").Scan(&schema); err != nil || schema != expectedSchema {
				t.Fatalf("settings PostgreSQL schema isolation failed current=%q", schema)
			}
		}
		runner, err := migrations.NewRunner(sessionmigration.Provider{}, administrationmigration.Provider{}, settingsmigration.Provider{})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := runner.Up(ctx, db); err != nil {
			t.Fatal("settings harness migration failed")
		}
		seedAdmin(t, ctx, db)
		registry, err := authorization.NewCapabilityRegistry(db)
		if err != nil {
			t.Fatal(err)
		}
		if err := settings.RegisterCapabilities(ctx, registry); err != nil {
			t.Fatal("settings capability registration failed")
		}
		policy, err := config.NewSessionPolicy(2*time.Minute, 8*time.Minute, time.Minute)
		if err != nil {
			t.Fatal(err)
		}
		sessions, err := session.NewService(db, policy, session.WithLoginFactPort(loginFactNoop{}))
		if err != nil {
			t.Fatal(err)
		}
		service, err := settings.NewService(db)
		if err != nil {
			t.Fatal(err)
		}
		authenticator, err := settings.NewIAMSessionRequestAdapter(sessions)
		if err != nil {
			t.Fatal(err)
		}
		handler, err := settings.NewHTTPHandler(service, authenticator, func(*http.Request) string { return "0123456789abcdef" })
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
		api.Handle("/settings/", handler)
		api.Handle("/iam/session/", sessionHandler)
		api.Handle("/iam/administration/", adminHandler)
		current.set(http.StripPrefix("/api", api))
		liveMu.Lock()
		live = runtime{db: db, sessions: sessions}
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
		value := live
		liveMu.RUnlock()
		if body.Enabled {
			registry, err := authorization.NewCapabilityRegistry(value.db)
			if err != nil || settings.RegisterCapabilities(r.Context(), registry) != nil {
				w.WriteHeader(500)
				return
			}
		} else if _, err := value.db.Bun().ExecContext(r.Context(), "DELETE FROM iam_role_permissions WHERE role_id=? AND permission_code IN (?,?,?,?,?)", "role-system-admin", settings.PermissionValuesWrite, settings.PermissionValuesDelete, settings.PermissionDictionariesWrite, settings.PermissionDictionariesDelete, settings.PermissionOptionsRead); err != nil {
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
		value := live
		liveMu.RUnlock()
		if _, err := value.db.Bun().ExecContext(r.Context(), "UPDATE iam_roles SET data_scope=? WHERE id=?", body.Scope, "role-system-admin"); err != nil {
			w.WriteHeader(500)
			return
		}
		w.WriteHeader(204)
	})
	mux.HandleFunc("/__test/revoke-session", func(w http.ResponseWriter, r *http.Request) {
		liveMu.RLock()
		value := live
		liveMu.RUnlock()
		if r.Method != http.MethodPost {
			w.WriteHeader(405)
			return
		}
		if err := value.sessions.RevokeAccount(r.Context(), adminID); err != nil {
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
	staticRoot, ready := os.Getenv(settingsStatic), os.Getenv(settingsReady)
	if info, err := os.Stat(staticRoot); err != nil || !info.IsDir() || ready == "" {
		t.Fatal("settings harness paths invalid")
	}
	mux.Handle("/", http.FileServer(http.Dir(staticRoot)))
	server := httptest.NewUnstartedServer(mux)
	server.StartTLS()
	defer server.Close()
	if err := os.MkdirAll(filepath.Dir(ready), 0700); err != nil {
		t.Fatal("ready directory failed")
	}
	if err := os.WriteFile(ready, []byte(server.URL), 0600); err != nil {
		t.Fatal("ready file failed")
	}
	select {
	case <-shutdown:
	case <-ctx.Done():
		t.Fatal("settings browser harness timed out")
	}
}

type switchingHandler struct {
	mu      sync.RWMutex
	handler http.Handler
}

func (v *switchingHandler) set(handler http.Handler) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.handler = handler
}
func (v *switchingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	v.mu.RLock()
	handler := v.handler
	v.mu.RUnlock()
	handler.ServeHTTP(w, r)
}
func seedAdmin(t *testing.T, ctx context.Context, db *database.Database) {
	repository := account.NewRepository(db.Dialect())
	var count int
	if err := db.Bun().QueryRowContext(ctx, "SELECT COUNT(*) FROM iam_accounts WHERE id=?", adminID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count == 0 {
		hash, err := account.HashPassword("administrator password")
		if err != nil {
			t.Fatal(err)
		}
		if err := db.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error {
			return repository.Create(ctx, tx, account.Credential{Profile: account.Profile{ID: adminID, Username: "admin", DisplayName: "Administrator", Email: "admin@example.test"}, PasswordHash: hash}, time.Now().UTC())
		}); err != nil {
			t.Fatal("settings account seed failed")
		}
	}
	if _, err := db.Bun().ExecContext(ctx, "INSERT INTO iam_account_roles(account_id,role_id) VALUES(?,?) ON CONFLICT(account_id,role_id) DO NOTHING", adminID, "role-system-admin"); err != nil {
		t.Fatal("settings role seed failed")
	}
}
func settingsDatabaseOpener(t *testing.T, ctx context.Context, profile string) (func() *database.Database, func(), string) {
	if profile == "sqlite" {
		if os.Getenv(postgresDSN) != "" {
			t.Fatal("SQLite settings harness received PostgreSQL material")
		}
		path := filepath.Join(t.TempDir(), "settings.sqlite3")
		return func() *database.Database {
			db, err := database.NewProcess().Open(ctx, database.Config{Profile: config.ProfileServerSQLite, SQLitePath: path})
			if err != nil {
				t.Fatal("SQLite settings open failed")
			}
			return db
		}, func() {}, ""
	}
	if profile != "postgres" {
		t.Fatal("settings profile invalid")
	}
	dsn := os.Getenv(postgresDSN)
	if dsn == "" {
		t.Fatal("PostgreSQL settings material required")
	}
	admin, err := database.NewProcess().Open(ctx, database.Config{Profile: config.ProfileServerPostgres, PostgresDSN: dsn})
	if err != nil {
		t.Fatal("PostgreSQL settings admin failed")
	}
	schema := fmt.Sprintf("t10_settings_%d", time.Now().UnixNano())
	if _, err := admin.SQL().ExecContext(ctx, "CREATE SCHEMA "+schema); err != nil {
		_ = admin.Close()
		t.Fatal("PostgreSQL settings schema failed")
	}
	parsed, err := url.Parse(dsn)
	if err != nil {
		t.Fatal("PostgreSQL settings material invalid")
	}
	query := parsed.Query()
	query.Set("search_path", schema)
	parsed.RawQuery = query.Encode()
	open := func() *database.Database {
		db, err := database.NewProcess().Open(ctx, database.Config{Profile: config.ProfileServerPostgres, PostgresDSN: parsed.String()})
		if err != nil {
			t.Fatal("isolated PostgreSQL settings failed")
		}
		return db
	}
	cleanup := func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if _, err := admin.SQL().ExecContext(cleanupCtx, "DROP SCHEMA IF EXISTS "+schema+" CASCADE"); err != nil {
			t.Error("PostgreSQL settings cleanup failed")
		}
		_ = admin.Close()
	}
	return open, cleanup, schema
}
func TestSettingsHarnessSearchPathIsURLSafe(t *testing.T) {
	value := "postgresql://localhost/example?sslmode=disable"
	parsed, err := url.Parse(value)
	if err != nil {
		t.Fatal(err)
	}
	query := parsed.Query()
	query.Set("search_path", "t10_settings_contract")
	parsed.RawQuery = query.Encode()
	if !strings.Contains(parsed.String(), "search_path=t10_settings_contract") {
		t.Fatal("settings search path missing")
	}
}
