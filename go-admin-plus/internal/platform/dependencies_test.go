package platform_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	corestorage "github.com/go-admin-team/go-admin-core/v2/storage"
	cachememory "github.com/go-admin-team/go-admin-core/v2/storage/cache"
	queuememory "github.com/go-admin-team/go-admin-core/v2/storage/queue"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"go-admin/internal/platform"
	"go-admin/internal/tenant"
)

func TestDependenciesCloseAllResourcesInReverseOrder(t *testing.T) {
	database, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	files, err := platform.NewLocalFileStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewLocalFileStore: %v", err)
	}
	t.Cleanup(func() { _ = files.Close() })
	cache := cachememory.NewMemCache()
	queue := queuememory.NewMemQueue(1)

	firstErr := errors.New("database close failed")
	lastErr := errors.New("queue close failed")
	var order []string
	dependencies, err := platform.NewDependencies(platform.AdapterSet{
		Database: database,
		Cache:    cache,
		Queue:    queue,
		Files:    files,
		Tenants:  tenant.Fixed("local"),
	},
		platform.ResourceStopper{Name: "database", Stop: recordStop(&order, "database", firstErr)},
		platform.ResourceStopper{Name: "cache", Stop: recordStop(&order, "cache", nil)},
		platform.ResourceStopper{Name: "queue", Stop: recordStop(&order, "queue", lastErr)},
	)
	if err != nil {
		t.Fatalf("NewDependencies: %v", err)
	}

	closeErr := dependencies.Close(context.Background())
	if !errors.Is(closeErr, firstErr) || !errors.Is(closeErr, lastErr) {
		t.Fatalf("Close error = %v, want both resource errors", closeErr)
	}
	if want := []string{"queue", "cache", "database"}; !reflect.DeepEqual(order, want) {
		t.Fatalf("close order = %v, want %v", order, want)
	}
	if err := dependencies.Close(context.Background()); !errors.Is(err, firstErr) || !errors.Is(err, lastErr) {
		t.Fatalf("second Close error = %v, want stable aggregate", err)
	}
	if want := []string{"queue", "cache", "database"}; !reflect.DeepEqual(order, want) {
		t.Fatalf("second Close repeated resources: %v", order)
	}

	var _ corestorage.Cache = dependencies.Cache()
	var _ corestorage.Queue = dependencies.Queue()
}

func recordStop(order *[]string, name string, result error) platform.StopFunc {
	return func(context.Context) error {
		*order = append(*order, name)
		return result
	}
}
