package desktop

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/go-admin-team/go-admin-core/v2/captcha"
	mycasbin "github.com/go-admin-team/go-admin-core/v2/casbin"
	corelogger "github.com/go-admin-team/go-admin-core/v2/logger"
	"github.com/go-admin-team/go-admin-core/v2/sdk"
	"github.com/go-admin-team/go-admin-core/v2/sdk/api"
	coreconfig "github.com/go-admin-team/go-admin-core/v2/sdk/config"
	"github.com/go-admin-team/go-admin-core/v2/sdk/pkg"
	corestorage "github.com/go-admin-team/go-admin-core/v2/storage"
	"gorm.io/gorm"

	"go-admin/cmd/migrate/migration"
	_ "go-admin/cmd/migrate/migration/version"
	_ "go-admin/cmd/migrate/migration/version-local"
	common "go-admin/common/middleware"
	"go-admin/internal/application"
	"go-admin/internal/modules"
	desktopplatform "go-admin/internal/platform/desktop"
	"go-admin/internal/profile"
)

type RuntimeConfig struct {
	DataRoot string
	Name     string
	Mode     string
}

func NewRuntimeBuilder(config RuntimeConfig) Builder {
	return func(ctx context.Context) (result Runtime, resultErr error) {
		instanceLock, err := desktopplatform.AcquireInstanceLock(config.DataRoot)
		if err != nil {
			return Runtime{}, err
		}
		lockOwned := true
		defer func() {
			if lockOwned {
				resultErr = errors.Join(resultErr, instanceLock.Close())
			}
		}()

		coreconfig.DatabaseConfig.Driver = "sqlite3"
		var migrationResult migration.Result
		upgrade, err := profile.UpgradeDesktop(ctx, profile.DesktopConfig{DataRoot: config.DataRoot}, func(migrationCtx context.Context, database *gorm.DB) error {
			migrationResult, err = migration.Migrate.Run(migrationCtx, database)
			return err
		})
		if err != nil {
			if upgrade.BackupPath != "" {
				return Runtime{}, fmt.Errorf("upgrade desktop database; recovery backup %q: %w", upgrade.BackupPath, err)
			}
			return Runtime{}, fmt.Errorf("upgrade desktop database: %w", err)
		}

		dependencies, layout, err := profile.BuildDesktop(ctx, profile.DesktopConfig{DataRoot: config.DataRoot})
		if err != nil {
			return Runtime{}, err
		}
		closeOnFailure := true
		var logOwner *desktopLogOwner
		defer func() {
			if closeOnFailure {
				resultErr = errors.Join(
					resultErr,
					dependencies.Close(context.Background()),
					logOwner.Close(),
				)
			}
		}()

		logOwner, err = configureDesktop(config, layout)
		if err != nil {
			return Runtime{}, err
		}
		log.Printf(
			"desktop database migration completed: applied=%d skipped=%d backup=%q",
			len(migrationResult.Applied),
			len(migrationResult.Skipped),
			upgrade.BackupPath,
		)
		enforcer := installDesktopAdapters(dependencies)
		engine, err := buildDesktopEngine()
		if err != nil {
			enforcer.StopAutoLoadPolicy()
			return Runtime{}, err
		}
		app, err := application.Build(
			application.Config{Name: coreconfig.ApplicationConfig.Name},
			application.Dependencies{Handler: engine},
			modules.Default(),
		)
		if err != nil {
			enforcer.StopAutoLoadPolicy()
			return Runtime{}, fmt.Errorf("build desktop application: %w", err)
		}

		closeOnFailure = false
		lockOwned = false
		return Runtime{
			Application: app,
			Close: func(closeCtx context.Context) error {
				enforcer.StopAutoLoadPolicy()
				dependenciesErr := dependencies.Close(closeCtx)
				loggerErr := logOwner.Close()
				lockErr := instanceLock.Close()
				return errors.Join(dependenciesErr, loggerErr, lockErr)
			},
		}, nil
	}
}

type desktopLogOwner struct {
	previous corelogger.Logger
	current  corelogger.Logger
	closer   interface{ Close() error }
	once     sync.Once
	err      error
}

func (owner *desktopLogOwner) Close() error {
	if owner == nil {
		return nil
	}
	owner.once.Do(func() {
		if corelogger.DefaultLogger == owner.current {
			corelogger.DefaultLogger = owner.previous
		}
		owner.err = owner.closer.Close()
	})
	return owner.err
}

func configureDesktop(config RuntimeConfig, layout profile.DesktopLayout) (*desktopLogOwner, error) {
	name := strings.TrimSpace(config.Name)
	if name == "" {
		name = "go-admin-desktop"
	}
	mode := strings.ToLower(strings.TrimSpace(config.Mode))
	if mode == "" {
		mode = "dev"
	}
	coreconfig.ApplicationConfig.Name = name
	coreconfig.ApplicationConfig.Mode = mode
	coreconfig.ApplicationConfig.Host = "127.0.0.1"
	coreconfig.ApplicationConfig.Port = 0
	coreconfig.ApplicationConfig.EnableDP = false
	coreconfig.DatabaseConfig.Driver = "sqlite3"
	coreconfig.DatabaseConfig.Source = layout.DatabasePath
	secret, err := generateLaunchToken()
	if err != nil {
		return nil, fmt.Errorf("generate desktop JWT secret: %w", err)
	}
	coreconfig.JwtConfig.Secret = secret
	coreconfig.JwtConfig.Timeout = 3600
	logPath := filepath.Join(layout.LogsDir, "app.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, fmt.Errorf("prepare desktop log file: %w", err)
	}
	if err := logFile.Close(); err != nil {
		return nil, fmt.Errorf("prepare desktop log file: %w", err)
	}

	coreconfig.LoggerConfig.Type = ""
	coreconfig.LoggerConfig.Adapter = "zap"
	coreconfig.LoggerConfig.Path = layout.LogsDir
	coreconfig.LoggerConfig.Level = "info"
	coreconfig.LoggerConfig.Stdout = "file"
	coreconfig.LoggerConfig.Encoder = "json"
	coreconfig.LoggerConfig.EnabledDB = false
	previousLogger := corelogger.DefaultLogger
	level, err := corelogger.GetLevel(coreconfig.LoggerConfig.Level)
	if err != nil {
		return nil, fmt.Errorf("configure desktop log level: %w", err)
	}
	// Desktop owns the file handle, while the shared config Setup API has no close hook.
	configuredLogger := corelogger.NewZapLogger(
		corelogger.WithName("go-admin"),
		corelogger.WithCallerSkipCount(2),
		corelogger.WithLevel(level),
		corelogger.WithStdout(false),
		corelogger.WithPath(logPath),
		corelogger.WithEncoder(coreconfig.LoggerConfig.Encoder),
		corelogger.WithEnableCaller(coreconfig.LoggerConfig.EnableCaller),
	)
	corelogger.DefaultLogger = configuredLogger
	closer, ok := configuredLogger.(interface{ Close() error })
	if !ok {
		corelogger.DefaultLogger = previousLogger
		return nil, errors.New("desktop logger does not support close")
	}
	return &desktopLogOwner{
		previous: previousLogger,
		current:  configuredLogger,
		closer:   closer,
	}, nil
}

func installDesktopAdapters(dependencies interface {
	Database() *gorm.DB
	Cache() corestorage.Cache
	Queue() corestorage.Queue
}) interface{ StopAutoLoadPolicy() } {
	enforcer := mycasbin.Setup(dependencies.Database(), "")
	sdk.Runtime.SetDbByTenant("*", dependencies.Database())
	sdk.Runtime.SetCasbinByTenant("*", enforcer)
	legacyCache := corestorage.LegacyAdapter(dependencies.Cache())
	sdk.Runtime.SetCacheAdapter(legacyCache)
	sdk.Runtime.SetQueueAdapter(corestorage.LegacyQueueAdapter(dependencies.Queue()))
	captcha.SetStore(captcha.NewCacheStore(legacyCache, 600))
	return enforcer
}

func buildDesktopEngine() (*gin.Engine, error) {
	if coreconfig.ApplicationConfig.Mode == pkg.ModeProd.String() {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.New()
	sdk.Runtime.SetEngine(engine)
	engine.Use(common.Sentinel()).
		Use(common.RequestId(pkg.TrafficKey)).
		Use(api.SetRequestLogger)
	common.InitMiddleware(engine)
	return engine, nil
}
