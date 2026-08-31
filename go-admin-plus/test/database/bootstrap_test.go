package database_test

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/app/product"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/iam/account"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/platform/config"
	productdatabase "github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/platform/database"
)

const bootstrapPassword = "administrator password"

func TestSystemAdminBootstrapSQLite(t *testing.T) {
	ctx := context.Background()
	db, err := databaseProcess(t).Open(ctx, productdatabase.Config{
		Profile:    config.ProfileServerSQLite,
		SQLitePath: filepath.Join(t.TempDir(), "bootstrap.db"),
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	runner, err := product.NewMigrationRunner()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := runner.Up(ctx, db); err != nil {
		t.Fatal(err)
	}

	script := bootstrapScript(t, "sqlite")
	for range 2 {
		if _, err := db.SQL().ExecContext(ctx, script); err != nil {
			t.Fatalf("execute bootstrap: %v", err)
		}
	}

	var accounts, assignments int
	var passwordHash string
	if err := db.SQL().QueryRowContext(ctx, `SELECT COUNT(*), MIN(password_hash) FROM iam_accounts WHERE username = 'admin'`).Scan(&accounts, &passwordHash); err != nil {
		t.Fatal(err)
	}
	if err := db.SQL().QueryRowContext(ctx, `SELECT COUNT(*) FROM iam_account_roles WHERE account_id = 'account-system-admin' AND role_id = 'role-system-admin'`).Scan(&assignments); err != nil {
		t.Fatal(err)
	}
	if accounts != 1 || assignments != 1 {
		t.Fatalf("bootstrap result = %d accounts, %d assignments", accounts, assignments)
	}
	if !account.VerifyPassword(passwordHash, bootstrapPassword) {
		t.Fatal("bootstrap password hash does not match the documented credential")
	}
}

func TestSystemAdminBootstrapDialectsStayAligned(t *testing.T) {
	hashPattern := regexp.MustCompile(`\$argon2id\$v=19\$m=65536,t=3,p=4\$[A-Za-z0-9+/]+\$[A-Za-z0-9+/]+`)
	sqliteScript := bootstrapScript(t, "sqlite")
	postgresScript := bootstrapScript(t, "postgres")
	sqliteHash := hashPattern.FindString(sqliteScript)
	postgresHash := hashPattern.FindString(postgresScript)
	if sqliteHash == "" || sqliteHash != postgresHash {
		t.Fatal("bootstrap password hashes differ between dialects")
	}
	for dialect, script := range map[string]string{"sqlite": sqliteScript, "postgres": postgresScript} {
		for _, contract := range []string{"account-system-admin", "role-system-admin", "ON CONFLICT(username) DO NOTHING", "ON CONFLICT(account_id, role_id) DO NOTHING"} {
			if !strings.Contains(script, contract) {
				t.Fatalf("%s bootstrap is missing %q", dialect, contract)
			}
		}
	}
}

func databaseProcess(t *testing.T) *productdatabase.Process {
	t.Helper()
	return productdatabase.NewProcess()
}

func bootstrapScript(t *testing.T, dialect string) string {
	t.Helper()
	_, source, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("bootstrap test source path unavailable")
	}
	path := filepath.Join(filepath.Dir(source), "..", "..", "..", "database", "bootstrap", dialect, "001-system-admin.sql")
	payload, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(payload)
}
