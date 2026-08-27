package product_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"go-admin/internal/app/product"
	"go-admin/internal/platform/config"
	"go-admin/internal/platform/coordination"
	"go-admin/internal/platform/database"
)

func TestBuildAssemblesEveryHTTPModuleAndWorkerLifecycle(t *testing.T) {
	ctx := context.Background()
	db, err := database.NewProcess().Open(ctx, database.Config{
		Profile:    config.ProfileServerSQLite,
		SQLitePath: filepath.Join(t.TempDir(), "runtime.sqlite3"),
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	repositoryRoot, err := filepath.Abs("../../../..")
	if err != nil {
		t.Fatal(err)
	}
	dataRoot, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := product.Build(ctx, db, product.Options{
		SessionPolicy:       config.DefaultSessionPolicy(),
		FilesRoot:           filepath.Join(dataRoot, "files"),
		RepositoryRoot:      repositoryRoot,
		GeneratorOutputRoot: filepath.Join(dataRoot, "generated"),
		GeneratorSchema:     "main",
		GeneratorTables:     []string{"demo_products"},
		WorkerOwner:         "product-test-worker",
		WorkerInterval:      time.Hour,
		AuditRetentionAge:   30 * 24 * time.Hour,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := runtime.Application.Start(ctx); err != nil {
		t.Fatal(err)
	}

	paths := []string{
		"/api/iam/session/current",
		"/api/iam/administration/manifest",
		"/api/audit/records?page=1&pageSize=20",
		"/api/organization/departments",
		"/api/settings/values?category=business&page=1&pageSize=20",
		"/api/generator/tables",
		"/api/scheduler/task-types",
		"/api/demo/products?page=1&pageSize=20",
		"/api/files/objects?page=1&pageSize=20",
	}
	for _, path := range paths {
		response := httptest.NewRecorder()
		runtime.Application.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusUnauthorized {
			t.Errorf("GET %s status = %d, want authentication rejection", path, response.Code)
		}
	}

	if err := runtime.Application.Stop(ctx); err != nil {
		t.Fatal(err)
	}
	lease, err := coordination.Acquire(ctx, db, coordination.Config{Owner: "product-test-after-stop"})
	if err != nil {
		t.Fatalf("worker lease remained owned after stop: %v", err)
	}
	if err := lease.Close(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestBuildRejectsIncompleteProductOptionsBeforeAssembly(t *testing.T) {
	if _, err := product.Build(context.Background(), nil, product.Options{}); err == nil {
		t.Fatal("Build accepted incomplete dependencies")
	}
}
