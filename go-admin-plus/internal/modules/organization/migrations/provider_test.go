package organizationmigration

import (
	"io/fs"
	"strings"
	"testing"

	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/platform/database"
)

func TestOrganizationMigrationsHaveEquivalentPrivateSchema(t *testing.T) {
	for _, dialect := range []database.Dialect{database.DialectPostgres, database.DialectSQLite} {
		files, err := (Provider{}).Migrations(dialect)
		if err != nil {
			t.Fatal(err)
		}
		content, err := fs.ReadFile(files, "7100000000000_organization.sql")
		if err != nil {
			t.Fatal(err)
		}
		text := strings.ToLower(string(content))
		for _, required := range []string{"organization_departments", "organization_positions", "department-root-001", "on delete restrict"} {
			if !strings.Contains(text, required) {
				t.Fatalf("%s migration misses %q", dialect, required)
			}
		}
		for _, forbidden := range []string{"tenant", "iam_accounts", "iam_roles", "casbin", "jwt"} {
			if strings.Contains(text, forbidden) {
				t.Fatalf("%s migration contains forbidden cross-module state %q", dialect, forbidden)
			}
		}
	}
}
