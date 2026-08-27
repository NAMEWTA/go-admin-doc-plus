package main

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunAppliesSQLiteMigrationsIdempotently(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "product.sqlite3")
	arguments := []string{"--profile", "server-sqlite", "--sqlite-path", databasePath}

	var first bytes.Buffer
	if err := run(context.Background(), arguments, map[string]string{}, &first); err != nil {
		t.Fatalf("first migration failed: %v", err)
	}
	if !strings.Contains(first.String(), "migration complete: applied=") || strings.Contains(first.String(), "applied=0 ") {
		t.Fatalf("first migration did not apply schema: %q", first.String())
	}

	var second bytes.Buffer
	if err := run(context.Background(), arguments, map[string]string{}, &second); err != nil {
		t.Fatalf("second migration failed: %v", err)
	}
	if !strings.Contains(second.String(), "applied=0 ") {
		t.Fatalf("second migration was not idempotent: %q", second.String())
	}
}

func TestRunRejectsDesktopAndMixedDatabaseProfiles(t *testing.T) {
	tests := [][]string{
		{"--profile", "desktop-sqlite"},
		{"--profile", "server-postgres", "--sqlite-path", "product.sqlite3"},
	}
	for _, arguments := range tests {
		if err := run(context.Background(), arguments, map[string]string{}, &bytes.Buffer{}); err == nil {
			t.Fatalf("expected arguments to fail: %v", arguments)
		}
	}
}

func TestParseOptionsRejectsPositionalArguments(t *testing.T) {
	if _, err := parseOptions([]string{"up"}); err == nil {
		t.Fatal("expected positional migration action to be rejected")
	}
}
