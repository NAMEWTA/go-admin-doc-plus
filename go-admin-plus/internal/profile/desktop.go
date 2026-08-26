package profile

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	cachememory "github.com/go-admin-team/go-admin-core/v2/storage/cache"
	queuememory "github.com/go-admin-team/go-admin-core/v2/storage/queue"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"go-admin/internal/platform"
	"go-admin/internal/tenant"
)

const (
	desktopDatabaseName = "go-admin.sqlite3"
	desktopBusyTimeout  = 5000
)

// DesktopConfig identifies the platform application-data root. QueueSize <= 0
// uses the in-memory queue default.
type DesktopConfig struct {
	DataRoot  string
	QueueSize int
}

// DesktopLayout contains canonical paths outside the installation directory.
type DesktopLayout struct {
	Root         string
	DatabaseDir  string
	DatabasePath string
	FilesDir     string
	LogsDir      string
	BackupsDir   string
	TempDir      string
}

// BuildDesktop creates SQLite, memory Cache/Queue, local FileStore and the
// fixed local tenant. The caller owns the returned Dependencies.
func BuildDesktop(ctx context.Context, config DesktopConfig) (*platform.Dependencies, DesktopLayout, error) {
	if ctx == nil {
		return nil, DesktopLayout{}, errors.New("build context is required")
	}
	if err := ctx.Err(); err != nil {
		return nil, DesktopLayout{}, err
	}
	layout, err := createDesktopLayout(config.DataRoot)
	if err != nil {
		return nil, DesktopLayout{}, err
	}

	database, sqlDatabase, err := openDesktopDatabase(ctx, layout.DatabasePath)
	if err != nil {
		return nil, DesktopLayout{}, err
	}
	cache := cachememory.NewMemCache()
	queue := queuememory.NewMemQueue(config.QueueSize)
	files, err := platform.NewLocalFileStore(layout.FilesDir)
	if err != nil {
		queue.Close()
		cache.Close()
		sqlDatabase.Close()
		return nil, DesktopLayout{}, fmt.Errorf("build desktop file store: %w", err)
	}

	dependencies, err := platform.NewDependencies(platform.AdapterSet{
		Database: database,
		Cache:    cache,
		Queue:    queue,
		Files:    files,
		Tenants:  tenant.Fixed("local"),
	},
		platform.ResourceStopper{Name: "sqlite database", Stop: ignoreContext(sqlDatabase.Close)},
		platform.ResourceStopper{Name: "memory cache", Stop: ignoreContext(cache.Close)},
		platform.ResourceStopper{Name: "memory queue", Stop: ignoreContext(queue.Close)},
		platform.ResourceStopper{Name: "desktop file store", Stop: ignoreContext(files.Close)},
	)
	if err != nil {
		files.Close()
		queue.Close()
		cache.Close()
		sqlDatabase.Close()
		return nil, DesktopLayout{}, fmt.Errorf("assemble desktop adapters: %w", err)
	}
	return dependencies, layout, nil
}

func createDesktopLayout(root string) (DesktopLayout, error) {
	root = strings.TrimSpace(root)
	if root == "" {
		return DesktopLayout{}, errors.New("desktop data root is required")
	}
	absolute, err := filepath.Abs(root)
	if err != nil {
		return DesktopLayout{}, fmt.Errorf("resolve desktop data root: %w", err)
	}
	if err := os.MkdirAll(absolute, 0o700); err != nil {
		return DesktopLayout{}, fmt.Errorf("create desktop data root: %w", err)
	}
	realRoot, err := filepath.EvalSymlinks(absolute)
	if err != nil {
		return DesktopLayout{}, fmt.Errorf("resolve desktop data root links: %w", err)
	}
	if err := os.Chmod(realRoot, 0o700); err != nil {
		return DesktopLayout{}, fmt.Errorf("secure desktop data root: %w", err)
	}
	layout := DesktopLayout{
		Root:        realRoot,
		DatabaseDir: filepath.Join(realRoot, "db"),
		FilesDir:    filepath.Join(realRoot, "files"),
		LogsDir:     filepath.Join(realRoot, "logs"),
		BackupsDir:  filepath.Join(realRoot, "backups"),
		TempDir:     filepath.Join(realRoot, "temp"),
	}
	layout.DatabasePath = filepath.Join(layout.DatabaseDir, desktopDatabaseName)
	for _, directory := range []string{layout.DatabaseDir, layout.FilesDir, layout.LogsDir, layout.BackupsDir, layout.TempDir} {
		if err := os.MkdirAll(directory, 0o700); err != nil {
			return DesktopLayout{}, fmt.Errorf("create desktop data directory: %w", err)
		}
		if err := os.Chmod(directory, 0o700); err != nil {
			return DesktopLayout{}, fmt.Errorf("secure desktop data directory: %w", err)
		}
	}
	return layout, nil
}

func openDesktopDatabase(ctx context.Context, path string) (*gorm.DB, *sql.DB, error) {
	database, err := gorm.Open(sqlite.Open(desktopSQLiteDSN(path)), &gorm.Config{
		DisableAutomaticPing: true,
		NamingStrategy:       schema.NamingStrategy{SingularTable: true},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("open desktop database: %w", err)
	}
	sqlDatabase, err := database.DB()
	if err != nil {
		return nil, nil, fmt.Errorf("access desktop database connection: %w", err)
	}
	sqlDatabase.SetMaxOpenConns(1)
	sqlDatabase.SetMaxIdleConns(1)
	if err := sqlDatabase.PingContext(ctx); err != nil {
		sqlDatabase.Close()
		return nil, nil, fmt.Errorf("ping desktop database: %w", err)
	}

	checks := []struct {
		query string
		want  string
	}{
		{query: "PRAGMA foreign_keys", want: "1"},
		{query: "PRAGMA busy_timeout", want: fmt.Sprint(desktopBusyTimeout)},
		{query: "PRAGMA journal_mode", want: "wal"},
	}
	for _, check := range checks {
		var value string
		if err := database.WithContext(ctx).Raw(check.query).Scan(&value).Error; err != nil {
			sqlDatabase.Close()
			return nil, nil, fmt.Errorf("verify desktop database policy: %w", err)
		}
		if !strings.EqualFold(value, check.want) {
			sqlDatabase.Close()
			return nil, nil, fmt.Errorf("desktop database policy %q is %q, want %q", check.query, value, check.want)
		}
	}
	return database, sqlDatabase, nil
}

func desktopSQLiteDSN(path string) string {
	normalized := filepath.ToSlash(path)
	if isWindowsDrivePath(normalized) {
		normalized = "/" + normalized
	}
	return (&url.URL{
		Scheme:   "file",
		Path:     normalized,
		RawQuery: "_foreign_keys=on&_busy_timeout=5000&_journal_mode=WAL",
	}).String()
}

func isWindowsDrivePath(path string) bool {
	if len(path) < 3 || path[1] != ':' || path[2] != '/' {
		return false
	}
	return (path[0] >= 'A' && path[0] <= 'Z') || (path[0] >= 'a' && path[0] <= 'z')
}

func ignoreContext(close func() error) platform.StopFunc {
	return func(context.Context) error { return close() }
}
