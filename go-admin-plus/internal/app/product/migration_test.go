package product

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/platform/config"
)

func TestMigrateAppliesSQLiteProductSchemaIdempotently(t *testing.T) {
	snapshot, err := config.Load(config.Input{
		Profile: config.ProfileServerSQLite,
		CLI:     map[string]string{"database.path": filepath.Join(t.TempDir(), "product.sqlite3")},
	})
	if err != nil {
		t.Fatal(err)
	}
	first, err := Migrate(context.Background(), snapshot)
	if err != nil {
		t.Fatalf("first migration: %v", err)
	}
	if first.Applied == 0 || first.CurrentVersion == 0 {
		t.Fatalf("first migration result = %#v", first)
	}
	second, err := Migrate(context.Background(), snapshot)
	if err != nil {
		t.Fatalf("second migration: %v", err)
	}
	if second.Applied != 0 || second.CurrentVersion != first.CurrentVersion {
		t.Fatalf("second migration result = %#v, first = %#v", second, first)
	}
}

func TestMigrateRejectsNonServerProfileAndNilContext(t *testing.T) {
	desktop, err := config.Load(config.Input{Profile: config.ProfileDesktopSQLite, Desktop: config.DesktopMaterial{
		DataDirectory: filepath.Join(t.TempDir(), "data"), LogDirectory: filepath.Join(t.TempDir(), "logs"),
		LoopbackPort: 4321, StartupToken: "0123456789abcdef0123456789abcdef",
	}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Migrate(context.Background(), desktop); err == nil {
		t.Fatal("Migrate accepted a Desktop profile")
	}
	if _, err := Migrate(nil, config.Snapshot{}); err == nil {
		t.Fatal("Migrate accepted a nil context")
	}
}
