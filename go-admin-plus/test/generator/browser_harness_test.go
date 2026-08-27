package generator_test

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
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
	"go-admin/internal/modules/generator"
	generatormigration "go-admin/internal/modules/generator/migrations/0010-config"
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

const generatorAdminID = "account-generator-admin"

type loginFactNoop struct{}

func (loginFactNoop) RecordLoginFact(context.Context, database.Tx, session.LoginFact) error {
	return nil
}

type harnessFailGate struct{}

func (harnessFailGate) Check(context.Context, string, generator.Preview) error {
	return errors.New("forced gate failure")
}
func (harnessFailGate) CompleteOutputGate() {}

type harnessAllowAuthorizer struct{}

func (harnessAllowAuthorizer) Require(context.Context, string, string) error { return nil }

func TestGeneratorBrowserHarnessServer(t *testing.T) {
	if os.Getenv("GO_ADMIN_GENERATOR_E2E_SERVE") != "1" {
		t.Skip("Generator browser host is disabled")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Minute)
	defer cancel()
	profile := os.Getenv("GO_ADMIN_GENERATOR_E2E_PROFILE")
	db, cleanup, currentSchema := openHarnessDatabase(t, ctx, profile)
	defer cleanup()
	defer db.Close()
	runner, err := migrations.NewRunner(sessionmigration.Provider{}, administrationmigration.Provider{}, productsmigration.Provider{}, generatormigration.Provider{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := runner.Up(ctx, db); err != nil {
		t.Fatal("Generator harness migration failed")
	}
	createGeneratorSourceTable(t, ctx, db)
	seedAdmin(t, ctx, db)
	registry, err := authorization.NewCapabilityRegistry(db)
	if err != nil {
		t.Fatal(err)
	}
	if err := generator.RegisterCapabilities(ctx, registry); err != nil {
		t.Fatal("Generator capability registration failed")
	}
	if err := demo.RegisterCapabilities(ctx, registry); err != nil {
		t.Fatal("products capability registration failed")
	}
	policy, _ := config.NewSessionPolicy(2*time.Minute, 8*time.Minute, time.Minute)
	sessions, err := session.NewService(db, policy, session.WithLoginFactPort(loginFactNoop{}))
	if err != nil {
		t.Fatal(err)
	}
	metadata, err := generator.NewSQLMetadataSource(ctx, db, generator.MetadataAllowlist{CurrentSchema: currentSchema, Tables: []string{"products"}})
	if err != nil {
		t.Fatal(err)
	}
	outputRoot := filepath.Join(t.TempDir(), "output")
	if err := os.Mkdir(outputRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	repositoryRoot, _ := filepath.Abs("../../..")
	gate, err := generator.NewWorkspaceCompileGate(repositoryRoot)
	if err != nil {
		t.Fatal(err)
	}
	writer, err := generator.NewAtomicWriter(outputRoot, gate)
	if err != nil {
		t.Fatal(err)
	}
	transportGenerator, err := generator.NewCanonicalTransportGenerator(repositoryRoot)
	if err != nil {
		t.Fatal(err)
	}
	renderer, err := generator.NewCanonicalRenderer(transportGenerator)
	if err != nil {
		t.Fatal(err)
	}
	authorizer, err := generator.NewIAMAuthorizationAdapter(db)
	if err != nil {
		t.Fatal(err)
	}
	store, err := generator.NewSQLConfigStore(db)
	if err != nil {
		t.Fatal(err)
	}
	service, err := generator.New(metadata, writer, authorizer, store, renderer, 10*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	authenticator, err := generator.NewIAMSessionRequestAdapter(sessions)
	if err != nil {
		t.Fatal(err)
	}
	generatorHandler, err := generator.NewHTTPHandler(service, authenticator, func(*http.Request) string { return "0123456789abcdef" })
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
	api.Handle("/generator/", generatorHandler)
	api.Handle("/iam/session/", sessionHandler)
	api.Handle("/iam/administration/", adminHandler)
	mux := http.NewServeMux()
	mux.Handle("/api/", http.StripPrefix("/api", api))
	var expectedMu sync.Mutex
	expectedDirectory := ""
	expectedHashes := map[string]string{}
	mux.HandleFunc("/__test/expected", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var request struct {
			Directory string `json:"directory"`
			Files     []struct {
				Path   string `json:"path"`
				SHA256 string `json:"sha256"`
			} `json:"files"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 256<<10)).Decode(&request); err != nil || request.Directory == "" || len(request.Files) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		hashes := make(map[string]string, len(request.Files))
		for _, file := range request.Files {
			if filepath.IsAbs(file.Path) || filepath.ToSlash(filepath.Clean(file.Path)) != file.Path || strings.HasPrefix(file.Path, "../") || len(file.SHA256) != 64 {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			hashes[file.Path] = file.SHA256
		}
		if len(hashes) != len(request.Files) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		expectedMu.Lock()
		expectedDirectory, expectedHashes = request.Directory, hashes
		expectedMu.Unlock()
		w.WriteHeader(http.StatusNoContent)
	})
	failRoot := filepath.Join(t.TempDir(), "failed-output")
	if err := os.Mkdir(failRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	failWriter, err := generator.NewAtomicWriter(failRoot, harnessFailGate{})
	if err != nil {
		t.Fatal(err)
	}
	failService, err := generator.New(metadata, failWriter, harnessAllowAuthorizer{}, store, renderer, 10*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	mux.HandleFunc("/__test/gate-failure", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(405)
			return
		}
		table, err := metadata.Describe(r.Context(), generator.TableRef{Schema: currentSchema, Name: "products"})
		if err != nil {
			w.WriteHeader(500)
			return
		}
		columns := make([]generator.ColumnDraft, 0, len(table.Columns))
		for _, value := range table.Columns {
			field := ""
			for _, part := range strings.Split(value.Name, "_") {
				if part == "id" {
					field += "ID"
				} else if part != "" {
					field += strings.ToUpper(part[:1]) + part[1:]
				}
			}
			columns = append(columns, generator.ColumnDraft{Name: value.Name, Field: field, Include: true, Searchable: value.Name == "name", Sortable: value.PrimaryKey})
		}
		value, err := failService.Preview(r.Context(), generatorAdminID, generator.Draft{Module: "gateprobe", Entity: "Probe", Plural: "probes", Table: table.Ref, Columns: columns})
		if err != nil {
			w.WriteHeader(500)
			return
		}
		if _, err := failService.Write(r.Context(), generatorAdminID, value.Token, true); !errors.Is(err, generator.ErrGateFailed) {
			w.WriteHeader(500)
			return
		}
		entries, _ := os.ReadDir(failRoot)
		if len(entries) != 0 {
			w.WriteHeader(500)
			return
		}
		w.WriteHeader(204)
	})
	mux.HandleFunc("/__test/output", func(w http.ResponseWriter, _ *http.Request) {
		entries, _ := os.ReadDir(outputRoot)
		failed, _ := os.ReadDir(failRoot)
		_ = json.NewEncoder(w).Encode(map[string]any{"entries": len(entries), "failedEntries": len(failed)})
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
	staticRoot, ready := os.Getenv("GO_ADMIN_GENERATOR_E2E_STATIC_DIR"), os.Getenv("GO_ADMIN_GENERATOR_E2E_READY_FILE")
	if info, err := os.Stat(staticRoot); err != nil || !info.IsDir() || ready == "" {
		t.Fatal("Generator harness paths are invalid")
	}
	mux.Handle("/", http.FileServer(http.Dir(staticRoot)))
	server := httptest.NewUnstartedServer(mux)
	server.StartTLS()
	defer server.Close()
	if err := os.MkdirAll(filepath.Dir(ready), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ready, []byte(server.URL), 0o600); err != nil {
		t.Fatal(err)
	}
	select {
	case <-shutdown:
	case <-ctx.Done():
		t.Fatal("Generator browser harness timed out")
	}
	entries, err := os.ReadDir(outputRoot)
	if err != nil || len(entries) != 1 {
		t.Fatalf("Generator output count=%d err=%v", len(entries), err)
	}
	expectedMu.Lock()
	directory, hashes := expectedDirectory, expectedHashes
	expectedMu.Unlock()
	if directory == "" || entries[0].Name() != directory || !entries[0].IsDir() {
		t.Fatalf("Generator output directory=%q expected=%q", entries[0].Name(), directory)
	}
	actual := map[string]string{}
	err = filepath.WalkDir(filepath.Join(outputRoot, directory), func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		info, infoErr := entry.Info()
		if infoErr != nil || !info.Mode().IsRegular() {
			return fmt.Errorf("non-regular generated output")
		}
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		relative, relativeErr := filepath.Rel(filepath.Join(outputRoot, directory), path)
		if relativeErr != nil {
			return relativeErr
		}
		actual[filepath.ToSlash(relative)] = fmt.Sprintf("%x", sha256.Sum256(content))
		return nil
	})
	if err != nil || len(actual) != len(hashes) {
		t.Fatalf("Generator output files=%d expected=%d err=%v", len(actual), len(hashes), err)
	}
	for path, hash := range hashes {
		if actual[path] != hash {
			t.Fatalf("Generator output hash mismatch: %s", path)
		}
	}
}

func createGeneratorSourceTable(t *testing.T, ctx context.Context, db *database.Database) {
	statement := `CREATE TABLE products (id TEXT PRIMARY KEY, name TEXT NOT NULL UNIQUE, enabled INTEGER NOT NULL, revision INTEGER NOT NULL, created_at TIMESTAMP NOT NULL, updated_at TIMESTAMP NOT NULL)`
	if db.Dialect() == database.DialectPostgres {
		statement = `CREATE TABLE products (id UUID PRIMARY KEY, name TEXT COLLATE "C" NOT NULL UNIQUE, enabled BOOLEAN NOT NULL, revision BIGINT NOT NULL, created_at TIMESTAMPTZ NOT NULL, updated_at TIMESTAMPTZ NOT NULL)`
	}
	if _, err := db.Bun().ExecContext(ctx, statement); err != nil {
		t.Fatal("Generator source table seed failed")
	}
}

func seedAdmin(t *testing.T, ctx context.Context, db *database.Database) {
	repository := account.NewRepository(db.Dialect())
	hash, err := account.HashPassword("administrator password")
	if err != nil {
		t.Fatal(err)
	}
	err = db.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error {
		return repository.Create(ctx, tx, account.Credential{Profile: account.Profile{ID: generatorAdminID, Username: "admin", DisplayName: "Admin", Email: "admin@generator.test"}, PasswordHash: hash}, time.Now().UTC())
	})
	if err != nil {
		t.Fatal("Generator account seed failed")
	}
	if _, err := db.Bun().ExecContext(ctx, `INSERT INTO iam_account_roles(account_id, role_id) VALUES (?, ?) ON CONFLICT(account_id, role_id) DO NOTHING`, generatorAdminID, "role-system-admin"); err != nil {
		t.Fatal("Generator role seed failed")
	}
}

func openHarnessDatabase(t *testing.T, ctx context.Context, profile string) (*database.Database, func(), string) {
	if profile == "sqlite" {
		path := filepath.Join(t.TempDir(), "generator.sqlite3")
		db, err := database.NewProcess().Open(ctx, database.Config{Profile: config.ProfileServerSQLite, SQLitePath: path})
		if err != nil {
			t.Fatal(err)
		}
		return db, func() {}, "main"
	}
	if profile != "postgres" {
		t.Fatal("Generator harness profile is invalid")
	}
	dsn := os.Getenv("GO_ADMIN_TEST_POSTGRES_DISPOSABLE_DSN")
	if dsn == "" {
		t.Fatal("PostgreSQL harness DSN is required")
	}
	admin, err := database.NewProcess().Open(ctx, database.Config{Profile: config.ProfileServerPostgres, PostgresDSN: dsn})
	if err != nil {
		t.Fatal("PostgreSQL admin open failed")
	}
	schema := fmt.Sprintf("t15_generator_%d", time.Now().UnixNano())
	if _, err := admin.SQL().ExecContext(ctx, `CREATE SCHEMA `+schema); err != nil {
		t.Fatal("PostgreSQL schema create failed")
	}
	parsed, err := url.Parse(dsn)
	if err != nil {
		t.Fatal("PostgreSQL DSN invalid")
	}
	query := parsed.Query()
	query.Set("search_path", schema)
	parsed.RawQuery = query.Encode()
	db, err := database.NewProcess().Open(ctx, database.Config{Profile: config.ProfileServerPostgres, PostgresDSN: parsed.String()})
	if err != nil {
		t.Fatal("isolated PostgreSQL open failed")
	}
	cleanup := func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = admin.SQL().ExecContext(cleanupCtx, `DROP SCHEMA IF EXISTS `+schema+` CASCADE`)
		_ = admin.Close()
	}
	return db, cleanup, schema
}

func TestGeneratorHarnessSchemaNameIsConstrained(t *testing.T) {
	if !strings.HasPrefix(fmt.Sprintf("t15_generator_%d", time.Now().UnixNano()), "t15_generator_") {
		t.Fatal("schema prefix drift")
	}
}
