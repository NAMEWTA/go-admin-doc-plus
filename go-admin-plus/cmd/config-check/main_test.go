package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunValidatesWithoutExposingConfiguration(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{
		"--profile", "server-sqlite",
		"--http-listen", "127.0.0.1:8123",
		"--sqlite-path", "/private/example.sqlite3",
	}, map[string]string{"GO_ADMIN_LOG_LEVEL": "warn"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run() code = %d, stderr = %q", code, stderr.String())
	}
	if stdout.String() != "{\"profile\":\"server-sqlite\",\"status\":\"valid\"}\n" {
		t.Fatalf("stdout = %q", stdout.String())
	}
	if strings.Contains(stdout.String(), "8123") || strings.Contains(stdout.String(), "example.sqlite3") {
		t.Fatalf("stdout leaked configuration: %s", stdout.String())
	}
}

func TestRunRejectsSecretFlagsAndRedactsSecretFileFailure(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"--profile", "server-postgres", "--database-dsn", "postgres://do-not-print"}, nil, &stdout, &stderr)
	if code != 2 || strings.Contains(stderr.String(), "postgres://do-not-print") {
		t.Fatalf("secret flag result code=%d stderr=%q", code, stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	secretPath := "/private/missing/database-secret"
	code = run([]string{"--profile", "server-postgres"}, map[string]string{
		"GO_ADMIN_DATABASE_DSN_FILE": secretPath,
	}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("unreadable secret result code=%d stderr=%q", code, stderr.String())
	}
	if strings.Contains(stderr.String(), secretPath) || !strings.Contains(stderr.String(), "database.dsn") {
		t.Fatalf("unreadable secret stderr = %q", stderr.String())
	}
}
