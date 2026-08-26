package platform_test

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go-admin/internal/platform"
)

func TestLocalFileStoreKeepsFilesInsideItsRoot(t *testing.T) {
	store, err := platform.NewLocalFileStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewLocalFileStore: %v", err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("Close store: %v", err)
		}
	})

	ctx := context.Background()
	if err := store.Put(ctx, "uploads/report.txt", strings.NewReader("offline data")); err != nil {
		t.Fatalf("Put: %v", err)
	}
	file, err := store.Open(ctx, "uploads/report.txt")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	contents, readErr := io.ReadAll(file)
	closeErr := file.Close()
	if readErr != nil {
		t.Fatalf("ReadAll: %v", readErr)
	}
	if closeErr != nil {
		t.Fatalf("Close: %v", closeErr)
	}
	if got, want := string(contents), "offline data"; got != want {
		t.Fatalf("contents = %q, want %q", got, want)
	}
	if err := store.Put(ctx, "uploads/report.txt", strings.NewReader("replaced")); err != nil {
		t.Fatalf("replace Put: %v", err)
	}
	file, err = store.Open(ctx, "uploads/report.txt")
	if err != nil {
		t.Fatalf("Open replaced file: %v", err)
	}
	contents, readErr = io.ReadAll(file)
	closeErr = file.Close()
	if readErr != nil || closeErr != nil || string(contents) != "replaced" {
		t.Fatalf("replaced contents = %q, read=%v, close=%v; want replaced", contents, readErr, closeErr)
	}

	if err := store.Put(ctx, "../escaped.txt", strings.NewReader("escape")); !errors.Is(err, platform.ErrInvalidFileKey) {
		t.Fatalf("traversal error = %v, want ErrInvalidFileKey", err)
	}
	if _, err := store.Open(ctx, "/absolute.txt"); !errors.Is(err, platform.ErrInvalidFileKey) {
		t.Fatalf("absolute path error = %v, want ErrInvalidFileKey", err)
	}

	if err := store.Delete(ctx, "uploads/report.txt"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := store.Open(ctx, "uploads/report.txt"); !errors.Is(err, platform.ErrFileNotFound) {
		t.Fatalf("Open after Delete error = %v, want ErrFileNotFound", err)
	}
}

func TestLocalFileStoreRejectsSymbolicLinkEscape(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "root")
	outside := filepath.Join(base, "outside")
	if err := os.MkdirAll(outside, 0o700); err != nil {
		t.Fatalf("create outside directory: %v", err)
	}
	store, err := platform.NewLocalFileStore(root)
	if err != nil {
		t.Fatalf("NewLocalFileStore: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	if err := os.Symlink(outside, filepath.Join(root, "escape")); err != nil {
		t.Fatalf("create escape symlink: %v", err)
	}
	if err := store.Put(context.Background(), "escape/leaked.txt", strings.NewReader("leak")); !errors.Is(err, platform.ErrInvalidFileKey) {
		t.Fatalf("symlink escape error = %v, want ErrInvalidFileKey", err)
	}
	if _, err := os.Stat(filepath.Join(outside, "leaked.txt")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("outside file exists or stat failed unexpectedly: %v", err)
	}
}
