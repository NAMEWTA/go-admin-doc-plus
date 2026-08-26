package audit_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	audit "go-admin/internal/modules/audit"
	auditmigration "go-admin/internal/modules/audit/migrations/0011-audit"
	"go-admin/internal/platform/config"
	"go-admin/internal/platform/database"
	reliablemigration "go-admin/internal/platform/migrations/reliable-runtime"
	"go-admin/internal/platform/outbox"
)

const (
	auditE2EServeEnv   = "GO_ADMIN_AUDIT_E2E_SERVE"
	auditE2EProfileEnv = "GO_ADMIN_AUDIT_E2E_PROFILE"
	auditE2EReadyEnv   = "GO_ADMIN_AUDIT_E2E_READY_FILE"
	auditE2EStaticEnv  = "GO_ADMIN_AUDIT_E2E_STATIC_DIR"
)

// TestAuditUIHarnessServer is the tracked required-E2E host. Source gates compile it and skip;
// the Lead-owned runner opts in for both SQLite and PostgreSQL and drives the actual Web client
// and controller against this real Outbox-to-Audit process.
func TestAuditUIHarnessServer(t *testing.T) {
	if os.Getenv(auditE2EServeEnv) != "1" {
		if _, present := os.LookupEnv(auditPostgresEnv); present {
			t.Fatal("Audit UI harness compile self-check received PostgreSQL connection material")
		}
		t.Skip(auditE2EServeEnv + " is not enabled")
	}
	profile := os.Getenv(auditE2EProfileEnv)
	readyFile := os.Getenv(auditE2EReadyEnv)
	staticRoot := os.Getenv(auditE2EStaticEnv)
	if readyFile == "" || staticRoot == "" {
		t.Fatal("Audit UI harness paths are required")
	}
	if info, err := os.Stat(staticRoot); err != nil || !info.IsDir() {
		t.Fatal("Audit UI harness static directory is unavailable")
	}
	_, hasPostgres := os.LookupEnv(auditPostgresEnv)
	if profile == "sqlite" && hasPostgres {
		t.Fatal("SQLite Audit UI harness received PostgreSQL connection material")
	}
	if profile == "postgres" && !hasPostgres {
		t.Fatal("PostgreSQL Audit UI harness connection material is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	db := openAuditUIHarnessDatabase(t, ctx, profile)
	migrate(t, db, reliablemigration.Provider{}, auditmigration.Provider{})
	store := newAuditStore(t, db)
	now := time.Date(2026, 8, 27, 9, 0, 0, 0, time.UTC)
	recorder, err := audit.NewLoginRecorder(db)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error {
		_, recordErr := recorder.Record(ctx, tx, audit.LoginFact{Outcome: audit.OutcomeSucceeded, ActorType: audit.ActorAccount, ActorRef: stringPointer("account:account-00000011"), Source: audit.SourceWeb, OccurredAt: now.Add(-2 * 24 * time.Hour)})
		return recordErr
	}); err != nil {
		t.Fatal("seed Audit login fact failed")
	}
	enqueue(t, db, store, outbox.Event{ID: "audit-ui-event-001", Topic: audit.TopicOperationUpdated, BusinessKey: "resource:demo:ui-record-revision-2:account-00000011", Payload: []byte(`{"source":"web"}`), OccurredAt: now.Add(-60 * 24 * time.Hour)})
	dispatch(t, db, store, mustConsumers(t), now)

	service := mustServiceWithPolicy(t, db, allowAll{}, audit.RetentionPolicy{MinimumAge: 30 * 24 * time.Hour, CleanupLimit: 10, Now: func() time.Time { return now }})
	authorized, err := audit.NewAuthorizedRequest(audit.Principal{ID: "auditor-00000001"}, "ccccccccccccccccccccccccccccccccccccccccccc", stringPointer(sessionCookie("rrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrr")))
	if err != nil {
		t.Fatal(err)
	}
	api, err := audit.NewHTTPHandler(service, auditUIAuthorizer{authorized: authorized}, func(*http.Request) string { return "0123456789abcdef" })
	if err != nil {
		t.Fatal(err)
	}

	shutdown := make(chan struct{})
	var shutdownOnce sync.Once
	mux := http.NewServeMux()
	mux.Handle("/api/", http.StripPrefix("/api", api))
	mux.HandleFunc("/__test/snapshot", func(response http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			response.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var count int
		if err := db.Bun().QueryRowContext(request.Context(), "SELECT COUNT(*) FROM audit_facts").Scan(&count); err != nil {
			response.WriteHeader(http.StatusInternalServerError)
			return
		}
		response.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(response).Encode(map[string]int{"count": count})
	})
	mux.HandleFunc("/__test/shutdown", func(response http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			response.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		response.WriteHeader(http.StatusNoContent)
		shutdownOnce.Do(func() { close(shutdown) })
	})
	mux.Handle("/", http.FileServer(http.Dir(staticRoot)))
	server := httptest.NewUnstartedServer(mux)
	server.StartTLS()
	defer server.Close()
	if err := os.MkdirAll(filepath.Dir(readyFile), 0o700); err != nil {
		t.Fatal("Audit UI harness readiness directory is unavailable")
	}
	if err := os.WriteFile(readyFile, []byte(server.URL), 0o600); err != nil {
		t.Fatal("Audit UI harness readiness file is unavailable")
	}
	select {
	case <-shutdown:
	case <-ctx.Done():
		t.Fatal("Audit UI harness timed out")
	}
}

type auditUIAuthorizer struct{ authorized audit.AuthorizedRequest }

func (authorizer auditUIAuthorizer) AuthorizeRequest(_ context.Context, request *http.Request) (audit.AuthorizedRequest, audit.RequestFailure) {
	if request.Method != http.MethodGet && request.Header.Get("X-CSRF-Token") != "ccccccccccccccccccccccccccccccccccccccccccc" {
		return authorizer.authorized, audit.RequestAuthorizationFailed
	}
	return authorizer.authorized, audit.RequestAuthorized
}

func openAuditUIHarnessDatabase(t *testing.T, ctx context.Context, profile string) *database.Database {
	t.Helper()
	switch profile {
	case "sqlite":
		db, err := database.NewProcess().Open(ctx, database.Config{Profile: config.ProfileServerSQLite, SQLitePath: filepath.Join(t.TempDir(), "audit-ui.sqlite3")})
		if err != nil {
			t.Fatal("open SQLite Audit UI harness failed")
		}
		t.Cleanup(func() { _ = db.Close() })
		return db
	case "postgres":
		return openIsolatedAuditPostgres(t, ctx, os.Getenv(auditPostgresEnv))
	default:
		t.Fatal("Audit UI harness profile is invalid")
		return nil
	}
}
