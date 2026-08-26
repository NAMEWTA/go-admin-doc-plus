package profile

import (
	"net/url"
	"testing"
)

func TestDesktopSQLiteDSN(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		wantPath string
	}{
		{
			name:     "unix absolute path",
			path:     "/tmp/Go Admin/go-admin.sqlite3",
			wantPath: "/tmp/Go Admin/go-admin.sqlite3",
		},
		{
			name:     "Windows drive path",
			path:     "D:/a/Go Admin/go-admin.sqlite3",
			wantPath: "/D:/a/Go Admin/go-admin.sqlite3",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dsn := desktopSQLiteDSN(test.path)
			parsed, err := url.Parse(dsn)
			if err != nil {
				t.Fatalf("parse DSN %q: %v", dsn, err)
			}
			if parsed.Scheme != "file" {
				t.Fatalf("scheme = %q, want file", parsed.Scheme)
			}
			if parsed.Host != "" {
				t.Fatalf("host = %q, want empty", parsed.Host)
			}
			if parsed.Path != test.wantPath {
				t.Fatalf("path = %q, want %q", parsed.Path, test.wantPath)
			}
			query := parsed.Query()
			if query.Get("_foreign_keys") != "on" {
				t.Fatalf("_foreign_keys = %q, want on", query.Get("_foreign_keys"))
			}
			if query.Get("_busy_timeout") != "5000" {
				t.Fatalf("_busy_timeout = %q, want 5000", query.Get("_busy_timeout"))
			}
			if query.Get("_journal_mode") != "WAL" {
				t.Fatalf("_journal_mode = %q, want WAL", query.Get("_journal_mode"))
			}
		})
	}
}
