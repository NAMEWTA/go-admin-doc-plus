package demo_test

import (
	"context"
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
	testCookie          = "__Host-demo_test"
)

// TestDemoBrowserHarnessServer is compiled by source gates and run only by the Lead matrix.
func TestDemoBrowserHarnessServer(t *testing.T) {
	if os.Getenv(demoServe) != "1" {
		t.Skip(demoServe + " is not enabled")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	profile := os.Getenv(demoProfile)
	opener, cleanup := demoDatabaseOpener(t, ctx, profile)
	defer cleanup()
	current := &switchingHandler{}
	open := func() *database.Database {
		db := opener()
		runner, err := migrations.NewRunner(productsmigration.Provider{})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := runner.Up(ctx, db); err != nil {
			t.Fatal("demo harness migration failed")
		}
		service, err := demo.NewService(db, harnessAuthorizer{dialect: db.Dialect()})
		if err != nil {
			t.Fatal(err)
		}
		handler, err := demo.NewHTTPHandler(service, harnessAuthenticator{}, testCookie, func(*http.Request) string { return "0123456789abcdef" })
		if err != nil {
			t.Fatal(err)
		}
		current.set(http.StripPrefix("/api", handler))
		return db
	}
	db := open()
	defer func() { _ = db.Close() }()
	mux := http.NewServeMux()
	mux.Handle("/api/demo/", current)
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
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: testCookie, Value: "browser-session", Path: "/", Secure: true, HttpOnly: true, SameSite: http.SameSiteStrictMode})
		files.ServeHTTP(w, r)
	}))
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

type harnessAuthorizer struct{ dialect database.Dialect }

func (value harnessAuthorizer) Dialect() database.Dialect { return value.dialect }
func (harnessAuthorizer) RequireInTx(context.Context, database.Tx, string, string) (demo.Scope, error) {
	return demo.ScopeAll, nil
}

type harnessAuthenticator struct{}

func (harnessAuthenticator) AuthorizeRequest(_ context.Context, token, csrf string, mutation bool) (demo.RequestIdentity, error) {
	if token != "browser-session" {
		return demo.RequestIdentity{}, demo.ErrAuthentication
	}
	if mutation && csrf != strings.Repeat("c", 32) {
		return demo.RequestIdentity{}, demo.ErrCSRF
	}
	return demo.RequestIdentity{ActorID: "demo-browser-admin", CSRF: strings.Repeat("c", 32)}, nil
}

func demoDatabaseOpener(t *testing.T, ctx context.Context, profile string) (func() *database.Database, func()) {
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
		}, func() {}
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
	return open, cleanup
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
