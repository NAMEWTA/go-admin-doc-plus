package product

import (
	"context"
	"errors"

	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/platform/config"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/platform/database"
)

// MigrationResult is the command-facing projection of one forward migration.
// It deliberately does not expose the platform migration implementation.
type MigrationResult struct {
	Applied        int
	CurrentVersion int64
}

// Migrate owns database selection, process lifecycle and product migration
// composition for one typed Server profile.
func Migrate(ctx context.Context, snapshot config.Snapshot) (result MigrationResult, resultErr error) {
	if ctx == nil {
		return MigrationResult{}, errors.New("migration context is required")
	}
	databaseConfig, err := migrationDatabaseConfig(snapshot)
	if err != nil {
		return MigrationResult{}, err
	}
	db, err := database.NewProcess().Open(ctx, databaseConfig)
	if err != nil {
		return MigrationResult{}, errors.New("migration database startup failed")
	}
	defer func() {
		if err := db.Close(); err != nil {
			resultErr = errors.Join(resultErr, errors.New("migration database shutdown failed"))
		}
	}()
	runner, err := NewMigrationRunner()
	if err != nil {
		return MigrationResult{}, errors.New("migration composition failed")
	}
	applied, err := runner.Up(ctx, db)
	if err != nil {
		return MigrationResult{}, err
	}
	return MigrationResult{Applied: applied.Applied, CurrentVersion: applied.CurrentVersion}, nil
}

func migrationDatabaseConfig(snapshot config.Snapshot) (database.Config, error) {
	if profile, ok := snapshot.ServerSQLite(); ok {
		return database.Config{Profile: config.ProfileServerSQLite, SQLitePath: profile.DatabasePath()}, nil
	}
	if profile, ok := snapshot.ServerPostgres(); ok {
		return database.Config{Profile: config.ProfileServerPostgres, PostgresDSN: profile.DatabaseDSN()}, nil
	}
	return database.Config{}, errors.New("migration runtime profile is invalid")
}
