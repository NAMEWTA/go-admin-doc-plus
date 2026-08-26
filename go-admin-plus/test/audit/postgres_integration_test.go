package audit_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	audit "go-admin/internal/modules/audit"
	auditmigration "go-admin/internal/modules/audit/migrations/0011-audit"
	"go-admin/internal/platform/config"
	"go-admin/internal/platform/database"
	reliablemigration "go-admin/internal/platform/migrations/reliable-runtime"
	"go-admin/internal/platform/outbox"
)

const auditPostgresEnv = "GO_ADMIN_TEST_POSTGRES_DISPOSABLE_DSN"

func TestAuditPostgresMigrationProjectionQueryAndCleanup(t *testing.T) {
	dsn := os.Getenv(auditPostgresEnv)
	if dsn == "" {
		t.Skip("set " + auditPostgresEnv + " to run the PostgreSQL Audit integration")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	db := openIsolatedAuditPostgres(t, ctx, dsn)
	migrate(t, db, reliablemigration.Provider{}, auditmigration.Provider{})
	migrate(t, db, reliablemigration.Provider{}, auditmigration.Provider{})
	if _, err := db.Bun().ExecContext(ctx, `INSERT INTO audit_facts (topic, business_key, actor_ref, payload, occurred_at) VALUES (?, ?, ?, ?, ?)`, audit.TopicLoginSucceeded, "login:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "account:x", []byte(`{"actorType":"account","source":"web"}`), time.Now().UTC()); err == nil {
		t.Fatal("PostgreSQL accepted a short non-opaque actor reference")
	}
	if _, err := db.Bun().ExecContext(ctx, `INSERT INTO audit_facts (topic, business_key, payload, occurred_at) VALUES (?, ?, ?, ?)`, audit.TopicOperationUpdated, "evil:bad space:x:y:system", []byte(`{"source":"web"}`), time.Now().UTC()); err == nil {
		t.Fatal("PostgreSQL accepted an ambiguous operation actor")
	}
	store := newAuditStore(t, db)
	now := time.Date(2026, 8, 27, 8, 0, 0, 0, time.UTC)
	enqueue(t, db, store, outbox.Event{ID: "audit-pg-event-001", Topic: audit.TopicOperationUpdated, BusinessKey: "resource:demo:record-pg-revision-2:account-00000009", Payload: []byte(`{"source":"server"}`), OccurredAt: now.Add(-60 * 24 * time.Hour)})
	dispatch(t, db, store, mustConsumers(t), now)
	if _, err := db.Bun().ExecContext(ctx, "UPDATE audit_facts SET payload = ? WHERE topic = ? AND business_key = ?", []byte(`{"source":"server"}`), audit.TopicOperationUpdated, "resource:demo:record-pg-revision-2:account-00000009"); err != nil {
		t.Fatal("update PostgreSQL payload fixture failed")
	}
	service := mustServiceWithPolicy(t, db, allowAll{}, audit.RetentionPolicy{MinimumAge: 30 * 24 * time.Hour, CleanupLimit: 10, Now: func() time.Time { return now }})
	page, err := service.List(ctx, audit.Principal{ID: "auditor-00000001"}, audit.Filter{Page: 1, PageSize: 20, Source: audit.SourceServer})
	if err != nil || page.Total != 1 || page.Records[0].Subject != "demo:record-pg-revision-2" || page.Records[0].ActorRef == nil || *page.Records[0].ActorRef != "account:account-00000009" {
		t.Fatalf("PostgreSQL Audit page = %#v, %v", page, err)
	}
	result, err := service.Cleanup(ctx, audit.Principal{ID: "auditor-00000001"}, audit.CleanupCommand{Before: now.Add(-45 * 24 * time.Hour), Confirmation: audit.CleanupConfirmation})
	if err != nil || result.Deleted != 1 {
		t.Fatalf("PostgreSQL Audit cleanup = %#v, %v", result, err)
	}
}

func openIsolatedAuditPostgres(t *testing.T, ctx context.Context, dsn string) *database.Database {
	t.Helper()
	admin, err := database.NewProcess().Open(ctx, database.Config{Profile: config.ProfileServerPostgres, PostgresDSN: dsn, MaxOpenConnections: 2, MaxIdleConnections: 2})
	if err != nil {
		t.Fatal("open PostgreSQL Audit administrator failed")
	}
	schema := fmt.Sprintf("audit_test_%d", time.Now().UnixNano())
	if _, err := admin.SQL().ExecContext(ctx, `CREATE SCHEMA `+schema); err != nil {
		_ = admin.Close()
		t.Fatal("create PostgreSQL Audit schema failed")
	}
	parsed, err := pgx.ParseConfig(dsn)
	if err != nil {
		cleanupAuditPostgres(t, admin, schema)
		t.Fatal("parse PostgreSQL Audit connection failed")
	}
	if parsed.RuntimeParams == nil {
		parsed.RuntimeParams = make(map[string]string)
	}
	parsed.RuntimeParams["search_path"] = schema
	db, err := database.NewProcess().Open(ctx, database.Config{Profile: config.ProfileServerPostgres, PostgresDSN: parsed.ConnString(), MaxOpenConnections: 4, MaxIdleConnections: 4})
	if err != nil {
		cleanupAuditPostgres(t, admin, schema)
		t.Fatal("open isolated PostgreSQL Audit database failed")
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error("close isolated PostgreSQL Audit database failed")
		}
		cleanupAuditPostgres(t, admin, schema)
	})
	return db
}

func cleanupAuditPostgres(t *testing.T, admin *database.Database, schema string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := admin.SQL().ExecContext(ctx, `DROP SCHEMA IF EXISTS `+schema+` CASCADE`); err != nil {
		t.Error("drop PostgreSQL Audit schema failed")
	}
	if err := admin.Close(); err != nil {
		t.Error("close PostgreSQL Audit administrator failed")
	}
}
