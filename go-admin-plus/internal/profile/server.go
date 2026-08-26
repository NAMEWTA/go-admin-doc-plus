package profile

import (
	"context"
	"errors"
	"fmt"
	"strings"

	coreredis "github.com/go-admin-team/go-admin-core/v2/storage/redis"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"go-admin/internal/platform"
	"go-admin/internal/tenant"
)

var (
	// ErrInvalidServerConfig identifies a missing or inconsistent profile input.
	ErrInvalidServerConfig = errors.New("invalid server profile configuration")
	// ErrDependencyUnavailable omits connection strings and credentials by design.
	ErrDependencyUnavailable = errors.New("server profile dependency unavailable")
)

// ServerConfig defines the reference PostgreSQL, Redis, persistent FileStore
// and trusted host mapping profile. Connection secrets are never returned in errors.
type ServerConfig struct {
	PostgresDSN   string
	RedisURL      string
	FileRoot      string
	TenantHosts   map[string]string
	QueueGroup    string
	QueueConsumer string
	QueuePrefix   string
}

// BuildServer connects and probes every required adapter. A partial build is
// closed before an error is returned; the caller owns a successful result.
func BuildServer(ctx context.Context, config ServerConfig) (*platform.Dependencies, error) {
	if ctx == nil {
		return nil, fmt.Errorf("%w: build context is required", ErrInvalidServerConfig)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := validateServerConfig(config); err != nil {
		return nil, err
	}
	files, err := platform.NewLocalFileStore(config.FileRoot)
	if err != nil {
		return nil, fmt.Errorf("%w: file store", ErrDependencyUnavailable)
	}
	tenants, err := tenant.NewServerResolver(config.TenantHosts)
	if err != nil {
		files.Close()
		return nil, fmt.Errorf("%w: tenant resolver", ErrInvalidServerConfig)
	}

	database, err := gorm.Open(postgres.Open(config.PostgresDSN), &gorm.Config{
		DisableAutomaticPing: true,
		Logger:               logger.Default.LogMode(logger.Silent),
		NamingStrategy:       schema.NamingStrategy{SingularTable: true},
	})
	if err != nil {
		files.Close()
		return nil, dependencyError(ctx, "PostgreSQL")
	}
	sqlDatabase, err := database.DB()
	if err != nil {
		files.Close()
		return nil, dependencyError(ctx, "PostgreSQL")
	}
	if err := sqlDatabase.PingContext(ctx); err != nil {
		sqlDatabase.Close()
		files.Close()
		return nil, dependencyError(ctx, "PostgreSQL")
	}

	cache, err := coreredis.Open(ctx, config.RedisURL)
	if err != nil {
		sqlDatabase.Close()
		files.Close()
		return nil, dependencyError(ctx, "Redis cache")
	}
	queue, err := coreredis.OpenQueue(ctx, config.RedisURL, coreredis.QueueOptions{
		Group:     config.QueueGroup,
		Consumer:  config.QueueConsumer,
		KeyPrefix: config.QueuePrefix,
	})
	if err != nil {
		cache.Close()
		sqlDatabase.Close()
		files.Close()
		return nil, dependencyError(ctx, "Redis queue")
	}

	dependencies, err := platform.NewDependencies(platform.AdapterSet{
		Database: database,
		Cache:    cache,
		Queue:    queue,
		Files:    files,
		Tenants:  tenants,
	},
		platform.ResourceStopper{Name: "server file store", Stop: ignoreContext(files.Close)},
		platform.ResourceStopper{Name: "PostgreSQL database", Stop: ignoreContext(sqlDatabase.Close)},
		platform.ResourceStopper{Name: "Redis cache", Stop: ignoreContext(cache.Close)},
		platform.ResourceStopper{Name: "Redis queue", Stop: ignoreContext(queue.Close)},
	)
	if err != nil {
		queue.Close()
		cache.Close()
		sqlDatabase.Close()
		files.Close()
		return nil, fmt.Errorf("assemble server adapters: %w", err)
	}
	return dependencies, nil
}

func validateServerConfig(config ServerConfig) error {
	missing := make([]string, 0, 4)
	if strings.TrimSpace(config.PostgresDSN) == "" {
		missing = append(missing, "PostgreSQL DSN")
	}
	if strings.TrimSpace(config.RedisURL) == "" {
		missing = append(missing, "Redis URL")
	}
	if strings.TrimSpace(config.FileRoot) == "" {
		missing = append(missing, "file root")
	}
	if len(config.TenantHosts) == 0 {
		missing = append(missing, "tenant hosts")
	}
	if len(missing) != 0 {
		return fmt.Errorf("%w: missing %s", ErrInvalidServerConfig, strings.Join(missing, ", "))
	}
	return nil
}

func dependencyError(ctx context.Context, name string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("%w: %s: %w", ErrDependencyUnavailable, name, err)
	}
	return fmt.Errorf("%w: %s", ErrDependencyUnavailable, name)
}
