package main

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/app/product"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/platform/config"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/platform/database"
)

func TestPreviousMigrationBaselineAdvancesOnlyAudit(t *testing.T) {
	db := openFixtureDatabase(t)
	runner, err := previousMigrationRunner()
	if err != nil {
		t.Fatal(err)
	}
	if err := validatePreviousMigrationBaseline(runner); err != nil {
		t.Fatal(err)
	}
	previous, err := runner.Up(context.Background(), db)
	if err != nil || previous.Applied != 16 || previous.CurrentVersion != 8_000_000_000_000 {
		t.Fatalf("previous migration result = %#v, %v", previous, err)
	}
	current, err := product.NewMigrationRunner()
	if err != nil {
		t.Fatal(err)
	}
	upgraded, err := current.Up(context.Background(), db)
	if err != nil || upgraded.Applied != 1 || upgraded.CurrentVersion != 8_100_000_000_000 {
		t.Fatalf("current migration result = %#v, %v", upgraded, err)
	}
}

func TestMigrationFailureFixtureRetainsAuditConflict(t *testing.T) {
	db := openFixtureDatabase(t)
	runner, err := previousMigrationRunner()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := runner.Up(context.Background(), db); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Bun().ExecContext(context.Background(), `CREATE TABLE audit_facts (migration_fault TEXT NOT NULL)`); err != nil {
		t.Fatal(err)
	}
	current, err := product.NewMigrationRunner()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := current.Up(context.Background(), db); err == nil || err.Error() != "migration execution failed" {
		t.Fatalf("audit conflict error = %v", err)
	}
}

func openFixtureDatabase(t *testing.T) *database.Database {
	t.Helper()
	db, err := database.NewProcess().Open(context.Background(), database.Config{
		Profile:    config.ProfileDesktopSQLite,
		SQLitePath: filepath.Join(t.TempDir(), "fixture.db"),
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}
