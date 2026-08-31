package datascopemigration

import (
	"io/fs"
	"strings"
	"testing"

	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/platform/database"
)

func TestMigrationPublishesEquivalentScopeTablesForBothDialects(t *testing.T) {
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
		for _, required := range []string{"iam_role_data_scopes", "organization-tree", "iam_account_organization", "iam_account_positions", "iam_role_data_scope_departments"} {
			if !strings.Contains(text, required) {
				t.Fatalf("%s migration is missing %q", dialect, required)
			}
		}
	}
}
