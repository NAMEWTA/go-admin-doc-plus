package profile_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"go-admin/internal/profile"
)

func TestBuildServerFailsClosedWithoutLeakingConnectionSecrets(t *testing.T) {
	if _, err := profile.BuildServer(context.Background(), profile.ServerConfig{}); !errors.Is(err, profile.ErrInvalidServerConfig) {
		t.Fatalf("empty config error = %v, want ErrInvalidServerConfig", err)
	}

	postgresSecret := "postgres-password-do-not-log"
	redisSecret := "redis-token-do-not-log"
	_, err := profile.BuildServer(context.Background(), profile.ServerConfig{
		PostgresDSN: "postgres://admin:" + postgresSecret + "@127.0.0.1:1/go_admin?sslmode=disable",
		RedisURL:    "redis://default:" + redisSecret + "@127.0.0.1:1/0",
		FileRoot:    t.TempDir(),
		TenantHosts: map[string]string{"admin.example.test": "main"},
	})
	if err == nil {
		t.Fatal("BuildServer unexpectedly succeeded with unreachable dependencies")
	}
	if strings.Contains(err.Error(), postgresSecret) || strings.Contains(err.Error(), redisSecret) {
		t.Fatalf("BuildServer leaked a connection secret: %v", err)
	}
}
