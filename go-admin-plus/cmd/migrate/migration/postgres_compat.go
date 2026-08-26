package migration

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// withDialectCompatibility keeps published migrations independent from
// driver-specific DDL bugs without changing their immutable version files.
func withDialectCompatibility(database *gorm.DB, version string) *gorm.DB {
	if database.Dialector.Name() != "postgres" || version != "1786700003000" {
		return database
	}

	compatible := database.Session(&gorm.Session{Context: database.Statement.Context})
	compatible.Dialector = postgresMigrationDialector{Dialector: database.Dialector}
	return compatible
}

type postgresMigrationDialector struct {
	gorm.Dialector
}

func (dialector postgresMigrationDialector) Migrator(database *gorm.DB) gorm.Migrator {
	return postgresMigrationMigrator{
		Migrator: dialector.Dialector.Migrator(database),
		database: database,
	}
}

type postgresMigrationMigrator struct {
	gorm.Migrator
	database *gorm.DB
}

// DropIndex avoids gorm.io/driver/postgres v1.6.2 rendering the invalid
// PostgreSQL identifier CURRENT_SCHEMA()."index_name".
func (migrator postgresMigrationMigrator) DropIndex(value interface{}, name string) error {
	schemaName, err := migrator.schemaName(value)
	if err != nil {
		return err
	}
	return migrator.database.Exec(
		"DROP INDEX ?.?",
		clause.Column{Name: schemaName},
		clause.Column{Name: name},
	).Error
}

func (migrator postgresMigrationMigrator) schemaName(value interface{}) (string, error) {
	if table, ok := value.(string); ok {
		if schemaName, _, qualified := strings.Cut(table, "."); qualified && schemaName != "" {
			return schemaName, nil
		}
	}

	var schemaName string
	if err := migrator.database.Raw("SELECT CURRENT_SCHEMA()").Scan(&schemaName).Error; err != nil {
		return "", fmt.Errorf("resolve PostgreSQL migration schema: %w", err)
	}
	if schemaName == "" {
		return "", fmt.Errorf("resolve PostgreSQL migration schema: empty schema")
	}
	return schemaName, nil
}
