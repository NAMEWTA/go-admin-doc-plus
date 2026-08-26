//go:build !bindings

package main

import (
	"path/filepath"
	"testing"
)

func TestDesktopDataRootAllowsAnExplicitDevelopmentOverride(t *testing.T) {
	override := filepath.Join(t.TempDir(), "isolated")
	t.Setenv("GO_ADMIN_DESKTOP_DATA_ROOT", override)

	root, err := desktopDataRoot()
	if err != nil {
		t.Fatalf("desktopDataRoot: %v", err)
	}
	if root != override {
		t.Fatalf("root = %q, want %q", root, override)
	}
}
