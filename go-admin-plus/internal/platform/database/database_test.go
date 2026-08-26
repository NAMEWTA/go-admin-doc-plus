package database_test

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"go-admin/internal/platform/config"
	"go-admin/internal/platform/database"
)

func TestProcessOpensOneSQLiteDatabaseAndRunsTransactions(t *testing.T) {
	t.Parallel()

	cfg := database.Config{
		Profile:    config.ProfileServerSQLite,
		SQLitePath: filepath.Join(t.TempDir(), "app.db"),
	}
	process := database.NewProcess()
	db, err := process.Open(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})

	if got := db.Dialect(); got != database.DialectSQLite {
		t.Fatalf("Dialect() = %q, want %q", got, database.DialectSQLite)
	}
	if _, err := db.SQL().ExecContext(context.Background(), `CREATE TABLE records (id INTEGER PRIMARY KEY, value TEXT NOT NULL)`); err != nil {
		t.Fatalf("create table: %v", err)
	}
	if err := db.WithinTx(context.Background(), func(ctx context.Context, tx database.Tx) error {
		_, err := tx.ExecContext(ctx, `INSERT INTO records(value) VALUES (?)`, "committed")
		return err
	}); err != nil {
		t.Fatalf("committed transaction: %v", err)
	}
	wantRollback := errors.New("rollback")
	if err := db.WithinTx(context.Background(), func(ctx context.Context, tx database.Tx) error {
		if _, err := tx.ExecContext(ctx, `INSERT INTO records(value) VALUES (?)`, "rolled-back"); err != nil {
			return err
		}
		return wantRollback
	}); !errors.Is(err, wantRollback) {
		t.Fatalf("rolled-back transaction error = %v, want %v", err, wantRollback)
	}

	var count int
	if err := db.SQL().QueryRowContext(context.Background(), `SELECT COUNT(*) FROM records`).Scan(&count); err != nil {
		t.Fatalf("count records: %v", err)
	}
	if count != 1 {
		t.Fatalf("record count = %d, want 1", count)
	}
	if _, err := process.Open(context.Background(), cfg); !errors.Is(err, database.ErrAlreadyOpened) {
		t.Fatalf("second Open() error = %v, want ErrAlreadyOpened", err)
	}
}

func TestConfigNeverSerializesConnectionMaterial(t *testing.T) {
	t.Parallel()

	const (
		dsn  = "postgres://admin:secret@db.internal/app"
		path = "/private/desktop/application.db"
	)
	cfg := database.Config{Profile: config.ProfileServerPostgres, PostgresDSN: dsn, SQLitePath: path}

	encoded, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	var logBuffer strings.Builder
	logger := slog.New(slog.NewJSONHandler(&logBuffer, nil))
	logger.Info("database config", "config", cfg)

	for name, output := range map[string]string{
		"String": cfg.String(), "GoString": cfg.GoString(), "JSON": string(encoded), "slog": logBuffer.String(),
	} {
		if strings.Contains(output, dsn) || strings.Contains(output, path) || strings.Contains(output, "secret") {
			t.Fatalf("%s exposed connection material: %s", name, output)
		}
	}
}

func TestOpenRejectsInvalidProfilesWithoutExposingConnectionMaterial(t *testing.T) {
	t.Parallel()

	tests := []database.Config{
		{Profile: config.ProfileServerPostgres, PostgresDSN: "not a valid postgres dsn secret"},
		{Profile: config.ProfileDesktopSQLite},
		{Profile: config.Profile("unsupported"), SQLitePath: "/private/secret.db"},
	}
	for _, cfg := range tests {
		_, err := database.NewProcess().Open(context.Background(), cfg)
		if err == nil {
			t.Fatalf("Open(%s) unexpectedly succeeded", cfg.Profile)
		}
		if strings.Contains(err.Error(), "secret") || strings.Contains(err.Error(), "/private") {
			t.Fatalf("Open(%s) exposed connection material: %v", cfg.Profile, err)
		}
	}
}

func TestDialectForEveryFormalProfile(t *testing.T) {
	t.Parallel()

	for profile, want := range map[config.Profile]database.Dialect{
		config.ProfileServerPostgres: database.DialectPostgres,
		config.ProfileServerSQLite:   database.DialectSQLite,
		config.ProfileDesktopSQLite:  database.DialectSQLite,
	} {
		got, err := database.DialectForProfile(profile)
		if err != nil || got != want {
			t.Fatalf("DialectForProfile(%q) = (%q, %v), want (%q, nil)", profile, got, err, want)
		}
	}
}

func TestProcessPublishesExactlyOneDatabaseUnderConcurrentOpen(t *testing.T) {
	t.Parallel()

	process := database.NewProcess()
	cfg := database.Config{Profile: config.ProfileDesktopSQLite, SQLitePath: filepath.Join(t.TempDir(), "desktop.db")}
	const callers = 12
	start := make(chan struct{})
	results := make(chan struct {
		db  *database.Database
		err error
	}, callers)
	var ready sync.WaitGroup
	ready.Add(callers)
	for range callers {
		go func() {
			ready.Done()
			<-start
			db, err := process.Open(context.Background(), cfg)
			results <- struct {
				db  *database.Database
				err error
			}{db, err}
		}()
	}
	ready.Wait()
	close(start)

	var winner *database.Database
	alreadyOpened := 0
	for range callers {
		result := <-results
		switch {
		case result.err == nil:
			if winner != nil {
				t.Fatal("multiple concurrent Open calls succeeded")
			}
			winner = result.db
		case errors.Is(result.err, database.ErrAlreadyOpened):
			alreadyOpened++
		default:
			t.Fatalf("Open() unexpected error = %v", result.err)
		}
	}
	if winner == nil || alreadyOpened != callers-1 {
		t.Fatalf("winner = %v, ErrAlreadyOpened count = %d", winner != nil, alreadyOpened)
	}
	if err := winner.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}
