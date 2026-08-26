package migration

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"sync"

	"gorm.io/gorm"

	commonmodels "go-admin/common/models"
)

// Step applies one immutable version and writes exactly one sys_migration row.
type Step func(db *gorm.DB, version string) error

// Result reports versions committed or already present in sequence order.
type Result struct {
	Applied []string
	Skipped []string
}

// Migrate is the application registry populated by published version packages.
var Migrate = newApplicationRegistry()

// Migration owns an append-only registry and executes a snapshot atomically.
type Migration struct {
	mutex         sync.RWMutex
	db            *gorm.DB
	prerequisites []func(*gorm.DB) error
	version       map[string]Step
}

// New returns an empty registry for isolated modules and tests.
func New() *Migration {
	return &Migration{version: make(map[string]Step)}
}

func newApplicationRegistry() *Migration {
	registry := New()
	registry.prerequisites = append(registry.prerequisites, func(database *gorm.DB) error {
		return database.AutoMigrate(&casbinRule{})
	})
	return registry
}

// GetDb returns the database configured through the legacy SetDb seam.
func (m *Migration) GetDb() *gorm.DB {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.db
}

// SetDb supports the legacy command wiring. New hosts should pass the database
// explicitly to Run so profile ownership remains visible.
func (m *Migration) SetDb(db *gorm.DB) {
	m.mutex.Lock()
	m.db = db
	m.mutex.Unlock()
}

// SetVersion preserves registration by published migration files. Invalid or
// duplicate versions panic during process initialization instead of silently
// replacing an immutable migration.
func (m *Migration) SetVersion(version string, step Step) {
	if err := m.Register(version, step); err != nil {
		panic(err)
	}
}

// Register adds a unique 13-digit version. Registration does not permit
// replacement because published migrations are immutable.
func (m *Migration) Register(version string, step Step) error {
	if !validVersion(version) {
		return fmt.Errorf("migration version %q must contain exactly 13 digits", version)
	}
	if step == nil {
		return fmt.Errorf("migration version %q has no step", version)
	}
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if _, exists := m.version[version]; exists {
		return fmt.Errorf("migration version %q is already registered", version)
	}
	m.version[version] = step
	return nil
}

// Run applies the registered sequence atomically. Every step must write its
// own sys_migration record; missing records are treated as a failed migration.
func (m *Migration) Run(ctx context.Context, database *gorm.DB) (Result, error) {
	if ctx == nil {
		return Result{}, errors.New("migration context is required")
	}
	if database == nil {
		return Result{}, errors.New("migration database is required")
	}
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}

	versions, steps, prerequisites := m.snapshot()
	result := Result{}
	err := database.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&commonmodels.Migration{}); err != nil {
			return fmt.Errorf("prepare migration ledger: %w", err)
		}
		for _, prepare := range prerequisites {
			if err := prepare(tx); err != nil {
				return fmt.Errorf("prepare application migration schema: %w", err)
			}
		}
		for _, version := range versions {
			if err := ctx.Err(); err != nil {
				return err
			}
			var count int64
			if err := tx.Model(&commonmodels.Migration{}).Where("version = ?", version).Count(&count).Error; err != nil {
				return fmt.Errorf("read migration %s state: %w", version, err)
			}
			if count > 0 {
				result.Skipped = append(result.Skipped, version)
				continue
			}
			if err := steps[version](withDialectCompatibility(tx, version), version); err != nil {
				return fmt.Errorf("apply migration %s: %w", version, err)
			}
			if err := tx.Model(&commonmodels.Migration{}).Where("version = ?", version).Count(&count).Error; err != nil {
				return fmt.Errorf("verify migration %s state: %w", version, err)
			}
			if count != 1 {
				return fmt.Errorf("migration %s wrote %d ledger records, want 1", version, count)
			}
			result.Applied = append(result.Applied, version)
		}
		return nil
	})
	if err != nil {
		return Result{}, err
	}
	return result, nil
}

// Migrate is the temporary compatibility entry point for the legacy command.
func (m *Migration) Migrate() error {
	database := m.GetDb()
	_, err := m.Run(context.Background(), database)
	return err
}

func (m *Migration) snapshot() ([]string, map[string]Step, []func(*gorm.DB) error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	versions := make([]string, 0, len(m.version))
	steps := make(map[string]Step, len(m.version))
	for version, step := range m.version {
		versions = append(versions, version)
		steps[version] = step
	}
	sort.Strings(versions)
	prerequisites := append([]func(*gorm.DB) error(nil), m.prerequisites...)
	return versions, steps, prerequisites
}

type casbinRule struct {
	ID    uint   `gorm:"primaryKey;autoIncrement"`
	Ptype string `gorm:"size:100"`
	V0    string `gorm:"size:100"`
	V1    string `gorm:"size:100"`
	V2    string `gorm:"size:100"`
	V3    string `gorm:"size:100"`
	V4    string `gorm:"size:100"`
	V5    string `gorm:"size:100"`
}

func (casbinRule) TableName() string { return "casbin_rule" }

func validVersion(version string) bool {
	if len(version) != 13 {
		return false
	}
	for _, character := range version {
		if character < '0' || character > '9' {
			return false
		}
	}
	return true
}

// GetFilename extracts the 13-digit version prefix used by published files.
func GetFilename(path string) string {
	name := filepath.Base(path)
	if len(name) < 13 {
		return ""
	}
	return name[:13]
}
