package demo_test

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"go-admin/internal/modules/demo"
	productsmigration "go-admin/internal/modules/demo/migrations/0010-products"
	"go-admin/internal/platform/config"
	"go-admin/internal/platform/database"
	"go-admin/internal/platform/migrations"
)

type allowAuthorizer struct {
	dialect database.Dialect
	scope   demo.Scope
	deny    map[string]bool
}

func (a allowAuthorizer) Dialect() database.Dialect { return a.dialect }
func (a allowAuthorizer) RequireInTx(_ context.Context, _ database.Tx, _ string, permission string) (demo.Scope, error) {
	if a.deny[permission] {
		return "", demo.ErrDenied
	}
	return a.scope, nil
}

func TestSQLiteCRUDConflictScopeAtomicDeleteAndRestart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "demo.sqlite3")
	db := openDatabase(t, path)
	service := newService(t, db, demo.ScopeAll, nil)
	first, err := service.Create(context.Background(), "owner-a", demo.ProductInput{SKU: "demo-01", Name: "First product", Description: "one", PriceCents: 123, Status: "active"})
	if err != nil {
		t.Fatal(err)
	}
	second, err := service.Create(context.Background(), "owner-b", demo.ProductInput{SKU: "DEMO-02", Name: "Second product", Description: "two", PriceCents: 99, Status: "inactive"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Create(context.Background(), "owner-a", demo.ProductInput{SKU: "demo-01", Name: "Duplicate SKU", PriceCents: 1, Status: "active"}); !errors.Is(err, demo.ErrConflict) {
		t.Fatalf("duplicate = %v", err)
	}
	page, err := service.List(context.Background(), "owner-a", demo.ListQuery{Search: "product", Page: 1, PageSize: 10, Sort: "priceCents", Direction: "ascending"})
	if err != nil || page.Total != 2 || len(page.Rows) != 2 || page.Rows[0].ID != second.ID {
		t.Fatalf("page = %#v err=%v", page, err)
	}
	updated, err := service.Update(context.Background(), "owner-a", first.ID, first.Revision, demo.ProductInput{SKU: first.SKU, Name: "Updated product", Description: "changed", PriceCents: 200, Status: "inactive"})
	if err != nil || updated.Revision != 2 {
		t.Fatalf("update = %#v err=%v", updated, err)
	}
	if _, err := service.Update(context.Background(), "owner-a", first.ID, 1, demo.ProductInput{SKU: first.SKU, Name: "Stale product", PriceCents: 1, Status: "active"}); !errors.Is(err, demo.ErrConflict) {
		t.Fatalf("stale update = %v", err)
	}
	if err := service.Delete(context.Background(), "owner-a", []demo.DeleteTarget{{ID: updated.ID, Revision: updated.Revision}, {ID: second.ID, Revision: 99}}); !errors.Is(err, demo.ErrConflict) {
		t.Fatalf("atomic delete = %v", err)
	}
	if _, err := service.Get(context.Background(), "owner-a", updated.ID); err != nil {
		t.Fatalf("first row was partially deleted: %v", err)
	}
	self := newService(t, db, demo.ScopeSelf, nil)
	visible, err := self.List(context.Background(), "owner-a", demo.ListQuery{Page: 1, PageSize: 20})
	if err != nil || visible.Total != 1 || visible.Rows[0].ID != updated.ID {
		t.Fatalf("self scope = %#v err=%v", visible, err)
	}
	if _, err := self.Get(context.Background(), "owner-a", second.ID); !errors.Is(err, demo.ErrNotFound) {
		t.Fatalf("foreign detail = %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db = openDatabase(t, path)
	service = newService(t, db, demo.ScopeAll, nil)
	got, err := service.Get(context.Background(), "owner-a", updated.ID)
	if err != nil || got.Name != "Updated product" {
		t.Fatalf("restarted product = %#v err=%v", got, err)
	}
	rows, err := db.Bun().QueryContext(context.Background(), `PRAGMA table_info(demo_products)`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	foundNameKey := false
	for rows.Next() {
		var position, notNull, primaryKey int
		var name, dataType string
		var defaultValue any
		if err := rows.Scan(&position, &name, &dataType, &notNull, &defaultValue, &primaryKey); err != nil {
			t.Fatal(err)
		}
		if strings.Contains(strings.ToLower(name), "tenant") {
			t.Fatalf("forbidden tenant column %q", name)
		}
		if name == "name_key" {
			foundNameKey = true
			if notNull != 1 {
				t.Fatal("name_key must be private and non-null")
			}
		}
	}
	if !foundNameKey {
		t.Fatal("deterministic name_key column is missing")
	}
}

func TestDeniedMutationHasNoStateChange(t *testing.T) {
	db := openDatabase(t, filepath.Join(t.TempDir(), "denied.sqlite3"))
	service := newService(t, db, demo.ScopeAll, map[string]bool{demo.PermissionProductsWrite: true})
	_, err := service.Create(context.Background(), "actor", demo.ProductInput{SKU: "DENIED", Name: "Denied product", Status: "active"})
	if !errors.Is(err, demo.ErrDenied) {
		t.Fatalf("create = %v", err)
	}
	var count int
	if err := db.Bun().QueryRowContext(context.Background(), `SELECT COUNT(*) FROM demo_products`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("count=%d err=%v", count, err)
	}
}

func TestSQLiteDialectCRUDContract(t *testing.T) {
	db := openDatabase(t, filepath.Join(t.TempDir(), "contract.sqlite3"))
	runDialectCRUDContract(t, db)
}

func runDialectCRUDContract(t *testing.T, db *database.Database) {
	t.Helper()
	all := newService(t, db, demo.ScopeAll, nil)
	now := time.Now().UTC()
	if _, err := db.Bun().ExecContext(context.Background(), `INSERT INTO demo_products
		(id, owner_account_id, sku, name, name_key, description, price_cents, status, revision, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, "00000000-0000-4000-8000-000000000098", "contract-owner-a", "BADKEY-01", "Valid name", "xy", "", 0, "active", 1, now, now); err == nil {
		t.Fatal("dialect accepted an invalid normalized name key")
	}
	first, err := all.Create(context.Background(), "contract-owner-a", demo.ProductInput{SKU: "CONTRACT-01", Name: "Contract product one", PriceCents: 10, Status: "active"})
	if err != nil {
		t.Fatal(err)
	}
	second, err := all.Create(context.Background(), "contract-owner-b", demo.ProductInput{SKU: "CONTRACT-02", Name: "Contract product two", PriceCents: 20, Status: "active"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := all.Create(context.Background(), "contract-owner-a", demo.ProductInput{SKU: "contract-01", Name: "Duplicate product", PriceCents: 1, Status: "active"}); !errors.Is(err, demo.ErrConflict) {
		t.Fatalf("unique conflict = %v", err)
	}
	if _, err := all.Update(context.Background(), "contract-owner-a", first.ID, first.Revision+1, demo.ProductInput{SKU: first.SKU, Name: "Stale product", PriceCents: 11, Status: "active"}); !errors.Is(err, demo.ErrConflict) {
		t.Fatalf("stale update = %v", err)
	}
	self := newService(t, db, demo.ScopeSelf, nil)
	if err := self.Delete(context.Background(), "contract-owner-a", []demo.DeleteTarget{{ID: second.ID, Revision: second.Revision}}); !errors.Is(err, demo.ErrDenied) {
		t.Fatalf("foreign self delete = %v", err)
	}
	if err := self.Delete(context.Background(), "contract-owner-a", []demo.DeleteTarget{{ID: first.ID, Revision: first.Revision + 1}}); !errors.Is(err, demo.ErrConflict) {
		t.Fatalf("owned stale self delete = %v", err)
	}
	if err := all.Delete(context.Background(), "contract-owner-a", []demo.DeleteTarget{{ID: first.ID, Revision: first.Revision}, {ID: second.ID, Revision: second.Revision + 1}}); !errors.Is(err, demo.ErrConflict) {
		t.Fatalf("batch conflict = %v", err)
	}
	if _, err := all.Get(context.Background(), "contract-owner-a", first.ID); err != nil {
		t.Fatalf("batch delete was not atomic: %v", err)
	}
	unicodeFixtures := []demo.ProductInput{
		{SKU: "UNICODE-01", Name: "界界界", Description: "Unicode fixture", PriceCents: 30, Status: "active"},
		{SKU: "UNICODE-02", Name: "😀Alpha", Description: "Astral fixture", PriceCents: 40, Status: "active"},
	}
	for _, fixture := range unicodeFixtures {
		if _, err := all.Create(context.Background(), "contract-owner-a", fixture); err != nil {
			t.Fatalf("Unicode fixture create = %v", err)
		}
	}
	searched, err := all.List(context.Background(), "contract-owner-a", demo.ListQuery{Search: "😀a", Page: 1, PageSize: 20, Sort: "name", Direction: "ascending"})
	if err != nil || searched.Total != 1 || len(searched.Rows) != 1 || searched.Rows[0].SKU != "UNICODE-02" {
		t.Fatalf("Unicode search = %#v err=%v", searched, err)
	}
	ordered, err := all.List(context.Background(), "contract-owner-a", demo.ListQuery{Search: "unicode-", Page: 1, PageSize: 20, Sort: "name", Direction: "ascending"})
	if err != nil || ordered.Total != 2 || len(ordered.Rows) != 2 || ordered.Rows[0].SKU != "UNICODE-01" || ordered.Rows[1].SKU != "UNICODE-02" {
		t.Fatalf("Unicode name ordering = %#v err=%v", ordered, err)
	}
	keyFixtures := []demo.ProductInput{
		{SKU: "KEY-PERCENT", Name: "% literal", Status: "active"},
		{SKU: "KEY-ASCII", Name: "<:@ collision", Status: "active"},
		{SKU: "KEY-UNDER", Name: "_ literal", Status: "active"},
		{SKU: "KEY-UNICODE", Name: "ä collision", Status: "active"},
	}
	for _, fixture := range keyFixtures {
		if _, err := all.Create(context.Background(), "contract-owner-a", fixture); err != nil {
			t.Fatalf("name-key fixture create = %v", err)
		}
	}
	for search, expectedSKU := range map[string]string{"ä": "KEY-UNICODE", "<:@": "KEY-ASCII", "%": "KEY-PERCENT", "_": "KEY-UNDER"} {
		found, err := all.List(context.Background(), "contract-owner-a", demo.ListQuery{Search: search, Page: 1, PageSize: 20, Sort: "name", Direction: "ascending"})
		if err != nil || found.Total != 1 || len(found.Rows) != 1 || found.Rows[0].SKU != expectedSKU {
			t.Fatalf("literal search %q = %#v err=%v", search, found, err)
		}
	}
	keyOrder, err := all.List(context.Background(), "contract-owner-a", demo.ListQuery{Search: "key-", Page: 1, PageSize: 20, Sort: "name", Direction: "ascending"})
	wantOrder := []string{"KEY-PERCENT", "KEY-ASCII", "KEY-UNDER", "KEY-UNICODE"}
	if err != nil || keyOrder.Total != len(wantOrder) || len(keyOrder.Rows) != len(wantOrder) {
		t.Fatalf("literal key ordering = %#v err=%v", keyOrder, err)
	}
	for index, want := range wantOrder {
		if keyOrder.Rows[index].SKU != want {
			t.Fatalf("literal key ordering[%d]=%q want=%q", index, keyOrder.Rows[index].SKU, want)
		}
	}
}

func TestUnknownScopeFailsClosedBeforeMutation(t *testing.T) {
	db := openDatabase(t, filepath.Join(t.TempDir(), "unknown-scope.sqlite3"))
	service := newService(t, db, demo.Scope("unexpected"), nil)
	if _, err := service.Create(context.Background(), "contract-owner", demo.ProductInput{SKU: "UNKNOWN-01", Name: "Unknown scope", Status: "active"}); !errors.Is(err, demo.ErrDenied) {
		t.Fatalf("unknown scope create = %v", err)
	}
	var count int
	if err := db.Bun().QueryRowContext(context.Background(), `SELECT COUNT(*) FROM demo_products`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("unknown scope mutated state count=%d err=%v", count, err)
	}
}

func openDatabase(t *testing.T, path string) *database.Database {
	t.Helper()
	db, err := database.NewProcess().Open(context.Background(), database.Config{Profile: config.ProfileServerSQLite, SQLitePath: path})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	runner, err := migrations.NewRunner(productsmigration.Provider{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := runner.Up(context.Background(), db); err != nil {
		t.Fatal(err)
	}
	return db
}

func newService(t *testing.T, db *database.Database, scope demo.Scope, deny map[string]bool) *demo.Service {
	t.Helper()
	service, err := demo.NewService(db, allowAuthorizer{dialect: db.Dialect(), scope: scope, deny: deny}, demo.WithClock(func() time.Time { return time.Date(2026, 8, 27, 1, 2, 3, 0, time.UTC) }))
	if err != nil {
		t.Fatal(err)
	}
	return service
}
