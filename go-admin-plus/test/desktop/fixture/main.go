package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go-admin/internal/modules/demo"
	productsmigration "go-admin/internal/modules/demo/migrations/0010-products"
	"go-admin/internal/modules/iam/account"
	"go-admin/internal/modules/iam/authorization"
	sessionmigration "go-admin/internal/modules/iam/migrations/0010-session-schema"
	administrationmigration "go-admin/internal/modules/iam/migrations/0020-administration-schema"
	"go-admin/internal/platform/config"
	"go-admin/internal/platform/database"
	"go-admin/internal/platform/migrations"
	reliablemigration "go-admin/internal/platform/migrations/reliable-runtime"
)

const fixturePassword = "administrator password"

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
	runner, err := migrations.NewRunner(
		reliablemigration.Provider{}, sessionmigration.Provider{}, administrationmigration.Provider{},
		productsmigration.Provider{},
	)
	if err != nil {
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
