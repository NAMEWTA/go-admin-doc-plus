package tenant_test

import (
	"errors"
	"net/http/httptest"
	"testing"

	"go-admin/internal/tenant"
)

func TestResolversUseFixedDesktopTenantAndTrustedServerHosts(t *testing.T) {
	desktop := tenant.Fixed("local")
	request := httptest.NewRequest("GET", "http://untrusted.example/api/v1/menu", nil)
	if got, err := desktop.Resolve(request); err != nil || got != "local" {
		t.Fatalf("desktop Resolve = %q, %v; want local, nil", got, err)
	}

	server, err := tenant.NewServerResolver(map[string]string{
		"admin.example.test": "tenant-a",
	})
	if err != nil {
		t.Fatalf("NewServerResolver: %v", err)
	}
	request = httptest.NewRequest("GET", "http://admin.example.test:8080/api/v1/menu", nil)
	request.Header.Set("X-Forwarded-Host", "attacker.example")
	if got, err := server.Resolve(request); err != nil || got != "tenant-a" {
		t.Fatalf("server Resolve = %q, %v; want tenant-a, nil", got, err)
	}

	request = httptest.NewRequest("GET", "http://unknown.example/api/v1/menu", nil)
	if _, err := server.Resolve(request); !errors.Is(err, tenant.ErrUnknownHost) {
		t.Fatalf("unknown host error = %v, want ErrUnknownHost", err)
	}
}
