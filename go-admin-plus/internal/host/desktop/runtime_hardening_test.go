package desktop

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	corelogger "github.com/go-admin-team/go-admin-core/v2/logger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	desktopplatform "go-admin/internal/platform/desktop"
)

type durabilityProbe struct {
	ID    int `gorm:"primaryKey"`
	Value string
}

func TestHostDoesNotOpenWindowWhenPreMigrationBackupFails(t *testing.T) {
	root := filepath.Join(t.TempDir(), "app-data")
	databasePath := filepath.Join(root, "db", "go-admin.sqlite3")
	if err := os.MkdirAll(databasePath, 0o700); err != nil {
		t.Fatalf("create invalid database fixture: %v", err)
	}
	windowRan := false
	host, err := New(
		Config{Assets: fstest.MapFS{"index.html": &fstest.MapFile{Data: []byte("desktop")}}},
		NewRuntimeBuilder(RuntimeConfig{DataRoot: root, Mode: "dev"}),
		func(context.Context, *Bridge, fs.FS) error {
			windowRan = true
			return nil
		},
	)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := host.Run(context.Background()); err == nil {
		t.Fatal("Run unexpectedly succeeded with invalid database fixture")
	}
	if windowRan {
		t.Fatal("window ran after backup failure")
	}
	lock, err := desktopplatform.AcquireInstanceLock(root)
	if err != nil {
		t.Fatalf("instance lock was not released after startup failure: %v", err)
	}
	if err := lock.Close(); err != nil {
		t.Fatalf("close verification lock: %v", err)
	}
}

func TestRuntimeBuilderHoldsInstanceLockUntilRuntimeClose(t *testing.T) {
	config := RuntimeConfig{DataRoot: filepath.Join(t.TempDir(), "app-data"), Mode: "dev"}
	builder := NewRuntimeBuilder(config)
	first, err := builder(context.Background())
	if err != nil {
		t.Fatalf("build first runtime: %v", err)
	}
	if _, err := builder(context.Background()); !errors.Is(err, desktopplatform.ErrInstanceLocked) {
		t.Fatalf("build second runtime error = %v, want ErrInstanceLocked", err)
	}
	if err := first.Close(context.Background()); err != nil {
		t.Fatalf("close first runtime: %v", err)
	}

	second, err := builder(context.Background())
	if err != nil {
		t.Fatalf("build after release: %v", err)
	}
	if err := second.Close(context.Background()); err != nil {
		t.Fatalf("close second runtime: %v", err)
	}
}

func TestRuntimeBuilderClosesDesktopLogAndRestoresPreviousLogger(t *testing.T) {
	root := filepath.Join(t.TempDir(), "app-data")
	previousLogger := corelogger.DefaultLogger
	runtime, err := NewRuntimeBuilder(RuntimeConfig{DataRoot: root, Mode: "dev"})(context.Background())
	if err != nil {
		t.Fatalf("build runtime: %v", err)
	}
	if corelogger.DefaultLogger == previousLogger {
		t.Fatal("desktop runtime did not install its owned logger")
	}
	corelogger.Info("desktop log close probe")
	logPath := filepath.Join(root, "logs", "app.log")
	if _, err := os.Stat(logPath); err != nil {
		t.Fatalf("stat desktop log: %v", err)
	}

	if err := runtime.Close(context.Background()); err != nil {
		t.Fatalf("close runtime: %v", err)
	}
	if corelogger.DefaultLogger != previousLogger {
		t.Fatal("desktop runtime did not restore the previous logger")
	}
	if err := os.Rename(logPath, logPath+".closed"); err != nil {
		t.Fatalf("move closed desktop log: %v", err)
	}
	contents, err := os.ReadFile(logPath + ".closed")
	if err != nil {
		t.Fatalf("read closed desktop log: %v", err)
	}
	if !strings.Contains(string(contents), "desktop log close probe") {
		t.Fatalf("desktop log omitted close probe: %q", contents)
	}
}

func TestRuntimeBuilderBacksUpExistingDatabaseBeforeStartupMigration(t *testing.T) {
	root := filepath.Join(t.TempDir(), "app-data")
	builder := NewRuntimeBuilder(RuntimeConfig{DataRoot: root, Mode: "dev"})
	first, err := builder(context.Background())
	if err != nil {
		t.Fatalf("build first runtime: %v", err)
	}
	if err := first.Close(context.Background()); err != nil {
		t.Fatalf("close first runtime: %v", err)
	}

	databasePath := filepath.Join(root, "db", "go-admin.sqlite3")
	database, err := gorm.Open(sqlite.Open(databasePath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open fixture database: %v", err)
	}
	if err := database.AutoMigrate(&durabilityProbe{}); err != nil {
		t.Fatalf("migrate fixture: %v", err)
	}
	if err := database.Create(&durabilityProbe{ID: 1, Value: "preserved"}).Error; err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	sqlDatabase, err := database.DB()
	if err != nil {
		t.Fatalf("access fixture connection: %v", err)
	}
	if err := sqlDatabase.Close(); err != nil {
		t.Fatalf("close fixture database: %v", err)
	}

	second, err := builder(context.Background())
	if err != nil {
		t.Fatalf("build second runtime: %v", err)
	}
	if err := second.Close(context.Background()); err != nil {
		t.Fatalf("close second runtime: %v", err)
	}

	backups, err := filepath.Glob(filepath.Join(root, "backups", "pre-migration-*", "go-admin.sqlite3"))
	if err != nil {
		t.Fatalf("glob backups: %v", err)
	}
	if len(backups) == 0 {
		t.Fatal("no pre-migration database backup was created")
	}
	backup, err := gorm.Open(sqlite.Open(backups[len(backups)-1]), &gorm.Config{})
	if err != nil {
		t.Fatalf("open backup: %v", err)
	}
	var probe durabilityProbe
	if err := backup.First(&probe, 1).Error; err != nil {
		t.Fatalf("read backup fixture: %v", err)
	}
	if probe.Value != "preserved" {
		t.Fatalf("backup value = %q, want preserved", probe.Value)
	}
	backupSQL, err := backup.DB()
	if err != nil {
		t.Fatalf("access backup connection: %v", err)
	}
	if err := backupSQL.Close(); err != nil {
		t.Fatalf("close backup: %v", err)
	}
}
