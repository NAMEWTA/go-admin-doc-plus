package profile_test

import (
	"context"
	"io"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	corestorage "github.com/go-admin-team/go-admin-core/v2/storage"

	"go-admin/internal/profile"
)

func TestBuildDesktopCreatesOfflineAdaptersAndSQLitePolicy(t *testing.T) {
	root := filepath.Join(t.TempDir(), "Go Admin Data")
	dependencies, layout, err := profile.BuildDesktop(context.Background(), profile.DesktopConfig{
		DataRoot: root,
	})
	if err != nil {
		t.Fatalf("BuildDesktop: %v", err)
	}
	t.Cleanup(func() {
		if err := dependencies.Close(context.Background()); err != nil {
			t.Errorf("Close: %v", err)
		}
	})

	for _, directory := range []string{layout.DatabaseDir, layout.FilesDir, layout.LogsDir, layout.BackupsDir, layout.TempDir} {
		info, err := os.Stat(directory)
		if err != nil {
			t.Fatalf("stat %s: %v", directory, err)
		}
		if !info.IsDir() {
			t.Fatalf("%s is not a directory", directory)
		}
	}
	if want := filepath.Join(layout.Root, "db", "go-admin.sqlite3"); layout.DatabasePath != want {
		t.Fatalf("DatabasePath = %q, want %q", layout.DatabasePath, want)
	}

	var foreignKeys int
	if err := dependencies.Database().Raw("PRAGMA foreign_keys").Scan(&foreignKeys).Error; err != nil {
		t.Fatalf("query foreign_keys: %v", err)
	}
	if foreignKeys != 1 {
		t.Fatalf("foreign_keys = %d, want 1", foreignKeys)
	}
	var busyTimeout int
	if err := dependencies.Database().Raw("PRAGMA busy_timeout").Scan(&busyTimeout).Error; err != nil {
		t.Fatalf("query busy_timeout: %v", err)
	}
	if busyTimeout != 5000 {
		t.Fatalf("busy_timeout = %d, want 5000", busyTimeout)
	}
	var journalMode string
	if err := dependencies.Database().Raw("PRAGMA journal_mode").Scan(&journalMode).Error; err != nil {
		t.Fatalf("query journal_mode: %v", err)
	}
	if strings.ToLower(journalMode) != "wal" {
		t.Fatalf("journal_mode = %q, want wal", journalMode)
	}

	request := httptest.NewRequest("GET", "http://anything.invalid/api/v1/menu", nil)
	if got, err := dependencies.Tenants().Resolve(request); err != nil || got != "local" {
		t.Fatalf("tenant Resolve = %q, %v; want local, nil", got, err)
	}
	if err := dependencies.Files().Put(context.Background(), "avatar.txt", strings.NewReader("local")); err != nil {
		t.Fatalf("Files.Put: %v", err)
	}
	file, err := dependencies.Files().Open(context.Background(), "avatar.txt")
	if err != nil {
		t.Fatalf("Files.Open: %v", err)
	}
	contents, err := io.ReadAll(file)
	file.Close()
	if err != nil || string(contents) != "local" {
		t.Fatalf("file contents = %q, %v; want local, nil", contents, err)
	}
}

func TestDesktopQueueStopsOnContextCancellation(t *testing.T) {
	dependencies, _, err := profile.BuildDesktop(context.Background(), profile.DesktopConfig{
		DataRoot: filepath.Join(t.TempDir(), "app-data"),
	})
	if err != nil {
		t.Fatalf("BuildDesktop: %v", err)
	}
	t.Cleanup(func() {
		if err := dependencies.Close(context.Background()); err != nil {
			t.Errorf("Close: %v", err)
		}
	})
	if err := dependencies.Queue().Subscribe("local", func(context.Context, corestorage.Message) error { return nil }); err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- dependencies.Queue().Start(ctx) }()
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Queue.Start: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("desktop queue did not stop after context cancellation")
	}
}
