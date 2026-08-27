package settings

import (
	"context"
	"errors"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"go-admin/internal/modules/iam/authorization"
	settingsmigration "go-admin/internal/modules/settings/migrations/0010-settings"
	"go-admin/internal/platform/config"
	"go-admin/internal/platform/database"
	"go-admin/internal/platform/migrations"
)

type testAuthorizer struct {
	dialect database.Dialect
	scope   Scope
	denied  map[string]bool
}

func (a testAuthorizer) Dialect() database.Dialect { return a.dialect }
func (a testAuthorizer) RequireInTx(_ context.Context, _ database.Tx, _ string, permission string) (Scope, error) {
	if a.denied[permission] {
		return "", ErrDenied
	}
	return a.scope, nil
}

func sqliteSettings(t *testing.T) (*database.Database, *Service) {
	t.Helper()
	db, err := database.NewProcess().Open(context.Background(), database.Config{Profile: config.ProfileServerSQLite, SQLitePath: filepath.Join(t.TempDir(), "settings.sqlite3")})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	runner, err := migrations.NewRunner(settingsmigration.Provider{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := runner.Up(context.Background(), db); err != nil {
		t.Fatal(err)
	}
	service, err := newService(db, testAuthorizer{dialect: db.Dialect(), scope: ScopeAll})
	if err != nil {
		t.Fatal(err)
	}
	return db, service
}

func TestSettingsSQLiteCRUDSearchConflictReferenceAndOptions(t *testing.T) {
	db, service := sqliteSettings(t)
	ctx := context.Background()
	actor := "account-admin"
	first, err := service.CreateSetting(ctx, actor, SettingInput{Category: CategoryBusiness, Key: "shop.title", Label: "界商店", Value: "Public title", Description: "Visible heading", Enabled: true})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.CreateSetting(ctx, actor, SettingInput{Category: CategoryBusiness, Key: first.Key, Label: "Duplicate", Value: "x"}); !errors.Is(err, ErrConflict) {
		t.Fatalf("duplicate=%v", err)
	}
	updated, err := service.UpdateSetting(ctx, actor, first.ID, first.Revision, SettingInput{Category: first.Category, Key: first.Key, Label: "😀 Store", Value: "New title", Description: "Updated", Enabled: true})
	if err != nil || updated.Revision != 2 {
		t.Fatalf("updated=%#v err=%v", updated, err)
	}
	if _, err := service.UpdateSetting(ctx, actor, first.ID, 1, SettingInput{Category: first.Category, Key: first.Key, Label: "stale", Value: "ignored"}); !errors.Is(err, ErrConflict) {
		t.Fatalf("stale=%v", err)
	}
	for _, fixture := range []SettingInput{{Category: CategoryBusiness, Key: "literal.percent", Label: "% literal", Value: "one"}, {Category: CategoryBusiness, Key: "literal.under", Label: "_ literal", Value: "two"}, {Category: CategoryBusiness, Key: "literal.ascii", Label: "<:@ collision", Value: "three"}, {Category: CategoryBusiness, Key: "literal.unicode", Label: "ä collision", Value: "four"}} {
		if _, err := service.CreateSetting(ctx, actor, fixture); err != nil {
			t.Fatal(err)
		}
	}
	for search, want := range map[string]string{"%": "literal.percent", "_": "literal.under", "<:@": "literal.ascii", "ä": "literal.unicode"} {
		page, err := service.ListSettings(ctx, actor, CategoryBusiness, ListQuery{Search: search, Page: 1, PerPage: 20})
		if err != nil || page.Total != 1 || page.Rows[0].Key != want {
			t.Fatalf("search %q=%#v err=%v", search, page, err)
		}
	}
	ordered, err := service.ListSettings(ctx, actor, CategoryBusiness, ListQuery{Search: "literal.", Page: 1, PerPage: 20})
	if err != nil {
		t.Fatal(err)
	}
	got := []string{}
	for _, row := range ordered.Rows {
		got = append(got, row.Key)
	}
	if !reflect.DeepEqual(got, []string{"literal.percent", "literal.ascii", "literal.under", "literal.unicode"}) {
		t.Fatalf("order=%v", got)
	}
	dictionary, err := service.CreateDictionary(ctx, actor, DictionaryInput{Key: "order.status", Name: "订单状态", Description: "Public workflow", Enabled: true})
	if err != nil {
		t.Fatal(err)
	}
	enabled, err := service.CreateItem(ctx, actor, dictionary.ID, DictionaryItemInput{Value: "paid", Label: "Paid", SortOrder: 20, Enabled: true})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.CreateItem(ctx, actor, dictionary.ID, DictionaryItemInput{Value: "draft", Label: "Draft", SortOrder: 10, Enabled: false}); err != nil {
		t.Fatal(err)
	}
	options, err := service.Options(ctx, actor, dictionary.Key)
	if err != nil || !reflect.DeepEqual(options, []DictionaryOption{{Value: "paid", Label: "Paid"}}) {
		t.Fatalf("options=%#v err=%v", options, err)
	}
	if err := service.DeleteDictionary(ctx, actor, dictionary.ID, dictionary.Revision); !errors.Is(err, ErrConflict) {
		t.Fatalf("referenced delete=%v", err)
	}
	if err := service.DeleteItem(ctx, actor, enabled.ID, enabled.Revision); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.Bun().QueryRowContext(ctx, "SELECT COUNT(*) FROM settings_values").Scan(&count); err != nil || count != 5 {
		t.Fatalf("count=%d err=%v", count, err)
	}
}

func TestSettingsRejectsSensitiveMaterialAndScopesBeforeMutation(t *testing.T) {
	db, _ := sqliteSettings(t)
	observer := &observationCapture{}
	service, err := newService(db, testAuthorizer{dialect: db.Dialect(), scope: ScopeAll}, WithObserver(observer))
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	inputs := []SettingInput{{Category: CategoryBusiness, Key: "database_url", Label: "Endpoint", Value: "postgres://user@host/db"}, {Category: CategoryUI, Key: "theme.token", Label: "Access token", Value: "visible"}, {Category: CategoryBusiness, Key: "runtime.profile", Label: "Profile", Value: "server-postgres"}, {Category: CategoryBusiness, Key: "welcome.text", Label: "Welcome", Value: "-----BEGIN PRIVATE KEY-----"}, {Category: CategoryBusiness, Key: "session-policy", Label: "Idle timeout", Value: "900"},
		{Category: CategoryBusiness, Key: "public.note", Label: "Pass/word", Value: "visible"}, {Category: CategoryBusiness, Key: "public.jwt", Label: "Credential", Value: "aaaaaaaa.bbbbbbbb.cccccccc"}, {Category: CategoryBusiness, Key: "public.bearer", Label: "Header", Value: "Bearer abcdefghijklmnop"}, {Category: CategoryBusiness, Key: "public.url", Label: "Endpoint", Value: "https://alice:credential@example.test/path"}, {Category: CategoryBusiness, Key: "public.opaque", Label: "Opaque", Value: strings.Repeat("A", 43)}}
	for _, input := range inputs {
		if _, err := service.CreateSetting(ctx, "actor", input); !errors.Is(err, ErrSensitive) {
			t.Fatalf("accepted sensitive input %#v: %v", input, err)
		}
	}
	denied, err := newService(db, testAuthorizer{dialect: db.Dialect(), scope: Scope("self")})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := denied.CreateDictionary(ctx, "actor", DictionaryInput{Key: "public.options", Name: "Public options"}); !errors.Is(err, ErrDenied) {
		t.Fatalf("self scope=%v", err)
	}
	var values, dictionaries int
	if err := db.Bun().QueryRowContext(ctx, "SELECT COUNT(*) FROM settings_values").Scan(&values); err != nil {
		t.Fatal(err)
	}
	if err := db.Bun().QueryRowContext(ctx, "SELECT COUNT(*) FROM settings_dictionary_types").Scan(&dictionaries); err != nil {
		t.Fatal(err)
	}
	if values != 0 || dictionaries != 0 {
		t.Fatalf("state changed values=%d dictionaries=%d", values, dictionaries)
	}
	if len(observer.values) != len(inputs) {
		t.Fatalf("signals=%#v", observer.values)
	}
	for _, value := range observer.values {
		if value.Outcome != "sensitive_rejected" {
			t.Fatalf("signal=%#v", value)
		}
	}
}

func TestSettingsRejectsReservedRuntimeAndSecretKeysWithoutBroadTextFalsePositives(t *testing.T) {
	rejected := []string{
		"api_key", "database.dsn", "client.credential", "runtime.log_level",
		"server.listen_address", "checkout.secret", "session.timeout",
	}
	for _, key := range rejected {
		if _, err := normalizeSetting(SettingInput{Category: CategoryBusiness, Key: key, Label: "Public label", Value: "visible"}); !errors.Is(err, ErrSensitive) {
			t.Fatalf("accepted reserved key %q: %v", key, err)
		}
	}
	allowed := []SettingInput{
		{Category: CategoryBusiness, Key: "shipping.tokenized_label", Label: "Tokenized label", Value: "visible"},
		{Category: CategoryBusiness, Key: "business.logistics_level", Label: "Credential-free checkout", Value: "visible"},
	}
	for _, input := range allowed {
		if _, err := normalizeSetting(input); err != nil {
			t.Fatalf("rejected ordinary business setting %#v: %v", input, err)
		}
	}
}

func TestSettingsUnicodeBoundariesAndCapabilities(t *testing.T) {
	for _, count := range []int{1, 120} {
		if _, err := normalizeDictionary(DictionaryInput{Key: "unicode.key", Name: strings.Repeat("😀", count)}); err != nil {
			t.Fatalf("rejected %d runes: %v", count, err)
		}
	}
	if _, err := normalizeDictionary(DictionaryInput{Key: "unicode.key", Name: strings.Repeat("😀", 121)}); !errors.Is(err, ErrValidation) {
		t.Fatal("accepted 121 runes")
	}
	if _, err := normalizeSetting(SettingInput{Category: CategoryUI, Key: "public.banner", Label: "Banner", Value: strings.Repeat("界", 500)}); err != nil {
		t.Fatal(err)
	}
	if _, err := normalizeSetting(SettingInput{Category: CategoryUI, Key: "public.banner", Label: "Banner", Value: strings.Repeat("界", 501)}); !errors.Is(err, ErrValidation) {
		t.Fatal("accepted 501 runes")
	}
	capture := &capabilityCapture{}
	if err := RegisterCapabilities(context.Background(), capture); err != nil {
		t.Fatal(err)
	}
	if len(capture.value.Permissions) != 7 || len(capture.value.Menus) != 2 {
		t.Fatalf("capabilities=%#v", capture.value)
	}
}

type capabilityCapture struct {
	value authorization.ModuleCapabilities
}

type observationCapture struct{ values []Observation }

func (c *observationCapture) Observe(value Observation) { c.values = append(c.values, value) }

func (c *capabilityCapture) Register(_ context.Context, value authorization.ModuleCapabilities) error {
	c.value = value
	return nil
}
