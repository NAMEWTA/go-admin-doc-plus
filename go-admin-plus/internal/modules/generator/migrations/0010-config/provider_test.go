package configmigration

import (
	"io/fs"
	"strings"
	"testing"

	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/platform/database"
)

func TestActorAccountIDMatchesIAMTextIdentityAcrossDialects(t *testing.T) {
	for _, dialect := range []database.Dialect{database.DialectSQLite, database.DialectPostgres} {
		migrations, err := (Provider{}).Migrations(dialect)
		if err != nil {
			t.Fatal(err)
		}
		content, err := fs.ReadFile(migrations, "6400000000000_generator_configs.sql")
		if err != nil {
			t.Fatal(err)
		}
		statement := string(content)
		if !strings.Contains(statement, "actor_account_id TEXT NOT NULL") || strings.Contains(statement, "actor_account_id UUID") {
			t.Fatalf("%s generator actor identity diverges from IAM TEXT IDs", dialect)
		}
	}
}
