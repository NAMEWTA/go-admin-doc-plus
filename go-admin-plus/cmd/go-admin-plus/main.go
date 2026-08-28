package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/app/product"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/platform/config"
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
	host, err := product.NewServerHost(product.ServerLaunch{
		Snapshot:       snapshot,
		DataRoot:       options.dataRoot,
		RepositoryRoot: options.repositoryRoot,
		Version:        version,
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
