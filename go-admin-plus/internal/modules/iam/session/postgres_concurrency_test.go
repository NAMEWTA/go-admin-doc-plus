package session

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"go-admin/internal/modules/iam/account"
	sessionmigration "go-admin/internal/modules/iam/migrations/0010-session-schema"
	"go-admin/internal/platform/config"
	"go-admin/internal/platform/database"
	"go-admin/internal/platform/migrations"
)

const postgresDisposableDSNEnv = "GO_ADMIN_TEST_POSTGRES_DISPOSABLE_DSN"

type postgresLoginFactNoop struct{}

func (postgresLoginFactNoop) RecordLoginFact(context.Context, database.Tx, LoginFact) error {
	return nil
}

// TestPostgresGenerationFencesConcurrentRotationAndRevoke is intentionally gated.
// Lead runs it against an isolated real PostgreSQL database in candidate verification.
func TestPostgresGenerationFencesConcurrentRotationAndRevoke(t *testing.T) {
	dsn := os.Getenv(postgresDisposableDSNEnv)
	if dsn == "" {
		t.Skip(postgresDisposableDSNEnv + " is not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	adminDB := openPostgres(t, ctx, dsn)
	schema := fmt.Sprintf("iam_session_%d", time.Now().UnixNano())
	if _, err := adminDB.SQL().ExecContext(ctx, `CREATE SCHEMA `+schema); err != nil {
		t.Fatalf("create isolated schema: %v", err)
	}
	t.Cleanup(func() {
		_, _ = adminDB.SQL().ExecContext(context.Background(), `DROP SCHEMA IF EXISTS `+schema+` CASCADE`)
	})

	parsed, err := pgx.ParseConfig(dsn)
	if err != nil {
		t.Fatal("parse disposable PostgreSQL connection material")
	}
	if parsed.RuntimeParams == nil {
		parsed.RuntimeParams = make(map[string]string)
	}
	parsed.RuntimeParams["search_path"] = schema
	db := openPostgres(t, ctx, parsed.ConnString())
	runner, err := migrations.NewRunner(sessionmigration.Provider{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := runner.Up(ctx, db); err != nil {
		t.Fatalf("migrate isolated schema: %v", err)
	}

	now := time.Date(2026, 8, 27, 8, 0, 0, 0, time.UTC)
	hash, err := account.HashPassword("correct horse battery")
	if err != nil {
		t.Fatal(err)
	}
	repository := account.NewRepository(db.Dialect())
	if err := db.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error {
		return repository.Create(ctx, tx, account.Credential{
			Profile:      account.Profile{ID: "account-00000001", Username: "admin", DisplayName: "Administrator", Email: "admin@example.test"},
			PasswordHash: hash,
		}, now)
	}); err != nil {
		t.Fatalf("seed account: %v", err)
	}
	policy, err := config.NewSessionPolicy(2*time.Hour, 8*time.Hour, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	clock := now
	service, err := NewService(db, policy, WithClock(func() time.Time { return clock }), WithLoginFactPort(postgresLoginFactNoop{}))
	if err != nil {
		t.Fatal(err)
	}
	issued, err := service.Login(ctx, "admin", "correct horse battery")
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	clock = clock.Add(61 * time.Minute)

	rotationLocked := make(chan struct{})
	releaseRotation := make(chan struct{})
	revokeRequested := make(chan struct{})
	revokeLocked := make(chan struct{})
	var rotationOnce, releaseOnce, requestOnce, revokeOnce sync.Once
	release := func() { releaseOnce.Do(func() { close(releaseRotation) }) }
	t.Cleanup(release)
	service.lockProbe = func(point accountLockPoint) {
		switch point {
		case accountLockHeld:
			rotationOnce.Do(func() {
				close(rotationLocked)
				<-releaseRotation
			})
		case accountRevokeLockRequested:
			requestOnce.Do(func() { close(revokeRequested) })
		case accountRevokeLockHeld:
			revokeOnce.Do(func() { close(revokeLocked) })
		}
	}

	rotated := make(chan Issued, 1)
	rotationErr := make(chan error, 1)
	go func() {
		value, currentErr := service.Current(ctx, issued.Token)
		rotated <- value
		rotationErr <- currentErr
	}()
	select {
	case <-rotationLocked:
	case <-ctx.Done():
		t.Fatal("rotation did not enter the account-locked critical section")
	}

	revokeErr := make(chan error, 1)
	go func() { revokeErr <- service.RevokeAccount(ctx, issued.Profile.ID) }()
	select {
	case <-revokeRequested:
	case <-ctx.Done():
		t.Fatal("revoke did not request the account lock")
	}
	select {
	case <-revokeLocked:
		t.Fatal("revoke acquired the account lock while rotation still held it")
	default:
	}
	release()

	replacement := <-rotated
	if err := <-rotationErr; err != nil {
		t.Fatalf("locked rotation failed: %v", err)
	}
	if err := <-revokeErr; err != nil {
		t.Fatalf("queued revoke failed: %v", err)
	}
	select {
	case <-revokeLocked:
	default:
		t.Fatal("revoke did not acquire the account lock after rotation committed")
	}
	for _, token := range []string{issued.Token, replacement.Token} {
		if _, err := service.Current(ctx, token); !errors.Is(err, ErrAuthentication) {
			t.Fatal("generation fence allowed a token after concurrent PostgreSQL revoke")
		}
	}
}

func openPostgres(t *testing.T, ctx context.Context, dsn string) *database.Database {
	t.Helper()
	db, err := database.NewProcess().Open(ctx, database.Config{
		Profile:            config.ProfileServerPostgres,
		PostgresDSN:        dsn,
		MaxOpenConnections: 4,
		MaxIdleConnections: 4,
	})
	if err != nil {
		t.Fatalf("open disposable PostgreSQL database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}
