package migrate

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	coreconfig "github.com/go-admin-team/go-admin-core/v2/sdk/config"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"go-admin/cmd/migrate/migration"
)

type migrationFixtureMarker struct {
	ID    int `gorm:"primaryKey"`
	Value string
}

func TestMigrateDatabaseReturnsRunnerFailure(t *testing.T) {
	database, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	sqlDatabase, err := database.DB()
	if err != nil {
		t.Fatalf("access fixture database: %v", err)
	}
	t.Cleanup(func() {
		if err := sqlDatabase.Close(); err != nil {
			t.Errorf("close fixture database: %v", err)
		}
	})
	wantErr := errors.New("injected migration failure")
	runner := migration.New()
	if err := runner.Register("1786700000001", func(*gorm.DB, string) error { return wantErr }); err != nil {
		t.Fatalf("Register: %v", err)
	}

	if err := migrateDatabase(context.Background(), database, "sqlite3", runner); !errors.Is(err, wantErr) {
		t.Fatalf("migrateDatabase error = %v, want injected failure", err)
	}
}

func TestPublishedMigrationsUpgradeSQLiteFixtureWithoutLosingData(t *testing.T) {
	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	projectRoot := filepath.Clean(filepath.Join(workingDirectory, "../.."))
	if err := os.Chdir(projectRoot); err != nil {
		t.Fatalf("Chdir project root: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(workingDirectory); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})

	database, err := gorm.Open(sqlite.Open(t.TempDir()+"/published.sqlite3"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	sqlDatabase, err := database.DB()
	if err != nil {
		t.Fatalf("access fixture database: %v", err)
	}
	t.Cleanup(func() {
		if err := sqlDatabase.Close(); err != nil {
			t.Errorf("close fixture database: %v", err)
		}
	})
	if err := database.AutoMigrate(&migrationFixtureMarker{}); err != nil {
		t.Fatalf("create fixture schema: %v", err)
	}
	if err := database.Create(&migrationFixtureMarker{ID: 1, Value: "preserve-me"}).Error; err != nil {
		t.Fatalf("create fixture data: %v", err)
	}
	previousDriver := coreconfig.DatabaseConfig.Driver
	coreconfig.DatabaseConfig.Driver = "sqlite3"
	t.Cleanup(func() { coreconfig.DatabaseConfig.Driver = previousDriver })

	first, err := migration.Migrate.Run(context.Background(), database)
	if err != nil {
		t.Fatalf("first published migration run: %v", err)
	}
	if len(first.Applied) == 0 {
		t.Fatal("published migration registry is empty")
	}
	second, err := migration.Migrate.Run(context.Background(), database)
	if err != nil {
		t.Fatalf("second published migration run: %v", err)
	}
	if len(second.Applied) != 0 || !reflect.DeepEqual(second.Skipped, first.Applied) {
		t.Fatalf("second result = %#v, want all %v skipped", second, first.Applied)
	}
	var marker migrationFixtureMarker
	if err := database.First(&marker, 1).Error; err != nil || marker.Value != "preserve-me" {
		t.Fatalf("fixture marker = %#v, %v; want preserve-me", marker, err)
	}
}
