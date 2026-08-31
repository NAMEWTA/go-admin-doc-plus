package authorization_test

import (
	"context"
	"errors"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/app/adapters"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/iam/account"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/iam/administration"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/iam/authorization"
	sessionmigration "github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/iam/migrations/0010-session-schema"
	administrationmigration "github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/iam/migrations/0020-administration-schema"
	datascopemigration "github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/iam/migrations/0050-data-scope"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/organization"
	organizationmigration "github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/organization/migrations"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/platform/config"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/platform/database"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/platform/migrations"
)

const (
	scopeActor  = "account-scope-actor-001"
	scopeTarget = "account-scope-target-01"
	departmentA = "department-scope-a01"
	departmentB = "department-scope-b01"
	departmentC = "department-scope-c01"
	positionA   = "position-scope-a001"
	positionC   = "position-scope-c001"
)

func TestFiveDataScopesUnionOnlyRolesGrantingThePermission(t *testing.T) {
	db, resolver, _ := newScopeFixture(t)
	ctx := context.Background()
	setAccountScope(t, db, scopeActor, departmentA)
	seedScopeRole(t, db, "role-scope-all-001", authorization.ScopeAll, "scope.all.read", nil, true)
	seedScopeRole(t, db, "role-scope-self-01", authorization.ScopeSelf, "scope.self.read", nil, true)
	seedScopeRole(t, db, "role-scope-org-001", authorization.ScopeOrganization, "scope.org.read", nil, true)
	seedScopeRole(t, db, "role-scope-tree-01", authorization.ScopeOrganizationTree, "scope.tree.read", nil, true)
	seedScopeRole(t, db, "role-scope-custom1", authorization.ScopeCustom, "scope.custom.read", []string{departmentC}, true)
	seedScopeRole(t, db, "role-scope-rogue01", authorization.ScopeAll, "scope.unrelated.read", nil, true)

	checks := []struct {
		permission        string
		wantAll, wantSelf bool
		departments       []string
	}{
		{"scope.all.read", true, false, []string{}},
		{"scope.self.read", false, true, []string{}},
		{"scope.org.read", false, false, []string{departmentA}},
		{"scope.tree.read", false, false, []string{departmentA, departmentB}},
		{"scope.custom.read", false, false, []string{departmentC}},
	}
	for _, check := range checks {
		resolved, err := resolver.Resolve(ctx, scopeActor, check.permission)
		if err != nil || resolved.All != check.wantAll || resolved.Self != check.wantSelf || !reflect.DeepEqual(resolved.DepartmentIDs, check.departments) {
			t.Fatalf("%s = %#v, %v", check.permission, resolved, err)
		}
	}
	organizationScope, _ := resolver.Resolve(ctx, scopeActor, "scope.org.read")
	if !organizationScope.Allows(scopeActor, "another-account", departmentA) || organizationScope.Allows(scopeActor, "another-account", departmentC) {
		t.Fatal("department predicate expanded beyond the resolved organization")
	}
	selfScope, _ := resolver.Resolve(ctx, scopeActor, "scope.self.read")
	if !selfScope.Allows(scopeActor, scopeActor, "") || selfScope.Allows(scopeActor, scopeTarget, "") {
		t.Fatal("self predicate authorized another account")
	}
}

func TestEmptyScopesStayEmptyAndOrganizationChangesApplyNextRequest(t *testing.T) {
	db, resolver, _ := newScopeFixture(t)
	seedScopeRole(t, db, "role-empty-org-0001", authorization.ScopeOrganization, "scope.empty.org", nil, true)
	seedScopeRole(t, db, "role-empty-custom1", authorization.ScopeCustom, "scope.empty.custom", nil, true)
	for _, permission := range []string{"scope.empty.org", "scope.empty.custom"} {
		resolved, err := resolver.Resolve(context.Background(), scopeActor, permission)
		if err != nil || resolved.All || resolved.Self || len(resolved.DepartmentIDs) != 0 {
			t.Fatalf("empty scope %s expanded = %#v, %v", permission, resolved, err)
		}
	}
	setAccountScope(t, db, scopeActor, departmentA)
	seedScopeRole(t, db, "role-moving-tree01", authorization.ScopeOrganizationTree, "scope.moving.tree", nil, true)
	before, err := resolver.Resolve(context.Background(), scopeActor, "scope.moving.tree")
	if err != nil || !reflect.DeepEqual(before.DepartmentIDs, []string{departmentA, departmentB}) {
		t.Fatalf("tree before move = %#v, %v", before, err)
	}
	mustScopeSQL(t, db, `UPDATE organization_departments SET parent_id = ? WHERE id = ?`, departmentC, departmentB)
	after, err := resolver.Resolve(context.Background(), scopeActor, "scope.moving.tree")
	if err != nil || !reflect.DeepEqual(after.DepartmentIDs, []string{departmentA}) {
		t.Fatalf("tree after move = %#v, %v", after, err)
	}
	mustScopeSQL(t, db, `UPDATE iam_roles SET enabled = ? WHERE id = ?`, false, "role-moving-tree01")
	if _, err := resolver.Resolve(context.Background(), scopeActor, "scope.moving.tree"); !errors.Is(err, authorization.ErrDenied) {
		t.Fatalf("disabled role remained effective: %v", err)
	}
}

func TestOrganizationAssignmentsAndReferencesFailClosed(t *testing.T) {
	db, resolver, projection := newScopeFixture(t)
	ctx := context.Background()
	admin, err := administration.NewDataScopeService(db, projection, administration.WithDataScopeClock(func() time.Time { return time.Date(2026, 8, 31, 8, 0, 0, 0, time.UTC) }))
	if err != nil {
		t.Fatal(err)
	}
	primary := departmentC
	if err := admin.SetAccountOrganization(ctx, scopeActor, scopeTarget, administration.AccountOrganization{PrimaryDepartmentID: &primary, PositionIDs: []string{positionA}}); !errors.Is(err, administration.ErrConflict) {
		t.Fatalf("cross-department position = %v", err)
	}
	if err := admin.SetAccountOrganization(ctx, scopeActor, scopeTarget, administration.AccountOrganization{PrimaryDepartmentID: &primary, PositionIDs: []string{positionC}}); err != nil {
		t.Fatal(err)
	}
	if err := admin.SetRoleDataScope(ctx, scopeActor, "role-editable-0001", administration.RoleDataScope{Scope: authorization.ScopeCustom, DepartmentIDs: []string{"missing-department"}}); !errors.Is(err, administration.ErrConflict) {
		t.Fatalf("invalid custom department = %v", err)
	}
	if err := admin.SetRoleDataScope(ctx, scopeActor, "role-editable-0001", administration.RoleDataScope{Scope: authorization.ScopeCustom, DepartmentIDs: []string{departmentB}}); err != nil {
		t.Fatal(err)
	}
	resolved, err := resolver.Resolve(ctx, scopeTarget, "scope.editable.read")
	if err != nil || !reflect.DeepEqual(resolved.DepartmentIDs, []string{departmentB}) {
		t.Fatalf("administrative custom scope = %#v, %v", resolved, err)
	}

	authorizationAdapters, err := adapters.NewAuthorization(db)
	if err != nil {
		t.Fatal(err)
	}
	organizationService, err := organization.NewService(db, authorizationAdapters.Organization())
	if err != nil {
		t.Fatal(err)
	}
	if err := organizationService.DeletePosition(ctx, scopeActor, positionC); !errors.Is(err, organization.ErrConflict) {
		t.Fatalf("assigned position deletion = %v", err)
	}
	if err := organizationService.DeleteDepartment(ctx, scopeActor, departmentB); !errors.Is(err, organization.ErrConflict) {
		t.Fatalf("custom department deletion = %v", err)
	}
}

func newScopeFixture(t *testing.T) (*database.Database, *authorization.ScopeResolver, *organization.ProjectionAdapter) {
	t.Helper()
	ctx := context.Background()
	db, err := database.NewProcess().Open(ctx, database.Config{Profile: config.ProfileServerSQLite, SQLitePath: filepath.Join(t.TempDir(), "scope.sqlite3")})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	runner, err := migrations.NewRunner(sessionmigration.Provider{}, administrationmigration.Provider{}, datascopemigration.Provider{}, organizationmigration.Provider{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := runner.Up(ctx, db); err != nil {
		t.Fatal(err)
	}
	hash, err := account.HashPassword("scope fixture password")
	if err != nil {
		t.Fatal(err)
	}
	repository := account.NewRepository(db.Dialect())
	if err := db.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error {
		for _, value := range []account.Credential{
			{Profile: account.Profile{ID: scopeActor, Username: "scope-actor", DisplayName: "Scope Actor", Email: "scope-actor@example.test"}, PasswordHash: hash},
			{Profile: account.Profile{ID: scopeTarget, Username: "scope-target", DisplayName: "Scope Target", Email: "scope-target@example.test"}, PasswordHash: hash},
		} {
			if err := repository.Create(ctx, tx, value, time.Now().UTC()); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	mustScopeSQL(t, db, `INSERT INTO iam_account_roles(account_id, role_id) VALUES (?, ?)`, scopeActor, "role-system-admin")
	for _, statement := range []struct {
		query string
		args  []any
	}{
		{`INSERT INTO organization_departments(id, department_key, name, parent_id, sort_order, protected, created_at, updated_at) VALUES (?, ?, ?, ?, 0, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`, []any{departmentA, "scope-a", "Scope A", "department-root-001", false}},
		{`INSERT INTO organization_departments(id, department_key, name, parent_id, sort_order, protected, created_at, updated_at) VALUES (?, ?, ?, ?, 0, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`, []any{departmentB, "scope-b", "Scope B", departmentA, false}},
		{`INSERT INTO organization_departments(id, department_key, name, parent_id, sort_order, protected, created_at, updated_at) VALUES (?, ?, ?, ?, 0, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`, []any{departmentC, "scope-c", "Scope C", "department-root-001", false}},
		{`INSERT INTO organization_positions(id, position_key, name, name_key, department_id, enabled, protected, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`, []any{positionA, "scope-position-a", "Scope Position A", "53.63.6f.70.65.20.50.6f.73.69.74.69.6f.6e.20.41.", departmentA, true, false}},
		{`INSERT INTO organization_positions(id, position_key, name, name_key, department_id, enabled, protected, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`, []any{positionC, "scope-position-c", "Scope Position C", "53.63.6f.70.65.20.50.6f.73.69.74.69.6f.6e.20.43.", departmentC, true, false}},
	} {
		mustScopeSQL(t, db, statement.query, statement.args...)
	}
	registry, err := authorization.NewCapabilityRegistry(db)
	if err != nil || organization.RegisterCapabilities(ctx, registry) != nil {
		t.Fatal("organization capability registration failed")
	}
	mustScopeSQL(t, db, `INSERT INTO iam_permissions(code, name, protected) VALUES (?, ?, ?)`, "scope.editable.read", "Read editable scope", true)
	mustScopeSQL(t, db, `INSERT INTO iam_roles(id, role_key, name, data_scope, enabled, protected, created_at, updated_at) VALUES (?, ?, ?, 'self', ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`, "role-editable-0001", "scope-editable", "Scope Editable", true, false)
	mustScopeSQL(t, db, `INSERT INTO iam_role_data_scopes(role_id, scope, updated_at) VALUES (?, 'self', CURRENT_TIMESTAMP)`, "role-editable-0001")
	mustScopeSQL(t, db, `INSERT INTO iam_role_permissions(role_id, permission_code) VALUES (?, ?)`, "role-editable-0001", "scope.editable.read")
	mustScopeSQL(t, db, `INSERT INTO iam_account_roles(account_id, role_id) VALUES (?, ?)`, scopeTarget, "role-editable-0001")
	projection, err := organization.NewProjectionAdapter(db)
	if err != nil {
		t.Fatal(err)
	}
	resolver, err := authorization.NewScopeResolver(db, projection)
	if err != nil {
		t.Fatal(err)
	}
	return db, resolver, projection
}

func seedScopeRole(t *testing.T, db *database.Database, roleID string, scope authorization.Scope, permission string, departments []string, enabled bool) {
	t.Helper()
	roleKey := roleID[5:]
	mustScopeSQL(t, db, `INSERT INTO iam_permissions(code, name, protected) VALUES (?, ?, ?) ON CONFLICT(code) DO NOTHING`, permission, permission, true)
	mustScopeSQL(t, db, `INSERT INTO iam_roles(id, role_key, name, data_scope, enabled, protected, created_at, updated_at) VALUES (?, ?, ?, 'self', ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`, roleID, roleKey, roleKey, enabled, false)
	mustScopeSQL(t, db, `INSERT INTO iam_role_data_scopes(role_id, scope, updated_at) VALUES (?, ?, CURRENT_TIMESTAMP)`, roleID, scope)
	mustScopeSQL(t, db, `INSERT INTO iam_role_permissions(role_id, permission_code) VALUES (?, ?)`, roleID, permission)
	mustScopeSQL(t, db, `INSERT INTO iam_account_roles(account_id, role_id) VALUES (?, ?)`, scopeActor, roleID)
	for _, departmentID := range departments {
		mustScopeSQL(t, db, `INSERT INTO iam_role_data_scope_departments(role_id, department_id) VALUES (?, ?)`, roleID, departmentID)
	}
}

func setAccountScope(t *testing.T, db *database.Database, accountID, departmentID string) {
	t.Helper()
	mustScopeSQL(t, db, `INSERT INTO iam_account_organization(account_id, primary_department_id, updated_at) VALUES (?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(account_id) DO UPDATE SET primary_department_id = excluded.primary_department_id, updated_at = excluded.updated_at`, accountID, departmentID)
}

func mustScopeSQL(t *testing.T, db *database.Database, query string, args ...any) {
	t.Helper()
	if _, err := db.Bun().ExecContext(context.Background(), query, args...); err != nil {
		t.Fatal(err)
	}
}
