package demo_test

import (
	"context"
	"os"
	"testing"

	"go-admin/internal/modules/demo"
	productsmigration "go-admin/internal/modules/demo/migrations/0010-products"
	"go-admin/internal/platform/migrations"
)

// TestPostgresCRUDContract is an environment-gated real-dialect asset. Source runs skip it;
// the Lead matrix supplies a disposable DSN and runs it with the browser profile.
func TestPostgresCRUDContract(t *testing.T) {
	if os.Getenv(postgresEnvironment) == "" {
		t.Skip(postgresEnvironment + " is not configured")
	}
	opener, cleanup := demoDatabaseOpener(t, context.Background(), "postgres")
	defer cleanup()
	db := opener()
	defer func() { _ = db.Close() }()
	runner, err := migrations.NewRunner(productsmigration.Provider{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := runner.Up(context.Background(), db); err != nil {
		t.Fatal(err)
	}
	service := newService(t, db, demo.ScopeAll, nil)
	created, err := service.Create(context.Background(), "postgres-owner", demo.ProductInput{SKU: "PG-DEMO", Name: "PostgreSQL product", PriceCents: 7, Status: "active"})
	if err != nil {
		t.Fatal(err)
	}
	updated, err := service.Update(context.Background(), "postgres-owner", created.ID, created.Revision, demo.ProductInput{SKU: created.SKU, Name: "PostgreSQL updated", PriceCents: 8, Status: "inactive"})
	if err != nil || updated.Revision != 2 {
		t.Fatalf("update=%#v err=%v", updated, err)
	}
	if err := service.Delete(context.Background(), "postgres-owner", []demo.DeleteTarget{{ID: updated.ID, Revision: updated.Revision}}); err != nil {
		t.Fatal(err)
	}
}
