package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/app/product"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/demo"
	productsmigration "github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/demo/migrations/0010-products"
	filesaccountlifecyclemigration "github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/files/account_lifecycle_migration"
	filesmigration "github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/files/migrations/0010-files"
	capacitymigration "github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/files/migrations/0020-capacity"
	configmigration "github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/generator/migrations/0010-config"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/iam/account"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/iam/authorization"
	sessionmigration "github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/iam/migrations/0010-session-schema"
	administrationmigration "github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/iam/migrations/0020-administration-schema"
	bootstraprecoverymigration "github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/iam/migrations/0030-bootstrap-recovery"
	sessionprotectionmigration "github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/iam/migrations/0040-session-protection"
	datascopemigration "github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/iam/migrations/0050-data-scope"
	accountlifecyclemigration "github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/iam/migrations/0060-account-lifecycle"
	organizationmigration "github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/organization/migrations"
	schedulermigration "github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/scheduler/migrations"
	settingsmigration "github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/settings/migrations/0010-settings"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/platform/config"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/platform/database"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/platform/migrations"
	reliablemigration "github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/platform/migrations/reliable-runtime"
)

const fixturePassword = "administrator password"
const pendingAuditMigration = "8100000000000_audit.sql"

func main() {
	root := flag.String("root", "", "isolated native E2E root")
	mode := flag.String("mode", "previous", "previous, migration-failure, or verify")
	expectedProduct := flag.String("expected-product", "", "product name required by verify mode")
	flag.Parse()
	if flag.NArg() != 0 || *root == "" || !filepath.IsAbs(*root) {
		fail()
	}
	canonical, err := filepath.EvalSymlinks(filepath.Clean(*root))
	if err != nil || canonical != filepath.Clean(*root) {
		fail()
	}
	data := filepath.Join(canonical, "data")
	logs := filepath.Join(canonical, "logs")
	if err := os.MkdirAll(data, 0o700); err != nil {
		fail()
	}
	if err := os.MkdirAll(logs, 0o700); err != nil {
		fail()
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	process := database.NewProcess()
	db, err := process.Open(ctx, database.Config{Profile: config.ProfileDesktopSQLite, SQLitePath: filepath.Join(data, "go-admin-plus.db")})
	if err != nil {
		fail()
	}
	defer db.Close()
	if *mode == "verify" {
		verifyFixture(ctx, db, data, *expectedProduct)
		return
	}
	if *mode != "previous" && *mode != "migration-failure" {
		fail()
	}
	runner, err := previousMigrationRunner()
	if err != nil || validatePreviousMigrationBaseline(runner) != nil {
		fail()
	}
	if _, err := runner.Up(ctx, db); err != nil {
		fail()
	}
	registry, err := authorization.NewCapabilityRegistry(db)
	if err != nil || demo.RegisterCapabilities(ctx, registry) != nil {
		fail()
	}
	hash, err := account.HashPassword(fixturePassword)
	if err != nil {
		fail()
	}
	repository := account.NewRepository(db.Dialect())
	err = db.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error {
		return repository.Create(ctx, tx, account.Credential{
			Profile:      account.Profile{ID: "account-desktop-e2e", Username: "admin", DisplayName: "Administrator", Email: "admin@example.test"},
			PasswordHash: hash,
		}, time.Now().UTC())
	})
	if err != nil && !errors.Is(err, account.ErrConflict) {
		fail()
	}
	if _, err := db.Bun().ExecContext(ctx, `INSERT INTO iam_account_roles(account_id, role_id) VALUES (?, ?) ON CONFLICT(account_id, role_id) DO NOTHING`, "account-desktop-e2e", "role-system-admin"); err != nil {
		fail()
	}
	if *mode == "migration-failure" {
		if _, err := db.Bun().ExecContext(ctx, `CREATE TABLE audit_facts (migration_fault TEXT NOT NULL)`); err != nil {
			fail()
		}
	}
	_, _ = fmt.Fprintln(os.Stdout, `{"state":"ready"}`)
}

func previousMigrationRunner() (*migrations.Runner, error) {
	return migrations.NewRunner(
		sessionmigration.Provider{},
		administrationmigration.Provider{},
		bootstraprecoverymigration.Provider{},
		sessionprotectionmigration.Provider{},
		datascopemigration.Provider{},
		accountlifecyclemigration.Provider{},
		organizationmigration.Provider{},
		productsmigration.Provider{},
		configmigration.Provider{},
		settingsmigration.Provider{},
		schedulermigration.Provider{},
		filesmigration.Provider{},
		capacitymigration.Provider{},
		filesaccountlifecyclemigration.Provider{},
		reliablemigration.Provider{},
	)
}

func validatePreviousMigrationBaseline(previous *migrations.Runner) error {
	current, err := product.NewMigrationRunner()
	if err != nil {
		return err
	}
	previousNames, err := composedMigrationNames(previous)
	if err != nil {
		return err
	}
	currentNames, err := composedMigrationNames(current)
	if err != nil {
		return err
	}
	for name := range previousNames {
		if _, exists := currentNames[name]; !exists {
			return errors.New("desktop fixture migration baseline invalid")
		}
		delete(currentNames, name)
	}
	if len(currentNames) != 1 {
		return errors.New("desktop fixture migration baseline invalid")
	}
	if _, exists := currentNames[pendingAuditMigration]; !exists {
		return errors.New("desktop fixture migration baseline invalid")
	}
	return nil
}

func composedMigrationNames(runner *migrations.Runner) (map[string]struct{}, error) {
	if runner == nil {
		return nil, errors.New("desktop fixture migration baseline invalid")
	}
	composed, err := runner.Compose(database.DialectSQLite)
	if err != nil {
		return nil, err
	}
	entries, err := fs.ReadDir(composed, ".")
	if err != nil {
		return nil, err
	}
	names := make(map[string]struct{}, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			return nil, errors.New("desktop fixture migration baseline invalid")
		}
		names[entry.Name()] = struct{}{}
	}
	return names, nil
}

func verifyFixture(ctx context.Context, db *database.Database, data, expectedProduct string) {
	var version int64
	if err := db.Bun().NewRaw(`SELECT COALESCE(MAX(version_id), 0) FROM goose_db_version WHERE is_applied = 1`).Scan(ctx, &version); err != nil || version != 8100000000000 {
		fail()
	}
	if expectedProduct != "" {
		var count int
		if err := db.Bun().NewRaw(`SELECT COUNT(*) FROM demo_products WHERE name = ?`, expectedProduct).Scan(ctx, &count); err != nil || count != 1 {
			fail()
		}
	}
	entries, err := os.ReadDir(filepath.Join(data, "backups"))
	if err != nil || len(entries) == 0 {
		fail()
	}
	_, _ = fmt.Fprintln(os.Stdout, `{"state":"verified","version":8100000000000}`)
}

func fail() {
	_, _ = fmt.Fprintln(os.Stderr, "desktop fixture failed")
	os.Exit(1)
}
