package datascopemigration

import (
	"io/fs"
	"strings"
	"testing"

	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/platform/database"
)

func TestMigrationPublishesEquivalentRoleScopesForBothDialects(t *testing.T) {
	for _, dialect := range []database.Dialect{database.DialectSQLite, database.DialectPostgres} {
		migrationFS, err := (Provider{}).Migrations(dialect)
		if err != nil {
			t.Fatal(err)
		}
		content, err := fs.ReadFile(migrationFS, "7110000000000_iam_data_scope.sql")
		if err != nil {
			t.Fatal(err)
		}
		text := string(content)
		for _, required := range []string{"iam_role_data_scopes", "CHECK (scope IN ('all','self'))"} {
			if !strings.Contains(text, required) {
				t.Fatalf("%s migration is missing %q", dialect, required)
			}
		}
		for _, removed := range []string{"organization-tree", "iam_account_organization", "iam_account_positions", "iam_role_data_scope_departments"} {
			if strings.Contains(text, removed) {
				t.Fatalf("%s migration still contains removed organization scope %q", dialect, removed)
			}
		}
	}
}
