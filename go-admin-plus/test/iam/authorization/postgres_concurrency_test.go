package authorization_test

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"go-admin/internal/modules/iam/account"
	"go-admin/internal/modules/iam/administration"
	"go-admin/internal/modules/iam/authorization"
	sessionmigration "go-admin/internal/modules/iam/migrations/0010-session-schema"
	administrationmigration "go-admin/internal/modules/iam/migrations/0020-administration-schema"
	"go-admin/internal/platform/config"
	"go-admin/internal/platform/database"
	"go-admin/internal/platform/migrations"
)

const postgresDSNEnvironment = "GO_ADMIN_TEST_POSTGRES_DISPOSABLE_DSN"

// TestPostgresRevocationWaitsForFinalAuthorizationFence is default-skipped because it creates and
// drops an isolated schema. Lead runs it against the disposable candidate PostgreSQL profile.
func TestPostgresRevocationWaitsForFinalAuthorizationFence(t *testing.T) {
	dsn := os.Getenv(postgresDSNEnvironment)
	if dsn == "" {
		t.Skip(postgresDSNEnvironment + " is not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()
	setup := openPostgresAuthorizationDB(t, ctx, withApplicationName(t, dsn, "t07_setup"))
	schema := "t07_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err := setup.SQL().ExecContext(ctx, `CREATE SCHEMA `+schema); err != nil {
		t.Fatal("create isolated schema failed")
	}
	t.Cleanup(func() {
		cleanup, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		if _, err := setup.SQL().ExecContext(cleanup, `DROP SCHEMA IF EXISTS `+schema+` CASCADE`); err != nil {
			t.Errorf("isolated schema cleanup failed")
		}
	})
	serviceDB := openPostgresAuthorizationDB(t, ctx, withSearchPath(t, withApplicationName(t, dsn, "t07_service"), schema))
	revokerDB := openPostgresAuthorizationDB(t, ctx, withSearchPath(t, withApplicationName(t, dsn, "t07_revoker"), schema))
	runner, err := migrations.NewRunner(sessionmigration.Provider{}, administrationmigration.Provider{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := runner.Up(ctx, serviceDB); err != nil {
		t.Fatal("isolated migrations failed")
	}
	hash, err := account.HashPassword("administrator password")
	if err != nil {
		t.Fatal(err)
	}
	repository := account.NewRepository(database.DialectPostgres)
	if err := serviceDB.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error {
		return repository.Create(ctx, tx, account.Credential{Profile: account.Profile{ID: adminID, Username: "admin", DisplayName: "Administrator", Email: "admin@example.test"}, PasswordHash: hash}, time.Now().UTC())
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := serviceDB.Bun().ExecContext(ctx, `INSERT INTO iam_account_roles(account_id, role_id) VALUES (?, ?)`, adminID, "role-system-admin"); err != nil {
		t.Fatal(err)
	}

	locked, release := make(chan struct{}), make(chan struct{})
	service, err := administration.NewService(serviceDB, authorization.NewService(serviceDB), administration.WithAuthorizationProbe(func(permission string) {
		if permission == authorization.PermissionRolesWrite {
			close(locked)
			<-release
		}
	}))
	if err != nil {
		t.Fatal(err)
	}
	command := make(chan error, 1)
	go func() {
		_, commandErr := service.CreateRole(ctx, adminID, "fenced-role", "Fenced role", authorization.ScopeAll)
		command <- commandErr
	}()
	<-locked
	revoked := make(chan error, 1)
	go func() {
		_, revokeErr := revokerDB.Bun().ExecContext(ctx, `DELETE FROM iam_role_permissions WHERE role_id = ? AND permission_code = ?`, "role-system-admin", authorization.PermissionRolesWrite)
		revoked <- revokeErr
	}()
	waitForPostgresLock(t, ctx, setup, "t07_revoker")
	close(release)
	if err := <-command; err != nil {
		t.Fatalf("already-authorized command failed: %v", err)
	}
	if err := <-revoked; err != nil {
		t.Fatal("revocation failed")
	}
	plain, _ := administration.NewService(serviceDB, authorization.NewService(serviceDB))
	if _, err := plain.CreateRole(ctx, adminID, "after-revoke", "After revoke", authorization.ScopeAll); !errors.Is(err, authorization.ErrDenied) {
		t.Fatalf("post-revoke command = %v", err)
	}
}

func waitForPostgresLock(t *testing.T, ctx context.Context, db *database.Database, applicationName string) {
	t.Helper()
	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()
	for {
		var waiting bool
		err := db.SQL().QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM pg_stat_activity WHERE application_name = $1 AND wait_event_type = 'Lock')`, applicationName).Scan(&waiting)
		if err != nil {
			t.Fatal("lock observation failed")
		}
		if waiting {
			return
		}
		select {
		case <-ctx.Done():
			t.Fatal("revocation never reached database lock wait")
		case <-ticker.C:
		}
	}
}

func openPostgresAuthorizationDB(t *testing.T, ctx context.Context, dsn string) *database.Database {
	t.Helper()
	db, err := database.NewProcess().Open(ctx, database.Config{Profile: config.ProfileServerPostgres, PostgresDSN: dsn})
	if err != nil {
		t.Fatal("postgres open failed")
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}
func withApplicationName(t *testing.T, dsn, name string) string {
	t.Helper()
	parsed, err := url.Parse(dsn)
	if err != nil {
		t.Fatal("postgres DSN is invalid")
	}
	query := parsed.Query()
	query.Set("application_name", name)
	parsed.RawQuery = query.Encode()
	return parsed.String()
}
func withSearchPath(t *testing.T, dsn, schema string) string {
	t.Helper()
	if !strings.HasPrefix(schema, "t07_") {
		t.Fatal("isolated schema name is invalid")
	}
	parsed, err := url.Parse(dsn)
	if err != nil {
		t.Fatal("postgres DSN is invalid")
	}
	query := parsed.Query()
	query.Set("search_path", schema)
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

func TestPostgresHarnessDiagnosticsDoNotContainDSN(t *testing.T) {
	const secret = "postgres://user:password@private.example/db"
	output := fmt.Sprintf("%v", database.Config{Profile: config.ProfileServerPostgres, PostgresDSN: secret})
	if strings.Contains(output, "password") || strings.Contains(output, "private.example") {
		t.Fatal("database config exposed DSN")
	}
}
