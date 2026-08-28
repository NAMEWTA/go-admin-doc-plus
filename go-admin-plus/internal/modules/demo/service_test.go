package demo

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/contracts/capabilities"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/platform/database"
)

type fakeDatabase struct{ dialect database.Dialect }

func (value fakeDatabase) Dialect() database.Dialect { return value.dialect }
func (value fakeDatabase) WithinTx(ctx context.Context, callback func(context.Context, database.Tx) error) error {
	return callback(ctx, nil)
}

type captureRegistrar struct {
	capabilities capabilities.ModuleCapabilities
}

func (value *captureRegistrar) Register(_ context.Context, capabilities capabilities.ModuleCapabilities) error {
	value.capabilities = capabilities
	return nil
}

func TestDemoDeclaresStablePermissionsThroughIAMRegistryPort(t *testing.T) {
	registrar := &captureRegistrar{}
	if err := RegisterCapabilities(context.Background(), registrar); err != nil {
		t.Fatal(err)
	}
	want := []string{PermissionProductsRead, PermissionProductsWrite, PermissionProductsDelete}
	if len(registrar.capabilities.Permissions) != len(want) {
		t.Fatalf("definitions=%#v", registrar.capabilities.Permissions)
	}
	for index, code := range want {
		if registrar.capabilities.Permissions[index].Code != code || registrar.capabilities.Permissions[index].Name == "" {
			t.Fatalf("definition[%d]=%#v", index, registrar.capabilities.Permissions[index])
		}
	}
	if len(registrar.capabilities.Menus) != 1 || registrar.capabilities.Menus[0].Key != "demo-products" || registrar.capabilities.Menus[0].PermissionCode != PermissionProductsRead {
		t.Fatalf("menu=%#v", registrar.capabilities.Menus)
	}
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
	for _, value := range []ProductInput{
		{SKU: "RUNE-003", Name: strings.Repeat("😀", 3), Description: strings.Repeat("界", 500), Status: "active"},
		{SKU: "RUNE-120", Name: strings.Repeat("😀", 120), Status: "active"},
	} {
		if _, ok := normalizeInput(value); !ok {
			t.Fatalf("rejected rune boundary name=%d description=%d", runeLength(value.Name), runeLength(value.Description))
		}
	}
	if _, ok := normalizeInput(ProductInput{SKU: "RUNE-121", Name: strings.Repeat("😀", 121), Status: "active"}); ok {
		t.Fatal("accepted 121-rune name")
	}
	if _, ok := normalizeInput(ProductInput{SKU: "RUNE-501", Name: "Valid name", Description: strings.Repeat("😀", 501), Status: "active"}); ok {
		t.Fatal("accepted 501-rune description")
	}
	if _, ok := normalizeQuery(ListQuery{Search: strings.Repeat("😀", 100), Page: 1, PageSize: 20}); !ok {
		t.Fatal("rejected 100-rune search")
	}
	if _, ok := normalizeQuery(ListQuery{Search: strings.Repeat("😀", 101), Page: 1, PageSize: 20}); ok {
		t.Fatal("accepted 101-rune search")
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

func TestNormalizedNameKeyUsesDeterministicUnicodeText(t *testing.T) {
	ascii, unicode := normalizedNameKey("<:@"), normalizedNameKey("ä")
	if ascii != "<:@" || unicode != "ä" || strings.Contains(ascii, unicode) {
		t.Fatalf("ambiguous normalized keys ascii=%q unicode=%q", ascii, unicode)
	}
}
