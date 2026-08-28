package product

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/application/health"
	serverhost "github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/host/server"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/platform/config"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/platform/database"
)

// ServerLaunch contains the typed launch material needed by the Server
// composition root. Process argument and environment parsing remain in cmd.
type ServerLaunch struct {
	Snapshot       config.Snapshot
	DataRoot       string
	RepositoryRoot string
	Version        string
}

type serverProfile struct {
	address            string
	database           database.Config
	sessionPolicy      config.SessionPolicy
	generatorSchema    string
	databaseCapability string
}

// NewServerHost composes the Server profile, database, product and network
// host without exposing concrete runtime dependencies to the command entry.
func NewServerHost(launch ServerLaunch) (*serverhost.Host, error) {
	launch.Version = strings.TrimSpace(launch.Version)
	if launch.Version == "" || strings.TrimSpace(launch.DataRoot) == "" || strings.TrimSpace(launch.RepositoryRoot) == "" {
		return nil, errors.New("server launch material is invalid")
	}
	profile, err := serverRuntimeProfile(launch.Snapshot)
	if err != nil {
		return nil, err
	}
	dataRoot, err := canonicalServerDataRoot(launch.DataRoot)
	if err != nil {
		return nil, errors.New("server data root failed")
	}
	repositoryRoot, err := canonicalServerRepositoryRoot(launch.RepositoryRoot)
	if err != nil {
		return nil, errors.New("server repository root failed")
	}

	return serverhost.New(serverhost.Config{
		Address:         profile.address,
		ReadTimeout:     15 * time.Second,
		WriteTimeout:    30 * time.Second,
		ShutdownTimeout: 10 * time.Second,
		Capabilities: health.Capabilities{
			Profile:  string(launch.Snapshot.Profile()),
			Version:  launch.Version,
			Database: profile.databaseCapability,
		},
	}, func(ctx context.Context) (serverhost.Runtime, error) {
		db, err := database.NewProcess().Open(ctx, profile.database)
		if err != nil {
			return serverhost.Runtime{}, errors.New("server database startup failed")
		}
		built, err := Build(ctx, db, Options{
			SessionPolicy:       profile.sessionPolicy,
			FilesRoot:           filepath.Join(dataRoot, "files"),
			RepositoryRoot:      repositoryRoot,
			GeneratorOutputRoot: filepath.Join(dataRoot, "generated"),
			GeneratorSchema:     profile.generatorSchema,
			GeneratorTables:     []string{"demo_products"},
			WorkerOwner:         fmt.Sprintf("server-%d", os.Getpid()),
			WorkerInterval:      time.Second,
			AuditRetentionAge:   30 * 24 * time.Hour,
		})
		if err != nil {
			_ = db.Close()
			return serverhost.Runtime{}, err
		}
		return serverhost.Runtime{
			Application: built.Application,
			Readiness:   built.Readiness,
			Close:       func(context.Context) error { return db.Close() },
		}, nil
	})
}

func serverRuntimeProfile(snapshot config.Snapshot) (serverProfile, error) {
	if profile, ok := snapshot.ServerSQLite(); ok {
		return serverProfile{
			address:            profile.HTTPListen(),
			database:           database.Config{Profile: config.ProfileServerSQLite, SQLitePath: profile.DatabasePath()},
			sessionPolicy:      profile.SessionPolicy(),
			generatorSchema:    "main",
			databaseCapability: "sqlite",
		}, nil
	}
	if profile, ok := snapshot.ServerPostgres(); ok {
		return serverProfile{
			address:            profile.HTTPListen(),
			database:           database.Config{Profile: config.ProfileServerPostgres, PostgresDSN: profile.DatabaseDSN()},
			sessionPolicy:      profile.SessionPolicy(),
			generatorSchema:    "public",
			databaseCapability: "postgres",
		}, nil
	}
	return serverProfile{}, errors.New("server runtime profile is invalid")
}

func canonicalServerDataRoot(root string) (string, error) {
	absolute, err := filepath.Abs(filepath.Clean(root))
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(absolute, 0o700); err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(absolute)
}

func canonicalServerRepositoryRoot(root string) (string, error) {
	absolute, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(absolute)
}
