package reliableruntime_test

import (
	"testing"

	"go-admin/internal/platform/database"
	"go-admin/internal/platform/migrations"
	reliablemigration "go-admin/internal/platform/migrations/reliable-runtime"
)

func TestReliableRuntimeMigrationComposesForBothDialects(t *testing.T) {
	t.Parallel()
	runner, err := migrations.NewRunner(reliablemigration.Provider{})
	if err != nil {
		t.Fatalf("NewRunner() error = %v", err)
	}
	for _, dialect := range []database.Dialect{database.DialectPostgres, database.DialectSQLite} {
		if _, err := runner.Compose(dialect); err != nil {
			t.Fatalf("Compose(%s) error = %v", dialect, err)
		}
	}
}
