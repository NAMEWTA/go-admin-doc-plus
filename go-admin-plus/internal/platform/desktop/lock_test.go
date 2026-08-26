package desktop

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestInstanceLockAllowsOnlyOneWriterAndCanBeReacquired(t *testing.T) {
	root := filepath.Join(t.TempDir(), "app-data")
	first, err := AcquireInstanceLock(root)
	if err != nil {
		t.Fatalf("Acquire first lock: %v", err)
	}
	second, err := AcquireInstanceLock(root)
	if !errors.Is(err, ErrInstanceLocked) {
		t.Fatalf("Acquire second lock error = %v, want ErrInstanceLocked", err)
	}
	if second != nil {
		t.Fatal("second lock is non-nil")
	}
	if err := first.Close(); err != nil {
		t.Fatalf("Close first lock: %v", err)
	}

	reacquired, err := AcquireInstanceLock(root)
	if err != nil {
		t.Fatalf("reacquire lock: %v", err)
	}
	if err := reacquired.Close(); err != nil {
		t.Fatalf("Close reacquired lock: %v", err)
	}
	if err := reacquired.Close(); err != nil {
		t.Fatalf("second Close: %v", err)
	}

	info, err := os.Stat(filepath.Join(root, instanceLockName))
	if err != nil {
		t.Fatalf("stat lock file: %v", err)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm()&0o077 != 0 {
		t.Fatalf("lock permissions = %o, want private", info.Mode().Perm())
	}
}
