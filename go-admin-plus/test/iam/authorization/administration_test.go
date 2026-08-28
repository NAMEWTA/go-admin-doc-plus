package authorization_test

import (
	"context"
	"errors"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/iam/account"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/iam/administration"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/iam/authorization"
	sessionmigration "github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/iam/migrations/0010-session-schema"
	administrationmigration "github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/iam/migrations/0020-administration-schema"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/platform/config"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/platform/database"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/platform/migrations"
)

func TestAdministrationConstructorOwnsAuthorizationDatabase(t *testing.T) {
	constructor := reflect.TypeOf(administration.NewService)
	options := reflect.TypeOf([]administration.Option{})
	if !constructor.IsVariadic() || constructor.NumIn() != 2 || constructor.In(1) != options {
		t.Fatalf("constructor permits a split authorization owner: %v", constructor)
	}
}

func TestAdministrationClosesRolePermissionAndDataScopeLoop(t *testing.T) {
	_, service := newAdministrationFixture(t)
	ctx := context.Background()
	created, err := service.CreateUser(ctx, adminID, administration.CreateUser{Username: "reader", DisplayName: "Reader", Email: "reader@example.test", Password: "reader password value"})
	if err != nil {
		t.Fatal(err)
	}
	role, err := service.CreateRole(ctx, adminID, "reader", "Reader", authorization.ScopeSelf)
	if err != nil {
		t.Fatal(err)
	}
	if err := service.SetRoleGrants(ctx, adminID, role.ID, []string{authorization.PermissionUsersRead, authorization.PermissionManifestRead}, nil); err != nil {
		t.Fatal(err)
	}
	if err := service.SetUserRoles(ctx, adminID, created.ID, []string{role.ID}); err != nil {
		t.Fatal(err)
	}

	page, err := service.ListUsers(ctx, created.ID, "", 1, 20)
	if err != nil || page.Total != 1 || len(page.Rows) != 1 || page.Rows[0].ID != created.ID {
		t.Fatalf("self-scoped list = %#v, %v", page, err)
	}
	if _, err := service.GetUser(ctx, created.ID, adminID); !errors.Is(err, administration.ErrDenied) {
		t.Fatalf("self-scoped detail escaped: %v", err)
	}
	if _, err := service.CreateUser(ctx, created.ID, administration.CreateUser{Username: "forbidden", DisplayName: "Forbidden", Email: "forbidden@example.test", Password: "forbidden password"}); !errors.Is(err, administration.ErrDenied) {
		t.Fatalf("self-scoped write escaped: %v", err)
	}

	role.Enabled = false
	if err := service.UpdateRole(ctx, adminID, role); err != nil {
		t.Fatal(err)
	}
	if _, err := service.ListUsers(ctx, created.ID, "", 1, 20); !errors.Is(err, administration.ErrDenied) {
		t.Fatalf("disabled role remained effective: %v", err)
	}
}

func TestProtectedReferencesAndDuplicateCommandsAreAtomic(t *testing.T) {
	db, service := newAdministrationFixture(t)
	ctx := context.Background()
	if err := service.DeleteRole(ctx, adminID, "role-system-admin"); !errors.Is(err, administration.ErrConflict) {
		t.Fatalf("protected role deletion = %v", err)
	}
	if err := service.DeleteMenu(ctx, adminID, "menu-iam-users-01"); !errors.Is(err, administration.ErrConflict) {
		t.Fatalf("protected menu deletion = %v", err)
	}
	if err := service.DeleteUser(ctx, adminID, adminID); !errors.Is(err, administration.ErrConflict) {
		t.Fatalf("protected user deletion = %v", err)
	}

	role, err := service.CreateRole(ctx, adminID, "operator", "Operator", authorization.ScopeAll)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.CreateRole(ctx, adminID, "operator", "Duplicate", authorization.ScopeAll); !errors.Is(err, administration.ErrConflict) {
		t.Fatalf("duplicate role = %v", err)
	}
	if err := service.SetRoleGrants(ctx, adminID, role.ID, []string{authorization.PermissionUsersRead, authorization.PermissionUsersRead}, nil); !errors.Is(err, administration.ErrConflict) {
		t.Fatalf("duplicate grant = %v", err)
	}
	var grantCount int
	if err := db.Bun().QueryRowContext(ctx, `SELECT COUNT(*) FROM iam_role_permissions WHERE role_id = ?`, role.ID).Scan(&grantCount); err != nil {
		t.Fatal(err)
	}
	if grantCount != 0 {
		t.Fatal("rejected duplicate command partially changed grants")
	}

	menu, err := service.CreateMenu(ctx, adminID, administration.Menu{Key: "operator", Label: "Operator", Path: "/iam/operator", PermissionCode: authorization.PermissionUsersRead, SortOrder: 50})
	if err != nil {
		t.Fatal(err)
	}
	if err := service.SetRoleGrants(ctx, adminID, role.ID, []string{authorization.PermissionUsersRead}, []string{menu.ID}); err != nil {
		t.Fatal(err)
	}
	if err := service.DeleteMenu(ctx, adminID, menu.ID); !errors.Is(err, administration.ErrConflict) {
		t.Fatalf("referenced menu deletion = %v", err)
	}
	target, err := service.CreateUser(ctx, adminID, administration.CreateUser{Username: "role-target", DisplayName: "Role Target", Email: "role-target@example.test", Password: "role target password"})
	if err != nil {
		t.Fatal(err)
	}
	if err := service.SetUserRoles(ctx, adminID, target.ID, []string{role.ID}); err != nil {
		t.Fatal(err)
	}
	if err := service.DeleteRole(ctx, adminID, role.ID); !errors.Is(err, administration.ErrConflict) {
		t.Fatalf("referenced role deletion = %v", err)
	}
}

func TestProtectedAdministratorRolesCannotBeCleared(t *testing.T) {
	db, service := newAdministrationFixture(t)
	ctx := context.Background()

	if err := service.SetUserRoles(ctx, adminID, adminID, nil); !errors.Is(err, administration.ErrConflict) {
		t.Fatalf("protected role assignment = %v", err)
	}
	var count int
	if err := db.Bun().QueryRowContext(ctx, `SELECT COUNT(*) FROM iam_account_roles WHERE account_id = ? AND role_id = ?`, adminID, "role-system-admin").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("protected role assignment changed join rows: %d", count)
	}
}

func TestMenuUpdateNormalizesAndValidatesBeforeMutation(t *testing.T) {
	db, service := newAdministrationFixture(t)
	ctx := context.Background()
	menu, err := service.CreateMenu(ctx, adminID, administration.Menu{
		Key: "reports", Label: "Reports", Path: "/iam/reports",
		PermissionCode: authorization.PermissionUsersRead, SortOrder: 20,
	})
	if err != nil {
		t.Fatal(err)
	}

	menu.Key = "reports-updated"
	menu.Label = "  Updated Reports  "
	menu.Path = "  /iam/reports-updated  "
	menu.PermissionCode = "  " + authorization.PermissionUsersRead + "  "
	menu.SortOrder = 21
	if err := service.UpdateMenu(ctx, adminID, menu); err != nil {
		t.Fatal(err)
	}
	var key, label, path, permission string
	var sortOrder int
	if err := db.Bun().QueryRowContext(ctx, `SELECT menu_key, label, path, permission_code, sort_order FROM iam_menus WHERE id = ?`, menu.ID).Scan(&key, &label, &path, &permission, &sortOrder); err != nil {
		t.Fatal(err)
	}
	if key != "reports-updated" || label != "Updated Reports" || path != "/iam/reports-updated" || permission != authorization.PermissionUsersRead || sortOrder != 21 {
		t.Fatalf("normalized menu = %q %q %q %q %d", key, label, path, permission, sortOrder)
	}

	for _, invalid := range []administration.Menu{
		{ID: menu.ID, Key: "Reports Updated", Label: "Updated Reports", Path: "/iam/reports-updated", PermissionCode: authorization.PermissionUsersRead, SortOrder: 21},
		{ID: menu.ID, Key: "reports-updated", Label: "Updated Reports", Path: "IAM/UPPER", PermissionCode: authorization.PermissionUsersRead, SortOrder: 21},
		{ID: menu.ID, Key: "reports-updated", Label: "Updated Reports", Path: "/iam/reports-updated", PermissionCode: authorization.PermissionUsersRead, SortOrder: -1},
		{ID: menu.ID, Key: "reports-updated", Label: "Updated Reports", Path: "/iam/reports-updated", PermissionCode: authorization.PermissionUsersRead, SortOrder: 100001},
	} {
		if err := service.UpdateMenu(ctx, adminID, invalid); !errors.Is(err, administration.ErrValidation) {
			t.Fatalf("invalid menu update = %v", err)
		}
	}
	if err := db.Bun().QueryRowContext(ctx, `SELECT menu_key, path, sort_order FROM iam_menus WHERE id = ?`, menu.ID).Scan(&key, &path, &sortOrder); err != nil {
		t.Fatal(err)
	}
	if key != "reports-updated" || path != "/iam/reports-updated" || sortOrder != 21 {
		t.Fatalf("invalid update changed menu = %q %q %d", key, path, sortOrder)
	}
}

func TestStableKeysBoundedPaginationAndMigrationConstraints(t *testing.T) {
	db, service := newAdministrationFixture(t)
	ctx := context.Background()
	if _, err := service.ListUsers(ctx, adminID, "", 1_000_001, 20); !errors.Is(err, administration.ErrValidation) {
		t.Fatalf("unbounded page = %v", err)
	}
	for _, key := range []string{"Bad Key", "-leading", "upper" + strings.ToUpper("case")} {
		if _, err := service.CreateRole(ctx, adminID, key, "Invalid role", authorization.ScopeAll); !errors.Is(err, administration.ErrValidation) {
			t.Fatalf("invalid role key %q = %v", key, err)
		}
		if _, err := service.CreateMenu(ctx, adminID, administration.Menu{Key: key, Label: "Invalid menu", Path: "/iam/invalid", PermissionCode: authorization.PermissionUsersRead}); !errors.Is(err, administration.ErrValidation) {
			t.Fatalf("invalid menu key %q = %v", key, err)
		}
	}
	if _, err := db.Bun().ExecContext(ctx, `INSERT INTO iam_roles(id, role_key, name, data_scope, enabled, protected, created_at, updated_at) VALUES (?, ?, ?, 'all', 1, 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`, "role-invalid-key1", "invalid key", "Invalid"); err == nil {
		t.Fatal("SQLite role key constraint accepted invalid key")
	}
	if _, err := db.Bun().ExecContext(ctx, `INSERT INTO iam_menus(id, menu_key, label, path, permission_code, sort_order, protected, created_at, updated_at) VALUES (?, ?, ?, ?, NULL, 1, 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`, "menu-invalid-null", "valid-menu", "Invalid", "/iam/invalid-null"); err == nil {
		t.Fatal("SQLite menu permission accepted NULL")
	}
	var permissionNotNull int
	if err := db.Bun().QueryRowContext(ctx, `SELECT "notnull" FROM pragma_table_info('iam_menus') WHERE name = 'permission_code'`).Scan(&permissionNotNull); err != nil || permissionNotNull != 1 {
		t.Fatalf("SQLite menu permission nullability = %d, %v", permissionNotNull, err)
	}
}

func TestSelfScopeCannotReadGlobalAdministrationMetadata(t *testing.T) {
	_, service := newAdministrationFixture(t)
	ctx := context.Background()
	user, err := service.CreateUser(ctx, adminID, administration.CreateUser{Username: "metadata-self", DisplayName: "Metadata Self", Email: "metadata-self@example.test", Password: "metadata self password"})
	if err != nil {
		t.Fatal(err)
	}
	role, err := service.CreateRole(ctx, adminID, "metadata-self", "Metadata self", authorization.ScopeSelf)
	if err != nil {
		t.Fatal(err)
	}
	if err := service.SetRoleGrants(ctx, adminID, role.ID, []string{authorization.PermissionRolesRead, authorization.PermissionMenusRead, authorization.PermissionPermissionsRead}, nil); err != nil {
		t.Fatal(err)
	}
	if err := service.SetUserRoles(ctx, adminID, user.ID, []string{role.ID}); err != nil {
		t.Fatal(err)
	}
	if _, err := service.ListRoles(ctx, user.ID); !errors.Is(err, administration.ErrDenied) {
		t.Fatalf("self role list = %v", err)
	}
	if _, err := service.ListMenus(ctx, user.ID); !errors.Is(err, administration.ErrDenied) {
		t.Fatalf("self menu list = %v", err)
	}
	if _, err := service.ListPermissions(ctx, user.ID); !errors.Is(err, administration.ErrDenied) {
		t.Fatalf("self permission list = %v", err)
	}
}

func TestBatchDeleteIsAtomicForProtectedMissingAndSuccessfulSets(t *testing.T) {
	db, service := newAdministrationFixture(t)
	ctx := context.Background()
	first, err := service.CreateUser(ctx, adminID, administration.CreateUser{Username: "batch-one", DisplayName: "Batch One", Email: "batch-one@example.test", Password: "batch one password"})
	if err != nil {
		t.Fatal(err)
	}
	second, err := service.CreateUser(ctx, adminID, administration.CreateUser{Username: "batch-two", DisplayName: "Batch Two", Email: "batch-two@example.test", Password: "batch two password"})
	if err != nil {
		t.Fatal(err)
	}
	if err := service.DeleteUsers(ctx, adminID, []string{first.ID, adminID}); !errors.Is(err, administration.ErrConflict) {
		t.Fatalf("protected batch = %v", err)
	}
	var count int
	if err := db.Bun().QueryRowContext(ctx, `SELECT COUNT(*) FROM iam_accounts WHERE id = ?`, first.ID).Scan(&count); err != nil || count != 1 {
		t.Fatal("protected batch partially deleted ordinary user")
	}
	if err := service.DeleteUsers(ctx, adminID, []string{first.ID, "missing-account-01"}); !errors.Is(err, administration.ErrNotFound) {
		t.Fatalf("missing batch = %v", err)
	}
	if err := service.DeleteUsers(ctx, adminID, []string{first.ID, second.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.Bun().QueryRowContext(ctx, `SELECT COUNT(*) FROM iam_accounts WHERE id IN (?, ?)`, first.ID, second.ID).Scan(&count); err != nil || count != 0 {
		t.Fatalf("successful batch remaining=%d err=%v", count, err)
	}
}

func TestPasswordResetHashesAndRevokesWithoutReturningSensitiveMaterial(t *testing.T) {
	db, service := newAdministrationFixture(t)
	ctx := context.Background()
	target, err := service.CreateUser(ctx, adminID, administration.CreateUser{Username: "target", DisplayName: "Target", Email: "target@example.test", Password: "original password value"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Bun().ExecContext(ctx, `INSERT INTO iam_sessions(id, account_id, token_hash, generation, csrf_hash, state, created_at, last_seen_at, idle_expires_at, absolute_expires_at, rotate_at) VALUES (?, ?, ?, 0, ?, 'active', ?, ?, ?, ?, ?)`, strings.Repeat("s", 43), target.ID, strings.Repeat("a", 64), strings.Repeat("b", 64), time.Now().UTC(), time.Now().UTC(), time.Now().Add(time.Hour).UTC(), time.Now().Add(2*time.Hour).UTC(), time.Now().Add(time.Hour).UTC()); err != nil {
		t.Fatal(err)
	}

	const replacement = "replacement password value"
	if err := service.ResetPassword(ctx, adminID, target.ID, replacement); err != nil {
		t.Fatal(err)
	}
	var hash, state string
	var generation int64
	if err := db.Bun().QueryRowContext(ctx, `SELECT password_hash, session_generation FROM iam_accounts WHERE id = ?`, target.ID).Scan(&hash, &generation); err != nil {
		t.Fatal(err)
	}
	if err := db.Bun().QueryRowContext(ctx, `SELECT state FROM iam_sessions WHERE account_id = ?`, target.ID).Scan(&state); err != nil {
		t.Fatal(err)
	}
	if hash == replacement || !strings.HasPrefix(hash, "$argon2id$") || !account.VerifyPassword(hash, replacement) || generation != 1 || state != "revoked" {
		t.Fatalf("reset state invalid: hashPolicy=%t generation=%d state=%q", strings.HasPrefix(hash, "$argon2id$"), generation, state)
	}
}

func TestDisableAdvancesGenerationAndRevokesWhileSelfScopeCannotChangeStatusOrReset(t *testing.T) {
	db, service := newAdministrationFixture(t)
	ctx := context.Background()
	target, err := service.CreateUser(ctx, adminID, administration.CreateUser{Username: "limited", DisplayName: "Limited", Email: "limited@example.test", Password: "limited password value"})
	if err != nil {
		t.Fatal(err)
	}
	role, err := service.CreateRole(ctx, adminID, "limited", "Limited", authorization.ScopeSelf)
	if err != nil {
		t.Fatal(err)
	}
	if err := service.SetRoleGrants(ctx, adminID, role.ID, []string{authorization.PermissionUsersRead, authorization.PermissionUsersWrite, authorization.PermissionUsersResetPassword}, nil); err != nil {
		t.Fatal(err)
	}
	if err := service.SetUserRoles(ctx, adminID, target.ID, []string{role.ID}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Bun().ExecContext(ctx, `INSERT INTO iam_sessions(id, account_id, token_hash, generation, csrf_hash, state, created_at, last_seen_at, idle_expires_at, absolute_expires_at, rotate_at) VALUES (?, ?, ?, 0, ?, 'active', ?, ?, ?, ?, ?)`, strings.Repeat("t", 43), target.ID, strings.Repeat("c", 64), strings.Repeat("d", 64), time.Now().UTC(), time.Now().UTC(), time.Now().Add(time.Hour).UTC(), time.Now().Add(2*time.Hour).UTC(), time.Now().Add(time.Hour).UTC()); err != nil {
		t.Fatal(err)
	}

	if _, err := service.UpdateUser(ctx, target.ID, target.ID, "Limited", "limited@example.test", false); !errors.Is(err, administration.ErrDenied) {
		t.Fatalf("self scope disabled account: %v", err)
	}
	if err := service.ResetPassword(ctx, target.ID, target.ID, "self reset password"); !errors.Is(err, administration.ErrDenied) {
		t.Fatalf("self scope reset password: %v", err)
	}
	if _, err := service.UpdateUser(ctx, adminID, target.ID, "Limited", "limited@example.test", false); err != nil {
		t.Fatal(err)
	}
	var generation int64
	var state string
	if err := db.Bun().QueryRowContext(ctx, `SELECT session_generation FROM iam_accounts WHERE id = ?`, target.ID).Scan(&generation); err != nil {
		t.Fatal(err)
	}
	if err := db.Bun().QueryRowContext(ctx, `SELECT state FROM iam_sessions WHERE account_id = ?`, target.ID).Scan(&state); err != nil {
		t.Fatal(err)
	}
	if generation != 1 || state != "revoked" {
		t.Fatalf("disable fence = generation %d, session %q", generation, state)
	}
}

func TestMigrationHasNoTenantCasbinOrJWTState(t *testing.T) {
	db, _ := newAdministrationFixture(t)
	for _, forbidden := range []string{"tenant", "casbin", "jwt"} {
		var count int
		if err := db.Bun().QueryRowContext(context.Background(), `SELECT COUNT(*) FROM sqlite_master WHERE lower(name) LIKE ?`, "%"+forbidden+"%").Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count != 0 {
			t.Fatalf("forbidden schema %q exists", forbidden)
		}
	}
}

const adminID = "account-admin-001"

func newAdministrationFixture(t *testing.T) (*database.Database, *administration.Service) {
	t.Helper()
	db, err := database.NewProcess().Open(context.Background(), database.Config{Profile: config.ProfileServerSQLite, SQLitePath: filepath.Join(t.TempDir(), "iam.db")})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	runner, err := migrations.NewRunner(sessionmigration.Provider{}, administrationmigration.Provider{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := runner.Up(context.Background(), db); err != nil {
		t.Fatal(err)
	}
	hash, err := account.HashPassword("administrator password")
	if err != nil {
		t.Fatal(err)
	}
	repository := account.NewRepository(db.Dialect())
	if err := db.WithinTx(context.Background(), func(ctx context.Context, tx database.Tx) error {
		return repository.Create(ctx, tx, account.Credential{Profile: account.Profile{ID: adminID, Username: "admin", DisplayName: "Administrator", Email: "admin@example.test"}, PasswordHash: hash}, time.Now().UTC())
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Bun().ExecContext(context.Background(), `INSERT INTO iam_account_roles(account_id, role_id) VALUES (?, ?)`, adminID, "role-system-admin"); err != nil {
		t.Fatal(err)
	}
	service, err := administration.NewService(db)
	if err != nil {
		t.Fatal(err)
	}
	return db, service
}
