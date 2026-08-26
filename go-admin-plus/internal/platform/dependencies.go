package platform

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	corestorage "github.com/go-admin-team/go-admin-core/v2/storage"
	"gorm.io/gorm"

	"go-admin/internal/tenant"
)

// AdapterSet is the host-independent infrastructure surface consumed by the
// application. The owning profile must provide every adapter.
type AdapterSet struct {
	Database *gorm.DB
	Cache    corestorage.Cache
	Queue    corestorage.Queue
	Files    FileStore
	Tenants  tenant.Resolver
}

// StopFunc releases one owned resource and must return promptly when its
// context is cancelled.
type StopFunc func(context.Context) error

// ResourceStopper names a resource for aggregated shutdown diagnostics.
type ResourceStopper struct {
	Name string
	Stop StopFunc
}

// Dependencies is the profile-owned adapter set. Close stops every owned
// resource in reverse acquisition order and returns a stable aggregate error.
type Dependencies struct {
	adapters  AdapterSet
	stoppers  []ResourceStopper
	closeOnce sync.Once
	closeErr  error
}

// NewDependencies validates a complete adapter set. Stoppers must be supplied
// in resource acquisition order so Close can safely reverse it.
func NewDependencies(adapters AdapterSet, stoppers ...ResourceStopper) (*Dependencies, error) {
	if adapters.Database == nil {
		return nil, errors.New("database adapter is required")
	}
	if adapters.Cache == nil {
		return nil, errors.New("cache adapter is required")
	}
	if adapters.Queue == nil {
		return nil, errors.New("queue adapter is required")
	}
	if adapters.Files == nil {
		return nil, errors.New("file store adapter is required")
	}
	if adapters.Tenants == nil {
		return nil, errors.New("tenant resolver is required")
	}
	owned := append([]ResourceStopper(nil), stoppers...)
	for index, stopper := range owned {
		if strings.TrimSpace(stopper.Name) == "" {
			return nil, fmt.Errorf("resource stopper at index %d has no name", index)
		}
		if stopper.Stop == nil {
			return nil, fmt.Errorf("resource stopper %q has no function", stopper.Name)
		}
	}
	return &Dependencies{adapters: adapters, stoppers: owned}, nil
}

// Database returns the profile database without transferring ownership.
func (d *Dependencies) Database() *gorm.DB { return d.adapters.Database }

// Cache returns the profile cache without transferring ownership.
func (d *Dependencies) Cache() corestorage.Cache { return d.adapters.Cache }

// Queue returns the profile queue without transferring ownership.
func (d *Dependencies) Queue() corestorage.Queue { return d.adapters.Queue }

// Files returns the profile file store without transferring ownership.
func (d *Dependencies) Files() FileStore { return d.adapters.Files }

// Tenants returns the profile tenant resolver without transferring ownership.
func (d *Dependencies) Tenants() tenant.Resolver { return d.adapters.Tenants }

// Close releases all resources once in reverse acquisition order. Repeated
// calls return the same aggregate result and do not retry resource shutdown.
func (d *Dependencies) Close(ctx context.Context) error {
	if ctx == nil {
		return errors.New("close context is required")
	}
	d.closeOnce.Do(func() {
		var stopErrors []error
		for index := len(d.stoppers) - 1; index >= 0; index-- {
			stopper := d.stoppers[index]
			if err := stopper.Stop(ctx); err != nil {
				stopErrors = append(stopErrors, fmt.Errorf("stop %s: %w", stopper.Name, err))
			}
		}
		d.closeErr = errors.Join(stopErrors...)
	})
	return d.closeErr
}
