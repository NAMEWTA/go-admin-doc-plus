package administration

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/iam/authorization"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/platform/config"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/platform/database"
)

func TestProjectEffectiveRoleScopesUsesProjectionAndLegacyFallback(t *testing.T) {
	ctx := context.Background()
	db, err := database.NewProcess().Open(ctx, database.Config{
		Profile:    config.ProfileServerSQLite,
		SQLitePath: filepath.Join(t.TempDir(), "role-scopes.sqlite3"),
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if _, err := db.Bun().ExecContext(ctx, `CREATE TABLE iam_role_data_scopes (role_id TEXT PRIMARY KEY, scope TEXT NOT NULL)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Bun().ExecContext(ctx, `INSERT INTO iam_role_data_scopes(role_id, scope) VALUES (?, ?)`, "projected-role", authorization.ScopeCustom); err != nil {
		t.Fatal(err)
	}

	server := &HTTPServer{dataScopes: &DataScopeService{db: db}}
	roles := []Role{
		{ID: "projected-role", Scope: authorization.ScopeSelf},
		{ID: "legacy-role", Scope: authorization.ScopeAll},
	}
	if err := server.projectEffectiveRoleScopes(ctx, roles); err != nil {
		t.Fatal(err)
	}
	if roles[0].Scope != authorization.ScopeCustom || roles[1].Scope != authorization.ScopeAll {
		t.Fatalf("effective scopes = %q, %q", roles[0].Scope, roles[1].Scope)
	}
}

func TestProjectEffectiveRoleScopesRejectsInvalidStoredScope(t *testing.T) {
	ctx := context.Background()
	db, err := database.NewProcess().Open(ctx, database.Config{
		Profile:    config.ProfileServerSQLite,
		SQLitePath: filepath.Join(t.TempDir(), "invalid-role-scope.sqlite3"),
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if _, err := db.Bun().ExecContext(ctx, `CREATE TABLE iam_role_data_scopes (role_id TEXT PRIMARY KEY, scope TEXT NOT NULL)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Bun().ExecContext(ctx, `INSERT INTO iam_role_data_scopes(role_id, scope) VALUES ('role-invalid', 'unexpected')`); err != nil {
		t.Fatal(err)
	}

	server := &HTTPServer{dataScopes: &DataScopeService{db: db}}
	roles := []Role{{ID: "role-invalid", Scope: authorization.ScopeSelf}}
	if err := server.projectEffectiveRoleScopes(ctx, roles); err == nil {
		t.Fatal("invalid persisted scope was accepted")
	}
}

func TestShouldSyncBaseRoleScopePreservesExtendedScopes(t *testing.T) {
	ctx := context.Background()
	db, err := database.NewProcess().Open(ctx, database.Config{
		Profile:    config.ProfileServerSQLite,
		SQLitePath: filepath.Join(t.TempDir(), "base-role-scopes.sqlite3"),
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if _, err := db.Bun().ExecContext(ctx, `CREATE TABLE iam_role_data_scopes (role_id TEXT PRIMARY KEY, scope TEXT NOT NULL)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Bun().ExecContext(ctx, `INSERT INTO iam_role_data_scopes(role_id, scope) VALUES ('base-role', 'self'), ('extended-role', 'organization-tree')`); err != nil {
		t.Fatal(err)
	}
	server := &HTTPServer{dataScopes: &DataScopeService{db: db}}
	for _, test := range []struct {
		roleID string
		want   bool
	}{
		{roleID: "base-role", want: true},
		{roleID: "missing-role", want: true},
		{roleID: "extended-role", want: false},
	} {
		got, err := server.shouldSyncBaseRoleScope(ctx, test.roleID)
		if err != nil || got != test.want {
			t.Fatalf("shouldSyncBaseRoleScope(%q) = %t, %v; want %t", test.roleID, got, err, test.want)
		}
	}
}
