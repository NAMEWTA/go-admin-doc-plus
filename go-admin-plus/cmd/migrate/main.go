package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/app/product"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/platform/config"
)

type commandOptions struct {
	profile    string
	configFile string
	sqlitePath string
}

func main() {
	if err := run(context.Background(), os.Args[1:], environment(), os.Stdout); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "go-admin-plus migration failed")
		os.Exit(1)
	}
}

func run(ctx context.Context, arguments []string, environmentValues map[string]string, output io.Writer) error {
	options, err := parseOptions(arguments)
	if err != nil {
		return err
	}
	profile := config.Profile(options.profile)
	if profile != config.ProfileServerSQLite && profile != config.ProfileServerPostgres {
		return errors.New("migration profile is unsupported")
	}
	if profile == config.ProfileServerPostgres && options.sqlitePath != "" {
		return errors.New("sqlite path cannot be used by the postgres profile")
	}
	cli := map[string]string{}
	if options.sqlitePath != "" {
		cli["database.path"] = options.sqlitePath
	}
	snapshot, err := config.Load(config.Input{
		Profile: profile, File: options.configFile, Environment: environmentValues, CLI: cli,
	})
	if err != nil {
		return errors.New("migration configuration failed")
	}
	result, err := product.Migrate(ctx, snapshot)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(output, "migration complete: applied=%d current_version=%d\n", result.Applied, result.CurrentVersion); err != nil {
		return errors.New("migration result output failed")
	}
	return nil
}

func parseOptions(arguments []string) (commandOptions, error) {
	set := flag.NewFlagSet("migrate", flag.ContinueOnError)
	set.SetOutput(io.Discard)
	options := commandOptions{}
	set.StringVar(&options.profile, "profile", string(config.ProfileServerSQLite), "server-sqlite or server-postgres")
	set.StringVar(&options.configFile, "config", "", "typed profile JSON file")
	set.StringVar(&options.sqlitePath, "sqlite-path", "", "SQLite database path override")
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
