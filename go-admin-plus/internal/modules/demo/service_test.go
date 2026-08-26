package demo

import (
	"context"
	"errors"
	"testing"
	"time"

	"go-admin/internal/platform/database"
)

type fakeDatabase struct{ dialect database.Dialect }

func (value fakeDatabase) Dialect() database.Dialect { return value.dialect }
func (value fakeDatabase) WithinTx(ctx context.Context, callback func(context.Context, database.Tx) error) error {
	return callback(ctx, nil)
}

type fakeAuthorizer struct {
	dialect database.Dialect
	scope   Scope
	err     error
}

func (value fakeAuthorizer) Dialect() database.Dialect { return value.dialect }
func (value fakeAuthorizer) RequireInTx(context.Context, database.Tx, string, string) (Scope, error) {
	return value.scope, value.err
}

func TestNewServiceRejectsDialectMismatch(t *testing.T) {
	_, err := NewService(fakeDatabase{database.DialectSQLite}, fakeAuthorizer{dialect: database.DialectPostgres})
	if err == nil {
		t.Fatal("expected dialect mismatch")
	}
}

func TestInputAndQueryValidation(t *testing.T) {
	input, valid := normalizeInput(ProductInput{SKU: " demo-01 ", Name: " Product ", Description: " note ", PriceCents: 42, Status: " ACTIVE "})
	if !valid || input.SKU != "DEMO-01" || input.Name != "Product" || input.Status != "active" {
		t.Fatalf("normalized input = %#v valid=%v", input, valid)
	}
	for _, invalid := range []ProductInput{{SKU: "ab", Name: "Product", Status: "active"}, {SKU: "ABC", Name: "x", Status: "active"}, {SKU: "ABC", Name: "Product", PriceCents: -1, Status: "active"}, {SKU: "ABC", Name: "Product", Status: "deleted"}} {
		if _, ok := normalizeInput(invalid); ok {
			t.Fatalf("accepted invalid input %#v", invalid)
		}
	}
	if _, ok := normalizeQuery(ListQuery{Page: MaximumPage + 1, PageSize: 20}); ok {
		t.Fatal("accepted overflowing page")
	}
	if _, ok := normalizeQuery(ListQuery{Page: 1, PageSize: 20, Sort: "price_cents desc; drop table"}); ok {
		t.Fatal("accepted unlisted sort")
	}
}

func TestNormalizePreservesContextSentinels(t *testing.T) {
	service := &Service{db: fakeDatabase{database.DialectSQLite}, now: time.Now}
	for _, sentinel := range []error{context.Canceled, context.DeadlineExceeded} {
		if err := service.normalize(context.Background(), errors.Join(errors.New("driver detail"), sentinel)); !errors.Is(err, sentinel) {
			t.Fatalf("lost sentinel %v: %v", sentinel, err)
		}
	}
}
