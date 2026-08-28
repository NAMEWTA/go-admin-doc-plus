package main

import (
	"path/filepath"
	"testing"

	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/platform/config"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/platform/database"
)

func TestParseOptionsKeepsServerProfilesExplicit(t *testing.T) {
	options, err := parseOptions([]string{"--profile", "server-postgres", "--listen", "127.0.0.1:9090", "--repository-root", "."})
	if err != nil {
		t.Fatal(err)
	}
	if options.profile != "server-postgres" || options.listen != "127.0.0.1:9090" {
		t.Fatalf("options = %#v", options)
	}
	if _, err := parseOptions([]string{"unexpected"}); err == nil {
		t.Fatal("parseOptions accepted positional input")
	}
}

func TestRuntimeProfileMapsOnlyServerDatabaseModes(t *testing.T) {
	sqlitePath := filepath.Join(t.TempDir(), "server.sqlite3")
	sqlite, err := config.Load(config.Input{Profile: config.ProfileServerSQLite, CLI: map[string]string{"database.path": sqlitePath}})
	if err != nil {
		t.Fatal(err)
	}
	address, db, _, schema, err := runtimeProfile(sqlite)
	if err != nil {
		t.Fatal(err)
	}
	if address != "127.0.0.1:8080" || db.Profile != config.ProfileServerSQLite || db.SQLitePath != sqlitePath || schema != "main" {
		t.Fatalf("SQLite runtime = address %q config %#v schema %q", address, db, schema)
	}

	desktop, err := config.Load(config.Input{Profile: config.ProfileDesktopSQLite, Desktop: config.DesktopMaterial{
		DataDirectory: filepath.Join(t.TempDir(), "data"), LogDirectory: filepath.Join(t.TempDir(), "logs"), LoopbackPort: 4321, StartupToken: "0123456789abcdef0123456789abcdef",
	}})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, _, _, err := runtimeProfile(desktop); err == nil {
		t.Fatal("runtimeProfile accepted desktop profile")
	}
}

func TestRuntimeProfileMapsPostgresWithoutExposingItToSQLite(t *testing.T) {
	snapshot, err := config.Load(config.Input{Profile: config.ProfileServerPostgres, Environment: map[string]string{"GO_ADMIN_DATABASE_DSN": "postgres://operator:secret@localhost/product"}})
	if err != nil {
		t.Fatal(err)
	}
	_, db, _, schema, err := runtimeProfile(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	if db.Profile != config.ProfileServerPostgres || db.PostgresDSN == "" || db.SQLitePath != "" || schema != "public" {
		t.Fatalf("PostgreSQL mapping = profile %q sqlite %q schema %q", db.Profile, db.SQLitePath, schema)
	}
	if dialect, err := database.DialectForProfile(db.Profile); err != nil || dialect != database.DialectPostgres {
		t.Fatalf("PostgreSQL dialect = %q, %v", dialect, err)
	}
}
