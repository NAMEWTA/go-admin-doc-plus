package generator

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/platform/database"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	_ "modernc.org/sqlite"
)

type staticMetadata struct{ table Table }

func (metadata staticMetadata) Tables(context.Context) ([]TableRef, error) {
	return []TableRef{metadata.table.Ref}, nil
}
func (metadata staticMetadata) Describe(_ context.Context, ref TableRef) (Table, error) {
	if ref != metadata.table.Ref {
		return Table{}, ErrDenied
	}
	return metadata.table, nil
}

type sqliteMetadataDatabase struct{ db *bun.DB }

func (value sqliteMetadataDatabase) Dialect() database.Dialect { return database.DialectSQLite }
func (value sqliteMetadataDatabase) Bun() *bun.DB              { return value.db }

type sqliteConfigDatabase struct{ db *bun.DB }

func (value sqliteConfigDatabase) Dialect() database.Dialect { return database.DialectSQLite }
func (value sqliteConfigDatabase) WithinTx(ctx context.Context, fn func(context.Context, database.Tx) error) error {
	return value.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error { return fn(ctx, tx) })
}

type failingGate struct{}

func (failingGate) Check(context.Context, string, Preview) error {
	return errors.New("compiler rejected output")
}
func (failingGate) CompleteOutputGate() {}

type allowAuthorizer struct{}

func (allowAuthorizer) Require(_ context.Context, actorID, _ string) error {
	if actorID == "" {
		return ErrDenied
	}
	return nil
}

type permissionAuthorizer map[string]bool

func (permissions permissionAuthorizer) Require(_ context.Context, actorID, permission string) error {
	if actorID == "" || !permissions[permission] {
		return ErrDenied
	}
	return nil
}

type mutatingGate struct{}

func (mutatingGate) Check(_ context.Context, root string, preview Preview) error {
	path, _ := safeJoin(root, preview.Files[0].Path)
	return os.WriteFile(path, []byte("mutated\n"), 0o640)
}
func (mutatingGate) CompleteOutputGate() {}

type passingGate struct{}

func (passingGate) Check(context.Context, string, Preview) error { return nil }
func (passingGate) CompleteOutputGate()                          {}

type rootReplacingGate struct {
	root, moved, replacement string
}

func (gate rootReplacingGate) Check(context.Context, string, Preview) error {
	if err := os.Rename(gate.root, gate.moved); err != nil {
		return err
	}
	return os.Symlink(gate.replacement, gate.root)
}
func (rootReplacingGate) CompleteOutputGate() {}

type blockingGate struct {
	entered chan struct{}
	release chan struct{}
}

func (gate blockingGate) Check(ctx context.Context, _ string, _ Preview) error {
	close(gate.entered)
	select {
	case <-gate.release:
		return nil
	case <-ctx.Done():
		return context.Cause(ctx)
	}
}
func (blockingGate) CompleteOutputGate() {}

type revocableAuthorizer struct {
	mu      sync.Mutex
	allowed bool
}

func (value *revocableAuthorizer) Require(_ context.Context, _ string, _ string) error {
	value.mu.Lock()
	defer value.mu.Unlock()
	if !value.allowed {
		return ErrDenied
	}
	return nil
}
func (value *revocableAuthorizer) revoke() { value.mu.Lock(); value.allowed = false; value.mu.Unlock() }

type memoryConfigStore struct{}

func (memoryConfigStore) Save(context.Context, string, Draft, Preview) error { return nil }
func (memoryConfigStore) Get(context.Context, string, string) (Draft, string, error) {
	return Draft{}, "", ErrNotFound
}

type baseRenderer struct{}

func (baseRenderer) Render(_ context.Context, model Model) ([]PreviewFile, error) {
	return renderBase(model)
}

func testTable() Table {
	return Table{Ref: TableRef{Schema: "main", Name: "products"}, Columns: []Column{
		{Name: "updated_at", DatabaseType: "TEXT", Kind: KindTime, Nullable: false, Ordinal: 6},
		{Name: "id", DatabaseType: "TEXT", Kind: KindUUID, PrimaryKey: true, Ordinal: 1},
		{Name: "name", DatabaseType: "TEXT", Kind: KindString, Ordinal: 2},
		{Name: "enabled", DatabaseType: "BOOLEAN", Kind: KindBoolean, Ordinal: 3},
		{Name: "revision", DatabaseType: "INTEGER", Kind: KindInt64, Ordinal: 4},
		{Name: "created_at", DatabaseType: "TEXT", Kind: KindTime, Ordinal: 5},
	}}
}

func testDraft() Draft {
	return Draft{Module: "catalog", Entity: "Product", Plural: "products", Table: testTable().Ref, Columns: []ColumnDraft{
		{Name: "enabled", Field: "Enabled", Include: true, Sortable: true},
		{Name: "id", Field: "ID", Include: true},
		{Name: "updated_at", Field: "UpdatedAt", Include: true, Sortable: true},
		{Name: "name", Field: "Name", Include: true, Searchable: true, Sortable: true},
		{Name: "revision", Field: "Revision", Include: true},
		{Name: "created_at", Field: "CreatedAt", Include: true},
	}}
}

func TestGeneratedServiceDeduplicatesPrimarySortAndUsesConditionalSearchImports(t *testing.T) {
	draft := testDraft()
	for index := range draft.Columns {
		draft.Columns[index].Sortable = draft.Columns[index].Name == "id"
	}
	model, err := normalize(testTable(), draft)
	if err != nil {
		t.Fatal(err)
	}
	files, err := renderBase(model)
	if err != nil {
		t.Fatal(err)
	}
	content := make(map[string]string, len(files))
	for _, file := range files {
		content[file.Path] = file.Content
	}
	service := content["go-admin-plus/internal/modules/catalog/service.go"]
	if strings.Count(service, `case "id":`) != 1 {
		t.Fatal("generated service duplicated the primary sort key")
	}
	sqliteTest := content["go-admin-plus/internal/modules/catalog/service_test.go"]
	if strings.Contains(sqliteTest, `"fmt"`) || strings.Contains(sqliteTest, `"sort"`) {
		t.Fatal("generated SQLite test retained unused search imports")
	}
	postgresTest := content["go-admin-plus/internal/modules/catalog/service_postgres_test.go"]
	if !strings.Contains(postgresTest, `"fmt"`) || strings.Contains(postgresTest, `"sort"`) {
		t.Fatal("generated PostgreSQL test imports do not match their use")
	}
}

func TestPreviewIsDeterministicAndUsesOneNormalizedModel(t *testing.T) {
	root := t.TempDir()
	writer, err := NewAtomicWriter(root, passingGate{})
	if err != nil {
		t.Fatal(err)
	}
	generator, err := New(staticMetadata{table: testTable()}, writer, allowAuthorizer{}, memoryConfigStore{}, baseRenderer{}, 5*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	generator.now = func() time.Time { return time.Date(2026, 8, 27, 0, 0, 0, 0, time.UTC) }

	first, err := generator.Preview(context.Background(), "actor-one", testDraft())
	if err != nil {
		t.Fatal(err)
	}
	second, err := generator.Preview(context.Background(), "actor-one", testDraft())
	if err != nil {
		t.Fatal(err)
	}
	if first.Token == second.Token || first.Digest != second.Digest {
		t.Fatalf("preview token/digest contract: %#v %#v", first, second)
	}
	for index := range first.Files {
		if first.Files[index] != second.Files[index] {
			t.Fatalf("file %d drifted", index)
		}
	}
	if !containsPath(first.Files, "contracts/openapi/modules/catalog.yaml") ||
		!containsPath(first.Files, "go-admin-plus/internal/modules/catalog/repository.go") ||
		!containsPath(first.Files, "go-admin-plus-ui/packages/web-domains/catalog/src/ProductPage.vue") {
		t.Fatalf("missing architecture outputs: %#v", first.Files)
	}
	result, err := generator.Write(context.Background(), "actor-one", first.Token, true)
	if err != nil {
		t.Fatal(err)
	}
	if result.Directory != "catalog-"+first.Digest[:12] {
		t.Fatalf("unexpected output directory %q", result.Directory)
	}
	for _, file := range first.Files {
		content, readErr := os.ReadFile(filepath.Join(root, result.Directory, filepath.FromSlash(file.Path)))
		if readErr != nil || string(content) != file.Content {
			t.Fatalf("written preview differs for %s: %v", file.Path, readErr)
		}
	}
}

func TestConfigAndPreviewRequireMetadataAndPreviewPermissions(t *testing.T) {
	writer, err := NewAtomicWriter(t.TempDir(), passingGate{})
	if err != nil {
		t.Fatal(err)
	}
	permissions := permissionAuthorizer{PermissionPreview: true}
	generator, err := New(staticMetadata{table: testTable()}, writer, permissions, memoryConfigStore{}, baseRenderer{}, 5*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := generator.Config(context.Background(), "actor-one", "catalog"); !errors.Is(err, ErrDenied) {
		t.Fatalf("config without metadata permission = %v", err)
	}
	if _, err := generator.Preview(context.Background(), "actor-one", testDraft()); !errors.Is(err, ErrDenied) {
		t.Fatalf("preview without metadata permission = %v", err)
	}
	permissions[PermissionMetadataRead] = true
	if _, err := generator.Preview(context.Background(), "actor-one", testDraft()); err != nil {
		t.Fatalf("preview with combined permissions = %v", err)
	}
}

func TestCanonicalRendererInvokesLintAndGeneration(t *testing.T) {
	model, err := normalize(testTable(), testDraft())
	if err != nil {
		t.Fatal(err)
	}
	repositoryRoot, err := filepath.Abs("../../../..")
	if err != nil {
		t.Fatal(err)
	}
	transport, err := NewCanonicalTransportGenerator(repositoryRoot)
	if err != nil {
		t.Fatal(err)
	}
	renderer, err := NewCanonicalRenderer(transport)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	files, err := renderer.Render(ctx, model)
	if err != nil {
		diagnosticRoot := t.TempDir()
		contractPath := filepath.Join(diagnosticRoot, "contracts/openapi/modules/catalog.yaml")
		if mkdirErr := os.MkdirAll(filepath.Dir(contractPath), 0o750); mkdirErr != nil {
			t.Fatal(mkdirErr)
		}
		if writeErr := os.WriteFile(contractPath, []byte(generatedOpenAPI(model)), 0o640); writeErr != nil {
			t.Fatal(writeErr)
		}
		if copyErr := copyDirectory(filepath.Join(repositoryRoot, "contracts/openapi/components"), filepath.Join(diagnosticRoot, "contracts/openapi/components")); copyErr != nil {
			t.Fatal(copyErr)
		}
		outputRoot := filepath.Join(diagnosticRoot, "output")
		if mkdirErr := os.Mkdir(outputRoot, 0o750); mkdirErr != nil {
			t.Fatal(mkdirErr)
		}
		cliURL := (&url.URL{Scheme: "file", Path: filepath.Join(repositoryRoot, "scripts/contracts/cli.mjs")}).String()
		script := `const { generate, lintContracts } = await import(process.argv[1]); lintContracts([process.argv[3]]); generate(process.argv[2], [process.argv[3]]);`
		command := exec.CommandContext(ctx, "node", "--input-type=module", "-e", script, cliURL, outputRoot, contractPath)
		command.Dir = repositoryRoot
		command.Env = minimalCommandEnvironment(filepath.Join(diagnosticRoot, "home"))
		output, _ := command.CombinedOutput()
		t.Fatalf("renderer: %v; canonical generate: %s", err, output)
	}
	for _, path := range []string{
		"go-admin-plus/internal/modules/catalog/transport/openapi.gen.go",
		"go-admin-plus/internal/modules/catalog/transport/openapi.json",
		"go-admin-plus-ui/packages/domains/catalog/src/generated/client.ts",
		"go-admin-plus-ui/packages/domains/catalog/src/generated/schema.ts",
	} {
		if !containsPath(files, path) {
			t.Fatalf("canonical output is missing %s", path)
		}
	}
	fixture := t.TempDir()
	safeHome := filepath.Join(fixture, ".home")
	if err := os.Mkdir(safeHome, 0o700); err != nil {
		t.Fatal(err)
	}
	environment := minimalCommandEnvironment(safeHome)
	if err := copyTrackedSkeleton(ctx, repositoryRoot, fixture); err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		path := filepath.Join(fixture, filepath.FromSlash(file.Path))
		if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(file.Content), 0o640); err != nil {
			t.Fatal(err)
		}
	}
	lockCommand := exec.CommandContext(ctx, pnpmExecutable(), "install", "--lockfile-only", "--no-frozen-lockfile", "--ignore-scripts")
	lockCommand.Dir = filepath.Join(fixture, "go-admin-plus-ui")
	lockCommand.Env = environment
	if output, err := lockCommand.CombinedOutput(); err != nil {
		t.Fatalf("fixture lock importer: %v\n%s", err, output)
	}
	for _, arguments := range [][]string{
		{"scripts/contracts/cli.mjs", "lint", "--contract", "contracts/openapi/modules/catalog.yaml"},
		{"scripts/contracts/cli.mjs", "generate"},
		{"scripts/contracts/cli.mjs", "generate", "--check"},
	} {
		command := exec.CommandContext(ctx, "node", arguments...)
		command.Dir = fixture
		command.Env = environment
		if output, err := command.CombinedOutput(); err != nil {
			t.Fatalf("node %s: %v\n%s", strings.Join(arguments, " "), err, output)
		}
	}
	for _, arguments := range [][]string{{"test", "./internal/modules/catalog/..."}, {"build", "./internal/modules/catalog/..."}} {
		command := exec.CommandContext(ctx, "go", arguments...)
		command.Dir = filepath.Join(fixture, "go-admin-plus")
		command.Env = append([]string(nil), environment...)
		if dsn := os.Getenv("GO_ADMIN_GENERATOR_POSTGRES_DSN"); dsn != "" && arguments[0] == "test" {
			command.Env = append(command.Env, "GO_ADMIN_GENERATOR_POSTGRES_DSN="+dsn)
			if marker := os.Getenv("GO_ADMIN_GENERATOR_POSTGRES_MARKER"); marker != "" {
				command.Env = append(command.Env, "GO_ADMIN_GENERATOR_POSTGRES_MARKER="+marker)
			}
		}
		if output, err := command.CombinedOutput(); err != nil {
			if arguments[0] == "test" {
				if source, readErr := os.ReadFile(filepath.Join(fixture, "go-admin-plus/internal/modules/catalog/service_postgres_test.go")); readErr == nil {
					lines := strings.Split(string(source), "\n")
					for index := 44; index < len(lines) && index < 56; index++ {
						t.Logf("generated PG %d: %s", index+1, lines[index])
					}
				}
			}
			t.Fatalf("go %s: %v\n%s", strings.Join(arguments, " "), err, output)
		}
	}
	uiRoot := filepath.Join(fixture, "go-admin-plus-ui")
	for _, arguments := range [][]string{
		{"install", "--frozen-lockfile", "--ignore-scripts"},
		{"--filter", "@go-admin-plus/domain-catalog", "typecheck"},
		{"--filter", "@go-admin-plus/domain-catalog", "test"},
		{"--filter", "@go-admin-plus/web-domain-catalog", "typecheck"},
		{"--filter", "@go-admin-plus/web-domain-catalog", "test"},
		{"check:workspace"},
	} {
		command := exec.CommandContext(ctx, pnpmExecutable(), arguments...)
		command.Dir = uiRoot
		command.Env = environment
		if output, err := command.CombinedOutput(); err != nil {
			t.Fatalf("pnpm %s: %v\n%s", strings.Join(arguments, " "), err, output)
		}
	}
	gate, err := NewWorkspaceCompileGate(repositoryRoot)
	if err != nil {
		t.Fatal(err)
	}
	preview := signedPreview(model.Module, files)
	gateCtx, gateCancel := context.WithTimeout(context.Background(), compileGateTimeout+time.Minute)
	defer gateCancel()
	if err := gate.Check(gateCtx, t.TempDir(), preview); err != nil {
		t.Fatalf("production compile gate: %v", err)
	}
}

func TestCompileSkeletonRejectsEnvironmentFiles(t *testing.T) {
	for _, path := range []string{".env", "go-admin-plus/.env.local", "go-admin-plus-ui/apps/admin-web/.env.production"} {
		if allowedSkeletonPath(path) {
			t.Fatalf("compile skeleton accepted %s", path)
		}
	}
}

func TestConfigStoreOwnershipUpsertAndErrorClassification(t *testing.T) {
	sqlDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	db := bun.NewDB(sqlDB, sqlitedialect.New())
	defer db.Close()
	if _, err := db.Exec(`CREATE TABLE generator_configs (
		id TEXT PRIMARY KEY, actor_account_id TEXT NOT NULL, module_name TEXT NOT NULL UNIQUE,
		source_schema TEXT NOT NULL, source_table TEXT NOT NULL, normalized_config TEXT NOT NULL,
		preview_digest TEXT NOT NULL, created_at TEXT NOT NULL, updated_at TEXT NOT NULL
	)`); err != nil {
		t.Fatal(err)
	}
	store, err := NewSQLConfigStore(sqliteConfigDatabase{db: db})
	if err != nil {
		t.Fatal(err)
	}
	preview := Preview{Module: "catalog", Digest: strings.Repeat("a", 64)}
	if err := store.Save(context.Background(), "actor-one", testDraft(), preview); err != nil {
		t.Fatal(err)
	}
	preview.Digest = strings.Repeat("b", 64)
	if err := store.Save(context.Background(), "actor-one", testDraft(), preview); err != nil {
		t.Fatalf("owner upsert: %v", err)
	}
	if err := store.Save(context.Background(), "actor-two", testDraft(), preview); !errors.Is(err, ErrDenied) {
		t.Fatalf("cross-actor upsert: %v", err)
	}
	_, digest, err := store.Get(context.Background(), "actor-one", "catalog")
	if err != nil || digest != preview.Digest {
		t.Fatalf("owner get digest=%q err=%v", digest, err)
	}
	if _, _, err := store.Get(context.Background(), "actor-two", "catalog"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("cross-actor get: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.Get(context.Background(), "actor-one", "catalog"); !errors.Is(err, ErrInternal) {
		t.Fatalf("database outage: %v", err)
	}
}

func TestConfigStatementsUseBunDialectPlaceholders(t *testing.T) {
	statement := configSaveStatement(database.DialectPostgres)
	if strings.Count(statement, "?") != 9 || strings.Contains(statement, "$9") {
		t.Fatalf("invalid Bun statement: %s", statement)
	}
}

func TestNullableOnlyBusinessInputHasValidUpdateRequiredList(t *testing.T) {
	table := testTable()
	for index := range table.Columns {
		if table.Columns[index].Name == "name" || table.Columns[index].Name == "enabled" {
			table.Columns[index].Nullable = true
		}
	}
	model, err := normalize(table, testDraft())
	if err != nil {
		t.Fatal(err)
	}
	contract := generatedOpenAPI(model)
	if !strings.Contains(contract, "required: [revision]") || strings.Contains(contract, "required: [, revision]") {
		t.Fatalf("invalid nullable-only update required list")
	}
}

func TestMinimalCommandEnvironmentDoesNotInheritCredentialsOrHome(t *testing.T) {
	for _, item := range minimalCommandEnvironment("") {
		key, _, _ := strings.Cut(item, "=")
		if key == "HOME" || key == "USERPROFILE" || strings.Contains(key, "TOKEN") || strings.Contains(key, "PASSWORD") || strings.Contains(key, "DSN") {
			t.Fatalf("unsafe subprocess environment key %s", key)
		}
	}
	safeHome := t.TempDir()
	joined := strings.Join(minimalCommandEnvironment(safeHome), "\n")
	if !strings.Contains(joined, "HOME="+safeHome) || strings.Contains(joined, "HOME="+os.Getenv("HOME")+"\n") {
		t.Fatalf("subprocess HOME is not isolated")
	}
}

func TestNormalizationRejectsIdentifierAndColumnConfusion(t *testing.T) {
	if lowerFirst("ID") != "id" || lowerFirst("URLValue") != "urlValue" || lowerFirst("Product") != "product" {
		t.Fatal("generated JSON identifier normalization is unstable")
	}
	cases := []Draft{
		func() Draft { value := testDraft(); value.Module = "../catalog"; return value }(),
		func() Draft { value := testDraft(); value.Entity = "product"; return value }(),
		func() Draft { value := testDraft(); value.Columns[0].Field = "ID"; return value }(),
		func() Draft { value := testDraft(); value.Columns = value.Columns[:3]; return value }(),
		func() Draft { value := testDraft(); value.Columns[1].Include = false; return value }(),
	}
	for index, draft := range cases {
		if _, err := normalize(testTable(), draft); !errors.Is(err, ErrInvalid) {
			t.Fatalf("case %d: %v", index, err)
		}
	}
}

func TestAtomicWriterRejectsTraversalOverwriteTamperAndPartialFailure(t *testing.T) {
	root := t.TempDir()
	writer, err := NewAtomicWriter(root, failingGate{})
	if err != nil {
		t.Fatal(err)
	}
	preview := signedPreview("catalog", []PreviewFile{{Path: "module/file.txt", Content: "stable\n"}})
	if _, err := writer.Write(context.Background(), preview); !errors.Is(err, ErrGateFailed) {
		t.Fatalf("gate failure: %v", err)
	}
	entries, _ := os.ReadDir(root)
	if len(entries) != 0 {
		t.Fatalf("gate left partial output: %#v", entries)
	}

	plain, _ := NewAtomicWriter(root, passingGate{})
	tampered := clonePreview(preview)
	tampered.Files[0].Content = "changed\n"
	if _, err := plain.Write(context.Background(), tampered); !errors.Is(err, ErrPreviewStale) {
		t.Fatalf("tamper: %v", err)
	}
	traversal := signedPreview("catalog", []PreviewFile{{Path: "../escaped.txt", Content: "bad"}})
	if _, err := plain.Write(context.Background(), traversal); !errors.Is(err, ErrInvalid) {
		t.Fatalf("traversal: %v", err)
	}
	if _, err := plain.Write(context.Background(), preview); err != nil {
		t.Fatal(err)
	}
	if _, err := plain.Write(context.Background(), preview); !errors.Is(err, ErrConflict) {
		t.Fatalf("overwrite: %v", err)
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(root), "escaped.txt")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("escape created: %v", err)
	}
}

func TestAtomicWriterConcurrentSameDigestCommitsExactlyOnce(t *testing.T) {
	root := t.TempDir()
	writer, err := NewAtomicWriter(root, passingGate{})
	if err != nil {
		t.Fatal(err)
	}
	preview := signedPreview("catalog", []PreviewFile{{Path: "module/file.txt", Content: "stable\n"}})
	start := make(chan struct{})
	results := make(chan error, 2)
	for range 2 {
		go func() {
			<-start
			_, err := writer.Write(context.Background(), preview)
			results <- err
		}()
	}
	close(start)
	var successes, conflicts int
	for range 2 {
		err := <-results
		switch {
		case err == nil:
			successes++
		case errors.Is(err, ErrConflict):
			conflicts++
		default:
			t.Fatalf("unexpected concurrent write result: %v", err)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("successes=%d conflicts=%d", successes, conflicts)
	}
	content, err := os.ReadFile(filepath.Join(root, "catalog-"+preview.Digest[:12], "module/file.txt"))
	if err != nil || string(content) != "stable\n" {
		t.Fatalf("committed content=%q err=%v", content, err)
	}
}

func TestAtomicWriterRejectsOutputRootReplacementDuringGate(t *testing.T) {
	parent := t.TempDir()
	root := filepath.Join(parent, "output")
	moved := filepath.Join(parent, "moved-output")
	replacement := filepath.Join(parent, "replacement")
	for _, directory := range []string{root, replacement} {
		if err := os.Mkdir(directory, 0o750); err != nil {
			t.Fatal(err)
		}
	}
	writer, err := NewAtomicWriter(root, rootReplacingGate{root: root, moved: moved, replacement: replacement})
	if err != nil {
		t.Fatal(err)
	}
	preview := signedPreview("catalog", []PreviewFile{{Path: "module/file.txt", Content: "stable\n"}})
	if _, err := writer.Write(context.Background(), preview); !errors.Is(err, ErrInvalid) {
		t.Fatalf("root replacement: %v", err)
	}
	for _, directory := range []string{moved, replacement} {
		entries, readErr := os.ReadDir(directory)
		if readErr != nil || len(entries) != 0 {
			t.Fatalf("root replacement published in %s: entries=%v err=%v", directory, entries, readErr)
		}
	}
}

func TestPreviewConfirmationIsActorBoundExpiringAndOneShot(t *testing.T) {
	clock := time.Date(2026, 8, 27, 0, 0, 0, 0, time.UTC)
	writer, _ := NewAtomicWriter(t.TempDir(), failingGate{})
	generator, _ := New(staticMetadata{table: testTable()}, writer, allowAuthorizer{}, memoryConfigStore{}, baseRenderer{}, time.Minute)
	generator.now = func() time.Time { return clock }
	preview, err := generator.Preview(context.Background(), "actor-one", testDraft())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := generator.Write(context.Background(), "actor-two", preview.Token, true); !errors.Is(err, ErrPreviewStale) {
		t.Fatalf("cross-actor confirmation: %v", err)
	}
	if _, err := generator.Write(context.Background(), "actor-one", preview.Token, true); !errors.Is(err, ErrGateFailed) {
		t.Fatalf("cross-actor attempt consumed owner token: %v", err)
	}
	preview, _ = generator.Preview(context.Background(), "actor-one", testDraft())
	clock = clock.Add(time.Minute)
	if _, err := generator.Write(context.Background(), "actor-one", preview.Token, true); !errors.Is(err, ErrPreviewStale) {
		t.Fatalf("expired token: %v", err)
	}
	preview, _ = generator.Preview(context.Background(), "actor-one", testDraft())
	if _, err := generator.Write(context.Background(), "actor-one", preview.Token, false); !errors.Is(err, ErrInvalid) {
		t.Fatalf("unconfirmed write: %v", err)
	}
}

func TestWriteReauthorizesAfterGateBeforeAtomicPublish(t *testing.T) {
	root := t.TempDir()
	gate := blockingGate{entered: make(chan struct{}), release: make(chan struct{})}
	writer, err := NewAtomicWriter(root, gate)
	if err != nil {
		t.Fatal(err)
	}
	authorizer := &revocableAuthorizer{allowed: true}
	generator, err := New(staticMetadata{table: testTable()}, writer, authorizer, memoryConfigStore{}, baseRenderer{}, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	preview, err := generator.Preview(context.Background(), "actor-one", testDraft())
	if err != nil {
		t.Fatal(err)
	}
	result := make(chan error, 1)
	go func() {
		_, err := generator.Write(context.Background(), "actor-one", preview.Token, true)
		result <- err
	}()
	<-gate.entered
	authorizer.revoke()
	close(gate.release)
	if err := <-result; !errors.Is(err, ErrDenied) {
		t.Fatalf("final authorization: %v", err)
	}
	entries, err := os.ReadDir(root)
	if err != nil || len(entries) != 0 {
		t.Fatalf("revoked write published output: entries=%v err=%v", entries, err)
	}
}

func TestGateCannotMutatePreviewOrAddOutput(t *testing.T) {
	root := t.TempDir()
	writer, _ := NewAtomicWriter(root, mutatingGate{})
	preview := signedPreview("catalog", []PreviewFile{{Path: "module/file.txt", Content: "stable\n"}})
	if _, err := writer.Write(context.Background(), preview); !errors.Is(err, ErrGateFailed) {
		t.Fatalf("mutating gate: %v", err)
	}
	entries, _ := os.ReadDir(root)
	if len(entries) != 0 {
		t.Fatalf("mutating gate left output: %#v", entries)
	}
}

func TestWriterRejectsSymlinkRoot(t *testing.T) {
	parent := t.TempDir()
	realRoot := filepath.Join(parent, "real")
	if err := os.Mkdir(realRoot, 0o750); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(parent, "link")
	if err := os.Symlink(realRoot, link); err != nil {
		t.Fatal(err)
	}
	if _, err := NewAtomicWriter(link); !errors.Is(err, ErrInvalid) {
		t.Fatalf("symlink root: %v", err)
	}
}

func TestSQLiteMetadataIsCurrentProfileAndAllowlistOnly(t *testing.T) {
	sqlDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()
	if _, err := sqlDB.Exec(`CREATE TABLE products (id TEXT PRIMARY KEY, name TEXT NOT NULL, quantity INTEGER, generated TEXT GENERATED ALWAYS AS (name) VIRTUAL); CREATE TABLE private_notes (id TEXT PRIMARY KEY)`); err != nil {
		t.Fatal(err)
	}
	db := bun.NewDB(sqlDB, sqlitedialect.New())
	source, err := NewSQLMetadataSource(context.Background(), sqliteMetadataDatabase{db: db}, MetadataAllowlist{CurrentSchema: "main", Tables: []string{"products"}})
	if err != nil {
		t.Fatal(err)
	}
	tables, err := source.Tables(context.Background())
	if err != nil || len(tables) != 1 || tables[0].Name != "products" {
		t.Fatalf("tables=%#v err=%v", tables, err)
	}
	table, err := source.Describe(context.Background(), TableRef{Schema: "main", Name: "products"})
	if err != nil {
		t.Fatal(err)
	}
	if len(table.Columns) != 3 || table.Columns[0].Name != "id" || !table.Columns[0].PrimaryKey || table.Columns[1].Nullable {
		t.Fatalf("columns=%#v", table.Columns)
	}
	if _, err := source.Describe(context.Background(), TableRef{Schema: "main", Name: "private_notes"}); !errors.Is(err, ErrDenied) {
		t.Fatalf("unauthorized table: %v", err)
	}
	if _, err := NewSQLMetadataSource(context.Background(), sqliteMetadataDatabase{db: db}, MetadataAllowlist{CurrentSchema: "sqlite_master", Tables: []string{"products"}}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("system schema: %v", err)
	}
}

func TestSystemSchemasAndUnsupportedTypesAreRejected(t *testing.T) {
	for _, schema := range []string{"pg_catalog", "information_schema", "pg_temp_1", "Main", "../main"} {
		if validMetadataSchema(database.DialectPostgres, schema) {
			t.Fatalf("accepted system/invalid schema %q", schema)
		}
	}
	if _, err := postgresKind("xml"); !errors.Is(err, ErrInvalid) {
		t.Fatalf("unsupported postgres type: %v", err)
	}
	if _, err := sqliteKind("GEOMETRY"); !errors.Is(err, ErrInvalid) {
		t.Fatalf("unsupported sqlite type: %v", err)
	}
}

func containsPath(files []PreviewFile, path string) bool {
	for _, file := range files {
		if file.Path == path {
			return true
		}
	}
	return false
}

func signedPreview(module string, files []PreviewFile) Preview {
	files = append([]PreviewFile(nil), files...)
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	digest := sha256.New()
	for index := range files {
		hash := sha256.Sum256([]byte(files[index].Content))
		files[index].SHA256 = hex.EncodeToString(hash[:])
		digest.Write([]byte(files[index].Path))
		digest.Write([]byte{0})
		digest.Write(hash[:])
	}
	return Preview{Module: module, Token: strings.Repeat("a", 64), Digest: hex.EncodeToString(digest.Sum(nil)), Files: files}
}

func TestGeneratedOutputHasNoLegacyOrDeepImports(t *testing.T) {
	model, err := normalize(testTable(), testDraft())
	if err != nil {
		t.Fatal(err)
	}
	files, err := renderBase(model)
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.ToLower(func() string {
		var values []string
		for _, file := range files {
			values = append(values, file.Path+"\n"+file.Content)
		}
		return strings.Join(values, "\n")
	}())
	for _, forbidden := range []string{"gorm.io", "casbin", "redis", "tenant_id", "@go-admin-plus/domain-demo/src"} {
		if strings.Contains(joined, forbidden) {
			t.Fatalf("generated legacy/deep import %q", forbidden)
		}
	}
}
