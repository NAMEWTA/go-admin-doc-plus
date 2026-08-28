package organization_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/iam/account"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/iam/authorization"
	sessionmigration "github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/iam/migrations/0010-session-schema"
	administrationmigration "github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/iam/migrations/0020-administration-schema"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/organization"
	organizationmigration "github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/organization/migrations"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/platform/config"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/platform/database"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/platform/migrations"
)

func TestOrganizationSQLiteDialectContract(t *testing.T) {
	db, cleanup := openOrganizationContractDatabase(t, "sqlite")
	defer cleanup()
	runOrganizationDialectContract(t, db)
}

func TestOrganizationPostgresDialectContract(t *testing.T) {
	if os.Getenv(postgresDSNEnvironment) == "" {
		t.Skip(postgresDSNEnvironment + " is not configured")
	}
	db, cleanup := openOrganizationContractDatabase(t, "postgres")
	defer cleanup()
	runOrganizationDialectContract(t, db)
}

func runOrganizationDialectContract(t *testing.T, db *database.Database) {
	t.Helper()
	ctx := context.Background()
	runner, err := migrations.NewRunner(sessionmigration.Provider{}, administrationmigration.Provider{}, organizationmigration.Provider{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := runner.Up(ctx, db); err != nil {
		t.Fatal(err)
	}
	hash, err := account.HashPassword("organization contract password")
	if err != nil {
		t.Fatal(err)
	}
	const actorID = "account-organization-contract-001"
	if err := db.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error {
		return account.NewRepository(db.Dialect()).Create(ctx, tx, account.Credential{Profile: account.Profile{ID: actorID, Username: "organization-contract", DisplayName: "Organization Contract", Email: "organization-contract@example.test"}, PasswordHash: hash}, time.Now().UTC())
	}); err != nil {
		t.Fatal(err)
	}
	registry, err := authorization.NewCapabilityRegistry(db)
	if err != nil || organization.RegisterCapabilities(ctx, registry) != nil {
		t.Fatal("organization registry failed")
	}
	if _, err := db.Bun().ExecContext(ctx, `INSERT INTO iam_account_roles(account_id, role_id) VALUES (?, ?)`, actorID, "role-system-admin"); err != nil {
		t.Fatal(err)
	}
	manifest, err := authorization.NewService(db).Manifest(ctx, actorID)
	if err != nil || manifest.Scope != authorization.ScopeAll || !contains(manifest.Permissions, organization.PermissionDepartmentsRead) || !contains(manifest.Permissions, organization.PermissionPositionsWrite) {
		t.Fatalf("organization manifest=%#v err=%v", manifest, err)
	}
	service, err := organization.NewService(db)
	if err != nil {
		t.Fatal(err)
	}
	department, err := service.CreateDepartment(ctx, actorID, organization.DepartmentInput{Key: "unicode-department", Name: strings.Repeat("😀", 100), ParentID: "department-root-001"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.CreateDepartment(ctx, actorID, organization.DepartmentInput{Key: "too-long-department", Name: strings.Repeat("😀", 101), ParentID: "department-root-001"}); !errors.Is(err, organization.ErrValidation) {
		t.Fatalf("101-code-point department name=%v", err)
	}
	fixtures := []organization.PositionInput{
		{Key: "position-ascii", Name: "<:@ collision", DepartmentID: department.ID, Enabled: true},
		{Key: "position-percent", Name: "% literal", DepartmentID: department.ID, Enabled: true},
		{Key: "position-under", Name: "_ literal", DepartmentID: department.ID, Enabled: true},
		{Key: "position-unicode", Name: "ä collision", DepartmentID: department.ID, Enabled: true},
	}
	for _, fixture := range fixtures {
		if _, err := service.CreatePosition(ctx, actorID, fixture); err != nil {
			t.Fatalf("create %q=%v", fixture.Name, err)
		}
	}
	for search, key := range map[string]string{"<:@": "position-ascii", "%": "position-percent", "_": "position-under", "ä": "position-unicode"} {
		page, err := service.ListPositions(ctx, actorID, search, 1, 20)
		if err != nil || page.Total != 1 || len(page.Rows) != 1 || page.Rows[0].Key != key {
			t.Fatalf("literal search %q=%#v err=%v", search, page, err)
		}
	}
	page, err := service.ListPositions(ctx, actorID, "position-", 1, 20)
	want := []string{"position-ascii", "position-percent", "position-under", "position-unicode"}
	if err != nil || len(page.Rows) != len(want) {
		t.Fatalf("C ordering=%#v err=%v", page, err)
	}
	for index, key := range want {
		if page.Rows[index].Key != key {
			t.Fatalf("C ordering[%d]=%q want=%q", index, page.Rows[index].Key, key)
		}
	}
	if _, err := service.ListPositions(ctx, actorID, strings.Repeat("界", 101), 1, 20); !errors.Is(err, organization.ErrValidation) {
		t.Fatalf("101-code-point search=%v", err)
	}
	if _, err := db.Bun().ExecContext(ctx, `UPDATE iam_roles SET data_scope = ? WHERE id = ?`, "self", "role-system-admin"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.ListDepartments(ctx, actorID); !errors.Is(err, organization.ErrDenied) {
		t.Fatalf("self scope list=%v", err)
	}
}

func openOrganizationContractDatabase(t *testing.T, profile string) (*database.Database, func()) {
	t.Helper()
	ctx := context.Background()
	if profile == "sqlite" {
		db, err := database.NewProcess().Open(ctx, database.Config{Profile: config.ProfileServerSQLite, SQLitePath: filepath.Join(t.TempDir(), "organization-contract.sqlite3")})
		if err != nil {
			t.Fatal(err)
		}
		return db, func() { _ = db.Close() }
	}
	dsn := os.Getenv(postgresDSNEnvironment)
	admin, err := database.NewProcess().Open(ctx, database.Config{Profile: config.ProfileServerPostgres, PostgresDSN: dsn})
	if err != nil {
		t.Fatal("PostgreSQL contract administrator failed")
	}
	schema := fmt.Sprintf("t09_contract_%d", time.Now().UnixNano())
	if _, err := admin.SQL().ExecContext(ctx, `CREATE SCHEMA `+schema); err != nil {
		t.Fatal("PostgreSQL contract schema failed")
	}
	var db *database.Database
	var cleanupOnce sync.Once
	cleanup := func() {
		cleanupOnce.Do(func() {
			if db != nil {
				_ = db.Close()
			}
			cleanupContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if _, err := admin.SQL().ExecContext(cleanupContext, `DROP SCHEMA IF EXISTS `+schema+` CASCADE`); err != nil {
				t.Error("PostgreSQL contract cleanup failed")
			} else {
				var exists bool
				if err := admin.Bun().QueryRowContext(cleanupContext, `SELECT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = ?)`, schema).Scan(&exists); err != nil || exists {
					t.Error("PostgreSQL contract schema residue detected")
				}
			}
			_ = admin.Close()
		})
	}
	t.Cleanup(cleanup)
	db, err = database.NewProcess().Open(ctx, database.Config{Profile: config.ProfileServerPostgres, PostgresDSN: organizationPostgresDSN(t, dsn, schema)})
	if err != nil {
		t.Fatal("isolated PostgreSQL contract open failed")
	}
	var currentSchema string
	if err := db.Bun().QueryRowContext(ctx, `SELECT current_schema()`).Scan(&currentSchema); err != nil || currentSchema != schema {
		t.Fatalf("isolated PostgreSQL contract schema=%q err=%v", currentSchema, err)
	}
	return db, cleanup
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
