package migration_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"go-admin/cmd/migrate/migration"
	commonmodels "go-admin/common/models"
)

type fixtureRecord struct {
	ID    int `gorm:"primaryKey"`
	Value string
}

func TestRunnerAppliesVersionsInOrderAndSkipsAppliedVersions(t *testing.T) {
	database := openMigrationFixture(t)
	runner := migration.New()
	var order []string
	for _, version := range []string{"1786700000002", "1786700000001"} {
		version := version
		if err := runner.Register(version, func(db *gorm.DB, current string) error {
			order = append(order, current)
			return db.Create(&commonmodels.Migration{Version: current}).Error
		}); err != nil {
			t.Fatalf("Register(%s): %v", version, err)
		}
	}

	result, err := runner.Run(context.Background(), database)
	if err != nil {
		t.Fatalf("first Run: %v", err)
	}
	if want := []string{"1786700000001", "1786700000002"}; !reflect.DeepEqual(result.Applied, want) {
		t.Fatalf("Applied = %v, want %v", result.Applied, want)
	}
	if want := []string{"1786700000001", "1786700000002"}; !reflect.DeepEqual(order, want) {
		t.Fatalf("execution order = %v, want %v", order, want)
	}

	result, err = runner.Run(context.Background(), database)
	if err != nil {
		t.Fatalf("second Run: %v", err)
	}
	if len(result.Applied) != 0 || !reflect.DeepEqual(result.Skipped, order) {
		t.Fatalf("second result = %#v, want all versions skipped", result)
	}
}

func TestRunnerRollsBackEveryVersionWhenOneFails(t *testing.T) {
	database := openMigrationFixture(t)
	if err := database.AutoMigrate(&fixtureRecord{}, &commonmodels.Migration{}); err != nil {
		t.Fatalf("AutoMigrate fixture: %v", err)
	}
	if err := database.Create(&fixtureRecord{ID: 1, Value: "before"}).Error; err != nil {
		t.Fatalf("create fixture: %v", err)
	}

	wantErr := errors.New("injected migration failure")
	runner := migration.New()
	if err := runner.Register("1786700000001", func(db *gorm.DB, version string) error {
		if err := db.Model(&fixtureRecord{}).Where("id = ?", 1).Update("value", "after").Error; err != nil {
			return err
		}
		if err := db.Create(&commonmodels.Migration{Version: version}).Error; err != nil {
			return err
		}
		return wantErr
	}); err != nil {
		t.Fatalf("Register: %v", err)
	}

	if _, err := runner.Run(context.Background(), database); !errors.Is(err, wantErr) {
		t.Fatalf("Run error = %v, want injected failure", err)
	}
	var record fixtureRecord
	if err := database.First(&record, 1).Error; err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	if record.Value != "before" {
		t.Fatalf("fixture value = %q, want rollback value before", record.Value)
	}
	var count int64
	if err := database.Model(&commonmodels.Migration{}).Where("version = ?", "1786700000001").Count(&count).Error; err != nil {
		t.Fatalf("count migration record: %v", err)
	}
	if count != 0 {
		t.Fatalf("migration record count = %d, want 0", count)
	}
}

func openMigrationFixture(t *testing.T) *gorm.DB {
	t.Helper()
	database, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite fixture: %v", err)
	}
	return database
}
