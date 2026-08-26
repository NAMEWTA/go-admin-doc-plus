package config_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	runtimeconfig "go-admin/internal/platform/config"
)

func TestLoadAppliesDocumentedPrecedence(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "runtime.json")
	if err := os.WriteFile(configPath, []byte(`{
  "profile": "server-sqlite",
  "http": {"listen": "127.0.0.1:8100"},
  "log": {"level": "debug"},
  "database": {"path": "from-file.sqlite3"}
}`), 0o600); err != nil {
		t.Fatal(err)
	}

	snapshot, err := runtimeconfig.Load(runtimeconfig.Input{
		Profile: runtimeconfig.ProfileServerSQLite,
		File:    configPath,
		Environment: map[string]string{
			"GO_ADMIN_HTTP_LISTEN": "127.0.0.1:8200",
			"GO_ADMIN_LOG_LEVEL":   "warn",
			"GO_ADMIN_SQLITE_PATH": "from-env.sqlite3",
		},
		CLI: map[string]string{
			"http.listen":   "127.0.0.1:8300",
			"database.path": "from-cli.sqlite3",
		},
	})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if snapshot.Profile() != runtimeconfig.ProfileServerSQLite {
		t.Fatalf("Profile() = %q", snapshot.Profile())
	}
	server, ok := snapshot.ServerSQLite()
	if !ok {
		t.Fatal("ServerSQLite() did not return the selected profile")
	}
	if server.HTTPListen() != "127.0.0.1:8300" {
		t.Fatalf("HTTPListen() = %q", server.HTTPListen())
	}
	if server.LogLevel() != "warn" {
		t.Fatalf("LogLevel() = %q", server.LogLevel())
	}
	if server.DatabasePath() != "from-cli.sqlite3" {
		t.Fatalf("DatabasePath() = %q", server.DatabasePath())
	}
}

func TestLoadAppliesDocumentedPrecedenceToPostgres(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "runtime.json")
	if err := os.WriteFile(configPath, []byte(`{
  "profile": "server-postgres",
  "http": {"listen": "127.0.0.1:8100"},
  "log": {"level": "debug"}
}`), 0o600); err != nil {
		t.Fatal(err)
	}
	snapshot, err := runtimeconfig.Load(runtimeconfig.Input{
		Profile: runtimeconfig.ProfileServerPostgres,
		File:    configPath,
		Environment: map[string]string{
			"GO_ADMIN_DATABASE_DSN": "postgres://runtime-secret",
			"GO_ADMIN_HTTP_LISTEN":  "127.0.0.1:8200",
			"GO_ADMIN_LOG_LEVEL":    "warn",
		},
		CLI: map[string]string{
			"http.listen": "127.0.0.1:8300",
			"log.level":   "error",
		},
	})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	postgres, ok := snapshot.ServerPostgres()
	if !ok || postgres.HTTPListen() != "127.0.0.1:8300" || postgres.LogLevel() != "error" {
		t.Fatalf("ServerPostgres() = %#v", postgres)
	}
}

func TestLoadResolvesPostgresSecretWithoutLeakingDiagnostics(t *testing.T) {
	secretPath := filepath.Join(t.TempDir(), "database-secret")
	secret := "postgres://operator:correct-horse@example.invalid/product"
	if err := os.WriteFile(secretPath, []byte(secret+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	snapshot, err := runtimeconfig.Load(runtimeconfig.Input{
		Profile: runtimeconfig.ProfileServerPostgres,
		Environment: map[string]string{
			"GO_ADMIN_DATABASE_DSN_FILE": secretPath,
		},
	})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	postgres, ok := snapshot.ServerPostgres()
	if !ok || postgres.DatabaseDSN() != secret {
		t.Fatal("ServerPostgres() did not resolve the secret file")
	}

	_, err = runtimeconfig.Load(runtimeconfig.Input{
		Profile: runtimeconfig.ProfileServerPostgres,
		Environment: map[string]string{
			"GO_ADMIN_DATABASE_DSN":      secret,
			"GO_ADMIN_DATABASE_DSN_FILE": secretPath,
		},
	})
	if err == nil {
		t.Fatal("Load() accepted conflicting secret sources")
	}
	for _, forbidden := range []string{secret, secretPath, "correct-horse"} {
		if strings.Contains(err.Error(), forbidden) {
			t.Fatalf("Load() error leaked %q: %v", forbidden, err)
		}
	}
	if !strings.Contains(err.Error(), "database.dsn") || !strings.Contains(err.Error(), "conflicting") {
		t.Fatalf("Load() error = %q, want field and rule", err)
	}
}

func TestLoadRejectsUnknownAndCrossProfileFields(t *testing.T) {
	for name, document := range map[string]string{
		"unknown root":        `{"profile":"server-sqlite","extraField":"do-not-print"}`,
		"unknown nested":      `{"profile":"server-sqlite","database":{"path":"data.sqlite3","extraScope":"do-not-print"}}`,
		"postgres-only field": `{"profile":"server-postgres","database":{"path":"/private/database.sqlite3"}}`,
	} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "runtime.json")
			if err := os.WriteFile(path, []byte(document), 0o600); err != nil {
				t.Fatal(err)
			}
			input := runtimeconfig.Input{Profile: runtimeconfig.ProfileServerSQLite, File: path}
			if name == "postgres-only field" {
				input.Profile = runtimeconfig.ProfileServerPostgres
				input.Environment = map[string]string{"GO_ADMIN_DATABASE_DSN": "postgres://private"}
			}
			_, err := runtimeconfig.Load(input)
			if err == nil {
				t.Fatal("Load() accepted a field outside the profile schema")
			}
			if !strings.Contains(err.Error(), "unknown field") {
				t.Fatalf("Load() error = %q, want unknown-field rule", err)
			}
			for _, forbidden := range []string{"do-not-print", "/private/database.sqlite3", path} {
				if strings.Contains(err.Error(), forbidden) {
					t.Fatalf("Load() error leaked %q: %v", forbidden, err)
				}
			}
		})
	}
}

func TestLoadDesktopUsesOnlyNativeHostMaterial(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "desktop.json")
	if err := os.WriteFile(configPath, []byte(`{
  "profile": "desktop-sqlite",
  "log": {"level": "debug"}
}`), 0o600); err != nil {
		t.Fatal(err)
	}
	dataDirectory := filepath.Join(t.TempDir(), "data")
	logDirectory := filepath.Join(t.TempDir(), "logs")
	const startupToken = "0123456789abcdef0123456789abcdef"

	snapshot, err := runtimeconfig.Load(runtimeconfig.Input{
		Profile: runtimeconfig.ProfileDesktopSQLite,
		File:    configPath,
		Environment: map[string]string{
			"GO_ADMIN_LOG_LEVEL": "warn",
		},
		CLI: map[string]string{"log.level": "error"},
		Desktop: runtimeconfig.DesktopMaterial{
			DataDirectory: dataDirectory,
			LogDirectory:  logDirectory,
			LoopbackPort:  43127,
			StartupToken:  startupToken,
		},
	})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	desktop, ok := snapshot.DesktopSQLite()
	if !ok {
		t.Fatal("DesktopSQLite() did not return the selected profile")
	}
	if desktop.LogLevel() != "error" || desktop.DataDirectory() != dataDirectory || desktop.LogDirectory() != logDirectory {
		t.Fatalf("DesktopSQLite() = %#v", desktop)
	}
	if desktop.LoopbackAddress() != "127.0.0.1:43127" || desktop.StartupToken() != startupToken {
		t.Fatal("DesktopSQLite() did not preserve native launch material")
	}
}

func TestLoadRejectsInvalidSourcesBeforeStartup(t *testing.T) {
	secretPath := filepath.Join(t.TempDir(), "missing-secret")
	profilePath := filepath.Join(t.TempDir(), "wrong-profile.json")
	if err := os.WriteFile(profilePath, []byte(`{"profile":"server-postgres"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name      string
		input     runtimeconfig.Input
		wantField string
		wantRule  string
		forbidden string
	}{
		{
			name:      "unknown environment field",
			input:     runtimeconfig.Input{Profile: runtimeconfig.ProfileServerSQLite, Environment: map[string]string{"GO_ADMIN_UNEXPECTED_SCOPE": "private-scope"}},
			wantField: "GO_ADMIN_UNEXPECTED_SCOPE",
			wantRule:  "unknown field",
			forbidden: "private-scope",
		},
		{
			name:      "secret CLI field",
			input:     runtimeconfig.Input{Profile: runtimeconfig.ProfileServerPostgres, Environment: map[string]string{"GO_ADMIN_DATABASE_DSN": "postgres://safe"}, CLI: map[string]string{"database.dsn": "postgres://leaked"}},
			wantField: "database.dsn",
			wantRule:  "not permitted",
			forbidden: "postgres://leaked",
		},
		{
			name:      "missing secret",
			input:     runtimeconfig.Input{Profile: runtimeconfig.ProfileServerPostgres},
			wantField: "database.dsn",
			wantRule:  "required",
		},
		{
			name:      "unreadable secret",
			input:     runtimeconfig.Input{Profile: runtimeconfig.ProfileServerPostgres, Environment: map[string]string{"GO_ADMIN_DATABASE_DSN_FILE": secretPath}},
			wantField: "database.dsn",
			wantRule:  "unreadable",
			forbidden: secretPath,
		},
		{
			name:      "profile conflict",
			input:     runtimeconfig.Input{Profile: runtimeconfig.ProfileServerSQLite, File: profilePath},
			wantField: "profile",
			wantRule:  "conflicts",
			forbidden: profilePath,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := runtimeconfig.Load(test.input)
			if err == nil {
				t.Fatal("Load() accepted invalid input")
			}
			if !strings.Contains(err.Error(), test.wantField) || !strings.Contains(err.Error(), test.wantRule) {
				t.Fatalf("Load() error = %q, want field %q and rule %q", err, test.wantField, test.wantRule)
			}
			if test.forbidden != "" && strings.Contains(err.Error(), test.forbidden) {
				t.Fatalf("Load() error leaked %q: %v", test.forbidden, err)
			}
		})
	}
}

func TestSnapshotIsOwnedAndSafeToFormat(t *testing.T) {
	const secret = "postgres://admin:never-log-this@example.invalid/product"
	environment := map[string]string{"GO_ADMIN_DATABASE_DSN": secret}
	snapshot, err := runtimeconfig.Load(runtimeconfig.Input{
		Profile:     runtimeconfig.ProfileServerPostgres,
		Environment: environment,
	})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	environment["GO_ADMIN_DATABASE_DSN"] = "mutated"
	postgres, ok := snapshot.ServerPostgres()
	if !ok || postgres.DatabaseDSN() != secret {
		t.Fatal("Snapshot changed after mutating its input")
	}
	for _, formatted := range []string{
		fmt.Sprint(runtimeconfig.Input{Profile: runtimeconfig.ProfileServerPostgres, Environment: environment}),
		fmt.Sprintf("%#v", runtimeconfig.Input{Profile: runtimeconfig.ProfileServerPostgres, Environment: environment}),
		fmt.Sprintf("%#v", runtimeconfig.DesktopMaterial{DataDirectory: "/private/data", StartupToken: secret}),
		fmt.Sprint(snapshot),
		fmt.Sprintf("%#v", snapshot),
		fmt.Sprint(postgres),
		fmt.Sprintf("%#v", postgres),
	} {
		if strings.Contains(formatted, secret) || strings.Contains(formatted, "never-log-this") || strings.Contains(formatted, "mutated") || strings.Contains(formatted, "/private/data") {
			t.Fatalf("formatted snapshot leaked secret: %s", formatted)
		}
		if !strings.Contains(formatted, "redacted") {
			t.Fatalf("formatted snapshot = %q, want redaction marker", formatted)
		}
	}
}

func TestLoadValidatesTypedValuesWithoutEchoingThem(t *testing.T) {
	dataDirectory := filepath.Join(t.TempDir(), "same-directory")
	tests := []struct {
		name      string
		input     runtimeconfig.Input
		wantField string
		forbidden string
	}{
		{
			name: "invalid listen address",
			input: runtimeconfig.Input{
				Profile: runtimeconfig.ProfileServerSQLite,
				CLI:     map[string]string{"http.listen": "not-an-address-secret"},
			},
			wantField: "http.listen",
			forbidden: "not-an-address-secret",
		},
		{
			name: "invalid log level",
			input: runtimeconfig.Input{
				Profile:     runtimeconfig.ProfileServerPostgres,
				Environment: map[string]string{"GO_ADMIN_DATABASE_DSN": "postgres://private"},
				CLI:         map[string]string{"log.level": "verbose-secret"},
			},
			wantField: "log.level",
			forbidden: "verbose-secret",
		},
		{
			name: "desktop directory conflict",
			input: runtimeconfig.Input{
				Profile: runtimeconfig.ProfileDesktopSQLite,
				Desktop: runtimeconfig.DesktopMaterial{
					DataDirectory: dataDirectory,
					LogDirectory:  dataDirectory,
					LoopbackPort:  41001,
					StartupToken:  "0123456789abcdef0123456789abcdef",
				},
			},
			wantField: "desktop.logDirectory",
			forbidden: dataDirectory,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := runtimeconfig.Load(test.input)
			if err == nil {
				t.Fatal("Load() accepted an invalid typed value")
			}
			if !strings.Contains(err.Error(), test.wantField) {
				t.Fatalf("Load() error = %q, want field %q", err, test.wantField)
			}
			if strings.Contains(err.Error(), test.forbidden) {
				t.Fatalf("Load() error leaked %q: %v", test.forbidden, err)
			}
		})
	}
}
