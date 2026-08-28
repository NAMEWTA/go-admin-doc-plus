package settings

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	settingsmigration "github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/settings/migrations/0010-settings"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/platform/config"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/platform/database"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/platform/migrations"
)

const postgresTestDSN = "GO_ADMIN_TEST_POSTGRES_DISPOSABLE_DSN"

func TestSettingsPostgresDialectContract(t *testing.T) {
	dsn := os.Getenv(postgresTestDSN)
	if dsn == "" {
		t.Skip(postgresTestDSN + " is not configured")
	}
	ctx := context.Background()
	admin, err := database.NewProcess().Open(ctx, database.Config{Profile: config.ProfileServerPostgres, PostgresDSN: dsn})
	if err != nil {
		t.Fatal("PostgreSQL administrator open failed")
	}
	schema := fmt.Sprintf("t10_settings_%d", time.Now().UnixNano())
	if _, err := admin.SQL().ExecContext(ctx, "CREATE SCHEMA "+schema); err != nil {
		_ = admin.Close()
		t.Fatal("PostgreSQL schema create failed")
	}
	t.Cleanup(func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if _, err := admin.SQL().ExecContext(cleanup, "DROP SCHEMA IF EXISTS "+schema+" CASCADE"); err != nil {
			t.Error("PostgreSQL schema cleanup failed")
		}
		_ = admin.Close()
	})
	parsed, err := url.Parse(dsn)
	if err != nil {
		t.Fatal("PostgreSQL test material invalid")
	}
	query := parsed.Query()
	query.Set("search_path", schema)
	parsed.RawQuery = query.Encode()
	db, err := database.NewProcess().Open(ctx, database.Config{Profile: config.ProfileServerPostgres, PostgresDSN: parsed.String()})
	if err != nil {
		t.Fatal("isolated PostgreSQL open failed")
	}
	defer func() { _ = db.Close() }()
	var current string
	if err := db.Bun().QueryRowContext(ctx, "SELECT current_schema()").Scan(&current); err != nil || current != schema {
		t.Fatalf("PostgreSQL schema isolation failed current=%q", current)
	}
	runner, err := migrations.NewRunner(settingsmigration.Provider{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := runner.Up(ctx, db); err != nil {
		t.Fatal(err)
	}
	service, err := newService(db, testAuthorizer{dialect: db.Dialect(), scope: ScopeAll})
	if err != nil {
		t.Fatal(err)
	}
	for _, fixture := range []SettingInput{{Category: CategoryBusiness, Key: "literal.percent", Label: "% literal", Value: "one"}, {Category: CategoryBusiness, Key: "literal.under", Label: "_ literal", Value: "two"}, {Category: CategoryBusiness, Key: "literal.ascii", Label: "<:@ collision", Value: "three"}, {Category: CategoryBusiness, Key: "literal.unicode", Label: "ä collision", Value: "four"}} {
		if _, err := service.CreateSetting(ctx, "actor", fixture); err != nil {
			t.Fatal(err)
		}
	}
	for search, want := range map[string]string{"%": "literal.percent", "_": "literal.under", "<:@": "literal.ascii", "ä": "literal.unicode"} {
		page, err := service.ListSettings(ctx, "actor", CategoryBusiness, ListQuery{Search: search, Page: 1, PerPage: 20})
		if err != nil || page.Total != 1 || page.Rows[0].Key != want {
			t.Fatalf("PostgreSQL literal search %q failed", search)
		}
	}
	page, err := service.ListSettings(ctx, "actor", CategoryBusiness, ListQuery{Search: "literal.", Page: 1, PerPage: 20})
	if err != nil {
		t.Fatal(err)
	}
	keys := []string{}
	for _, row := range page.Rows {
		keys = append(keys, row.Key)
	}
	if strings.Join(keys, ",") != "literal.percent,literal.ascii,literal.under,literal.unicode" {
		t.Fatalf("PostgreSQL deterministic order=%v", keys)
	}
}

func TestSettingsPostgresSearchPathIsURLSafe(t *testing.T) {
	parsed, err := url.Parse("postgresql://localhost/example?sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	query := parsed.Query()
	query.Set("search_path", "t10_settings_contract")
	parsed.RawQuery = query.Encode()
	if !strings.Contains(parsed.String(), "search_path=t10_settings_contract") || !strings.Contains(parsed.String(), "sslmode=disable") {
		t.Fatal("search_path was not preserved")
	}
}
