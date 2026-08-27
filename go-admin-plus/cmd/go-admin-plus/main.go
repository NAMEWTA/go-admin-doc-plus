package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go-admin/internal/app/product"
	"go-admin/internal/application/health"
	serverhost "go-admin/internal/host/server"
	"go-admin/internal/platform/config"
	"go-admin/internal/platform/database"
)

const version = "0.1.0-dev"

type commandOptions struct {
	profile        string
	configFile     string
	listen         string
	sqlitePath     string
	dataRoot       string
	repositoryRoot string
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "go-admin-plus server failed")
		os.Exit(1)
	}
}

func run(arguments []string) error {
	options, err := parseOptions(arguments)
	if err != nil {
		return err
	}
	profile := config.Profile(options.profile)
	if profile != config.ProfileServerSQLite && profile != config.ProfileServerPostgres {
		return errors.New("server profile is unsupported")
	}
	if profile == config.ProfileServerPostgres && options.sqlitePath != "" {
		return errors.New("sqlite path cannot be used by the postgres profile")
	}
	cli := map[string]string{}
	if options.listen != "" {
		cli["http.listen"] = options.listen
	}
	if options.sqlitePath != "" {
		cli["database.path"] = options.sqlitePath
	}
	snapshot, err := config.Load(config.Input{
		Profile: profile, File: options.configFile, Environment: environment(), CLI: cli,
	})
	if err != nil {
		return errors.New("server configuration failed")
	}
	dataRoot, err := canonicalDirectory(options.dataRoot)
	if err != nil {
		return errors.New("server data root failed")
	}
	repositoryRoot, err := filepath.Abs(options.repositoryRoot)
	if err != nil {
		return errors.New("server repository root failed")
	}
	repositoryRoot, err = filepath.EvalSymlinks(repositoryRoot)
	if err != nil {
		return errors.New("server repository root failed")
	}
	address, databaseConfig, sessionPolicy, generatorSchema, err := runtimeProfile(snapshot)
	if err != nil {
		return err
	}

	host, err := serverhost.New(serverhost.Config{
		Address: address, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second, ShutdownTimeout: 10 * time.Second,
		Capabilities: health.Capabilities{HostProfile: "server", Version: version},
	}, func(ctx context.Context) (serverhost.Runtime, error) {
		db, err := database.NewProcess().Open(ctx, databaseConfig)
		if err != nil {
			return serverhost.Runtime{}, errors.New("server database startup failed")
		}
		built, err := product.Build(ctx, db, product.Options{
			SessionPolicy:       sessionPolicy,
			FilesRoot:           filepath.Join(dataRoot, "files"),
			RepositoryRoot:      repositoryRoot,
			GeneratorOutputRoot: filepath.Join(dataRoot, "generated"),
			GeneratorSchema:     generatorSchema,
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
	if err != nil {
		return errors.New("server host configuration failed")
	}
	return host.Run(context.Background())
}

func parseOptions(arguments []string) (commandOptions, error) {
	set := flag.NewFlagSet("go-admin-plus", flag.ContinueOnError)
	set.SetOutput(os.Stderr)
	options := commandOptions{}
	set.StringVar(&options.profile, "profile", string(config.ProfileServerSQLite), "server-sqlite or server-postgres")
	set.StringVar(&options.configFile, "config", "", "typed profile JSON file")
	set.StringVar(&options.listen, "listen", "", "HTTP listen override")
	set.StringVar(&options.sqlitePath, "sqlite-path", "", "SQLite database path override")
	set.StringVar(&options.dataRoot, "data-root", ".go-admin-plus", "runtime data root")
	set.StringVar(&options.repositoryRoot, "repository-root", ".", "generator repository skeleton root")
	if err := set.Parse(arguments); err != nil {
		return commandOptions{}, err
	}
	if set.NArg() != 0 {
		return commandOptions{}, errors.New("unexpected positional arguments")
	}
	return options, nil
}

func environment() map[string]string {
	result := make(map[string]string)
	for _, entry := range os.Environ() {
		key, value, found := strings.Cut(entry, "=")
		if found && strings.HasPrefix(key, "GO_ADMIN_") {
			result[key] = value
		}
	}
	return result
}

func canonicalDirectory(path string) (string, error) {
	absolute, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(absolute, 0o700); err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(absolute)
}

func runtimeProfile(snapshot config.Snapshot) (string, database.Config, config.SessionPolicy, string, error) {
	if profile, ok := snapshot.ServerSQLite(); ok {
		return profile.HTTPListen(), database.Config{Profile: config.ProfileServerSQLite, SQLitePath: profile.DatabasePath()}, profile.SessionPolicy(), "main", nil
	}
	if profile, ok := snapshot.ServerPostgres(); ok {
		return profile.HTTPListen(), database.Config{Profile: config.ProfileServerPostgres, PostgresDSN: profile.DatabaseDSN()}, profile.SessionPolicy(), "public", nil
	}
	return "", database.Config{}, config.SessionPolicy{}, "", errors.New("server runtime profile is invalid")
}
