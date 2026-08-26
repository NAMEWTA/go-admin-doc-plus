package migrate

import (
	"context"
	"fmt"
	"io"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	corestorage "github.com/go-admin-team/go-admin-core/v2/storage"

	"go-admin/cmd/migrate/migration"
	"go-admin/internal/profile"
)

func TestServerProfileAndPublishedMigrationsOnPostgres(t *testing.T) {
	postgresDSN := os.Getenv("GO_ADMIN_POSTGRES_TEST_DSN")
	redisURL := os.Getenv("GO_ADMIN_REDIS_TEST_URL")
	if postgresDSN == "" || redisURL == "" {
		t.Skip("GO_ADMIN_POSTGRES_TEST_DSN and GO_ADMIN_REDIS_TEST_URL must point to disposable services")
	}
	restoreWorkingDirectory(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	dependencies, err := profile.BuildServer(ctx, profile.ServerConfig{
		PostgresDSN: postgresDSN,
		RedisURL:    redisURL,
		FileRoot:    t.TempDir(),
		TenantHosts: map[string]string{"admin.integration.test": "main"},
		QueuePrefix: fmt.Sprintf("go-admin-test-%d:", time.Now().UnixNano()),
	})
	if err != nil {
		t.Fatalf("BuildServer: %v", err)
	}
	t.Cleanup(func() {
		if err := dependencies.Close(context.Background()); err != nil {
			t.Errorf("Close: %v", err)
		}
	})

	first, err := migration.Migrate.Run(ctx, dependencies.Database())
	if err != nil {
		t.Fatalf("first PostgreSQL migration run: %v", err)
	}
	if len(first.Applied) == 0 {
		t.Fatal("disposable PostgreSQL database was not empty")
	}
	second, err := migration.Migrate.Run(ctx, dependencies.Database())
	if err != nil {
		t.Fatalf("second PostgreSQL migration run: %v", err)
	}
	if len(second.Applied) != 0 || !reflect.DeepEqual(second.Skipped, first.Applied) {
		t.Fatalf("second migration result = %#v, want all %v skipped", second, first.Applied)
	}

	if err := dependencies.Cache().Set(ctx, "profile-contract", "redis", time.Minute); err != nil {
		t.Fatalf("Cache.Set: %v", err)
	}
	if value, err := dependencies.Cache().Get(ctx, "profile-contract"); err != nil || value != "redis" {
		t.Fatalf("Cache.Get = %q, %v; want redis, nil", value, err)
	}
	delivered := make(chan corestorage.Message, 1)
	if err := dependencies.Queue().Subscribe("profile-contract", func(_ context.Context, message corestorage.Message) error {
		delivered <- message
		return nil
	}); err != nil {
		t.Fatalf("Queue.Subscribe: %v", err)
	}
	queueCtx, stopQueue := context.WithCancel(ctx)
	queueDone := make(chan error, 1)
	go func() { queueDone <- dependencies.Queue().Start(queueCtx) }()
	if err := dependencies.Queue().Publish(ctx, corestorage.Message{
		Topic:  "profile-contract",
		Values: map[string]interface{}{"value": "redis"},
	}); err != nil {
		t.Fatalf("Queue.Publish: %v", err)
	}
	select {
	case message := <-delivered:
		if message.Values["value"] != "redis" {
			t.Fatalf("queue value = %#v, want redis", message.Values["value"])
		}
	case <-ctx.Done():
		t.Fatalf("queue delivery: %v", ctx.Err())
	}
	stopQueue()
	if err := <-queueDone; err != nil {
		t.Fatalf("Queue.Start after cancellation: %v", err)
	}

	if err := dependencies.Files().Put(ctx, "contract.txt", strings.NewReader("server")); err != nil {
		t.Fatalf("Files.Put: %v", err)
	}
	file, err := dependencies.Files().Open(ctx, "contract.txt")
	if err != nil {
		t.Fatalf("Files.Open: %v", err)
	}
	contents, readErr := io.ReadAll(file)
	file.Close()
	if readErr != nil || string(contents) != "server" {
		t.Fatalf("file contents = %q, %v; want server, nil", contents, readErr)
	}
	request := httptest.NewRequest("GET", "http://admin.integration.test/api/v1/menu", nil)
	if tenantID, err := dependencies.Tenants().Resolve(request); err != nil || tenantID != "main" {
		t.Fatalf("tenant Resolve = %q, %v; want main, nil", tenantID, err)
	}
}

func restoreWorkingDirectory(t *testing.T) {
	t.Helper()
	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	projectRoot := filepath.Clean(filepath.Join(workingDirectory, "../.."))
	if err := os.Chdir(projectRoot); err != nil {
		t.Fatalf("Chdir project root: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(workingDirectory); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})
}
