package profile_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"go-admin/cmd/migrate/migration"
	commonmodels "go-admin/common/models"
	"go-admin/internal/profile"
)

type oldDesktopRecord struct {
	ID   int `gorm:"primaryKey"`
	Name string
}

func (oldDesktopRecord) TableName() string { return "desktop_records" }

type upgradedDesktopRecord struct {
	ID   int `gorm:"primaryKey"`
	Name string
	Note string
}

func (upgradedDesktopRecord) TableName() string { return "desktop_records" }

func TestUpgradeDesktopBacksUpOldFixtureBeforeSuccessfulMigration(t *testing.T) {
	config := profile.DesktopConfig{DataRoot: filepath.Join(t.TempDir(), "app-data")}
	dependencies, layout, err := profile.BuildDesktop(context.Background(), config)
	if err != nil {
		t.Fatalf("BuildDesktop: %v", err)
	}
	if err := dependencies.Database().AutoMigrate(&oldDesktopRecord{}, &commonmodels.Migration{}); err != nil {
		t.Fatalf("create old schema: %v", err)
	}
	if err := dependencies.Database().Create(&oldDesktopRecord{ID: 1, Name: "kept"}).Error; err != nil {
		t.Fatalf("create old data: %v", err)
	}
	if err := dependencies.Close(context.Background()); err != nil {
		t.Fatalf("close old profile: %v", err)
	}

	runner := migration.New()
	if err := runner.Register("1786700000001", func(db *gorm.DB, version string) error {
		if err := db.AutoMigrate(&upgradedDesktopRecord{}); err != nil {
			return err
		}
		return db.Create(&commonmodels.Migration{Version: version}).Error
	}); err != nil {
		t.Fatalf("Register: %v", err)
	}
	result, err := profile.UpgradeDesktop(context.Background(), config, func(ctx context.Context, db *gorm.DB) error {
		_, err := runner.Run(ctx, db)
		return err
	})
	if err != nil {
		t.Fatalf("UpgradeDesktop: %v", err)
	}
	backupDatabase := filepath.Join(result.BackupPath, filepath.Base(layout.DatabasePath))
	if _, err := os.Stat(backupDatabase); err != nil {
		t.Fatalf("stat backup database: %v", err)
	}

	upgraded := openSQLiteFile(t, layout.DatabasePath)
	if !upgraded.Migrator().HasColumn(&upgradedDesktopRecord{}, "Note") {
		t.Fatal("upgraded database does not have note column")
	}
	var current oldDesktopRecord
	if err := upgraded.First(&current, 1).Error; err != nil || current.Name != "kept" {
		t.Fatalf("upgraded data = %#v, %v; want kept", current, err)
	}

	backup := openSQLiteFile(t, backupDatabase)
	if backup.Migrator().HasColumn(&upgradedDesktopRecord{}, "Note") {
		t.Fatal("pre-migration backup unexpectedly has new note column")
	}
	var backedUp oldDesktopRecord
	if err := backup.First(&backedUp, 1).Error; err != nil || backedUp.Name != "kept" {
		t.Fatalf("backup data = %#v, %v; want kept", backedUp, err)
	}
}

func TestUpgradeDesktopKeepsOriginalDataAndBackupWhenMigrationFails(t *testing.T) {
	config := profile.DesktopConfig{DataRoot: filepath.Join(t.TempDir(), "app-data")}
	dependencies, layout, err := profile.BuildDesktop(context.Background(), config)
	if err != nil {
		t.Fatalf("BuildDesktop: %v", err)
	}
	if err := dependencies.Database().AutoMigrate(&oldDesktopRecord{}, &commonmodels.Migration{}); err != nil {
		t.Fatalf("create old schema: %v", err)
	}
	if err := dependencies.Database().Create(&oldDesktopRecord{ID: 1, Name: "before"}).Error; err != nil {
		t.Fatalf("create old data: %v", err)
	}
	if err := dependencies.Close(context.Background()); err != nil {
		t.Fatalf("close old profile: %v", err)
	}

	wantErr := errors.New("injected desktop migration failure")
	runner := migration.New()
	if err := runner.Register("1786700000001", func(db *gorm.DB, version string) error {
		if err := db.Model(&oldDesktopRecord{}).Where("id = ?", 1).Update("name", "after").Error; err != nil {
			return err
		}
		return wantErr
	}); err != nil {
		t.Fatalf("Register: %v", err)
	}
	result, err := profile.UpgradeDesktop(context.Background(), config, func(ctx context.Context, db *gorm.DB) error {
		_, err := runner.Run(ctx, db)
		return err
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("UpgradeDesktop error = %v, want injected failure", err)
	}
	if result.BackupPath == "" {
		t.Fatal("UpgradeDesktop did not preserve a pre-migration backup")
	}

	for name, path := range map[string]string{
		"current": layout.DatabasePath,
		"backup":  filepath.Join(result.BackupPath, filepath.Base(layout.DatabasePath)),
	} {
		database := openSQLiteFile(t, path)
		var record oldDesktopRecord
		if err := database.First(&record, 1).Error; err != nil || record.Name != "before" {
			t.Fatalf("%s data = %#v, %v; want before", name, record, err)
		}
	}
}

func TestUpgradeDesktopDoesNotMigrateWhenBackupCannotBeCreated(t *testing.T) {
	config := profile.DesktopConfig{DataRoot: filepath.Join(t.TempDir(), "app-data")}
	dependencies, layout, err := profile.BuildDesktop(context.Background(), config)
	if err != nil {
		t.Fatalf("BuildDesktop: %v", err)
	}
	if err := dependencies.Close(context.Background()); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if err := os.Remove(layout.DatabasePath); err != nil {
		t.Fatalf("remove fixture database: %v", err)
	}
	if err := os.Mkdir(layout.DatabasePath, 0o700); err != nil {
		t.Fatalf("replace fixture database with invalid source: %v", err)
	}

	migrationCalled := false
	result, err := profile.UpgradeDesktop(context.Background(), config, func(context.Context, *gorm.DB) error {
		migrationCalled = true
		return nil
	})
	if err == nil {
		t.Fatal("UpgradeDesktop unexpectedly succeeded with an invalid backup source")
	}
	if migrationCalled {
		t.Fatal("migration ran after backup failure")
	}
	if result.BackupPath != "" {
		t.Fatalf("BackupPath = %q, want no published incomplete backup", result.BackupPath)
	}
}

func openSQLiteFile(t *testing.T, path string) *gorm.DB {
	t.Helper()
	database, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite file %s: %v", path, err)
	}
	t.Cleanup(func() {
		sqlDatabase, err := database.DB()
		if err == nil {
			sqlDatabase.Close()
		}
	})
	return database
}
