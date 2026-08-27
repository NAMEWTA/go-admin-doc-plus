package demo_test

import (
	"context"
	"os"
	"testing"

	productsmigration "go-admin/internal/modules/demo/migrations/0010-products"
	"go-admin/internal/platform/migrations"
)

// TestPostgresCRUDContract is an environment-gated real-dialect asset. Source runs skip it;
// the Lead matrix supplies a disposable DSN and runs it with the browser profile.
func TestPostgresCRUDContract(t *testing.T) {
	if os.Getenv(postgresEnvironment) == "" {
		t.Skip(postgresEnvironment + " is not configured")
	}
	opener, cleanup, expectedSchema := demoDatabaseOpener(t, context.Background(), "postgres")
	defer cleanup()
	db := opener()
	defer func() { _ = db.Close() }()
	var schema string
	if err := db.Bun().QueryRowContext(context.Background(), `SELECT current_schema()`).Scan(&schema); err != nil || schema != expectedSchema {
		t.Fatalf("PostgreSQL contract is not isolated schema=%q err=%v", schema, err)
	}
	runner, err := migrations.NewRunner(productsmigration.Provider{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := runner.Up(context.Background(), db); err != nil {
		t.Fatal(err)
	}
	runDialectCRUDContract(t, db)
}
