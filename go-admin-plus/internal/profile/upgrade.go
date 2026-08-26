package profile

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"gorm.io/gorm"
)

// MigrationFunc applies the one shared application migration registry.
type MigrationFunc func(context.Context, *gorm.DB) error

// DesktopUpgradeResult always returns the canonical layout and publishes a
// BackupPath only when the pre-migration backup completed.
type DesktopUpgradeResult struct {
	Layout     DesktopLayout
	BackupPath string
}

// UpgradeDesktop requires the caller to hold the desktop single-instance lock
// and to have no open database connections. It publishes a backup path only
// after the main database and any SQLite sidecars have been copied completely.
func UpgradeDesktop(ctx context.Context, config DesktopConfig, migrate MigrationFunc) (DesktopUpgradeResult, error) {
	if ctx == nil {
		return DesktopUpgradeResult{}, errors.New("upgrade context is required")
	}
	if migrate == nil {
		return DesktopUpgradeResult{}, errors.New("desktop migration function is required")
	}
	if err := ctx.Err(); err != nil {
		return DesktopUpgradeResult{}, err
	}
	layout, err := createDesktopLayout(config.DataRoot)
	if err != nil {
		return DesktopUpgradeResult{}, err
	}
	result := DesktopUpgradeResult{Layout: layout}
	backupPath, err := backupSQLite(ctx, layout.DatabasePath, layout.BackupsDir)
	if err != nil {
		return result, fmt.Errorf("back up desktop database: %w", err)
	}
	result.BackupPath = backupPath

	database, sqlDatabase, err := openDesktopDatabase(ctx, layout.DatabasePath)
	if err != nil {
		return result, err
	}
	migrationErr := migrate(ctx, database)
	closeErr := sqlDatabase.Close()
	if migrationErr != nil || closeErr != nil {
		return result, errors.Join(
			wrapIfPresent("migrate desktop database", migrationErr),
			wrapIfPresent("close desktop database", closeErr),
		)
	}
	return result, nil
}

func backupSQLite(ctx context.Context, databasePath, backupsDir string) (string, error) {
	info, err := os.Stat(databasePath)
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", errors.New("desktop database is not a regular file")
	}
	backupPath, err := os.MkdirTemp(backupsDir, "pre-migration-*")
	if err != nil {
		return "", err
	}
	complete := false
	defer func() {
		if !complete {
			_ = os.RemoveAll(backupPath)
		}
	}()

	for _, source := range []string{databasePath, databasePath + "-wal", databasePath + "-shm"} {
		if err := copyBackupFile(ctx, source, filepath.Join(backupPath, filepath.Base(source))); errors.Is(err, os.ErrNotExist) {
			continue
		} else if err != nil {
			return "", err
		}
	}
	if err := syncDirectory(backupPath); err != nil {
		return "", err
	}
	complete = true
	return backupPath, nil
}

func copyBackupFile(ctx context.Context, sourcePath, destinationPath string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	source, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer source.Close()
	destination, err := os.OpenFile(destinationPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(destination, &upgradeContextReader{ctx: ctx, source: source})
	syncErr := destination.Sync()
	closeErr := destination.Close()
	return errors.Join(copyErr, syncErr, closeErr)
}

type upgradeContextReader struct {
	ctx    context.Context
	source io.Reader
}

func (r *upgradeContextReader) Read(buffer []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.source.Read(buffer)
}

func wrapIfPresent(operation string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", operation, err)
}
