package openapi_test

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-admin-team/go-admin-core/v2/sdk"
	"github.com/go-admin-team/go-admin-core/v2/sdk/config"

	"go-admin/internal/application"
	"go-admin/internal/modules"
)

type document struct {
	OpenAPI    string                     `json:"openapi"`
	Paths      map[string]json.RawMessage `json:"paths"`
	Components struct {
		SecuritySchemes map[string]struct {
			Type         string `json:"type"`
			Scheme       string `json:"scheme"`
			BearerFormat string `json:"bearerFormat"`
		} `json:"securitySchemes"`
		Schemas map[string]struct {
			Properties map[string]json.RawMessage `json:"properties"`
		} `json:"schemas"`
	} `json:"components"`
}

func TestCanonicalOpenAPIIncludesBusinessAndOperationalContracts(t *testing.T) {
	contents, err := os.ReadFile("openapi.json")
	if err != nil {
		t.Fatalf("read canonical OpenAPI: %v", err)
	}
	var spec document
	if err := json.Unmarshal(contents, &spec); err != nil {
		t.Fatalf("parse canonical OpenAPI: %v", err)
	}
	if spec.OpenAPI != "3.0.3" {
		t.Fatalf("openapi = %q, want 3.0.3", spec.OpenAPI)
	}
	for _, path := range []string{
		"/api/v1/login",
		"/api/v1/menurole",
		"/api/v1/demo-product",
		"/api/v1/sysjob",
		"/health/live",
		"/health/ready",
		"/api/v1/runtime/capabilities",
	} {
		if _, exists := spec.Paths[path]; !exists {
			t.Errorf("canonical OpenAPI omitted %s", path)
		}
	}
	bearer := spec.Components.SecuritySchemes["Bearer"]
	if bearer.Type != "http" || bearer.Scheme != "bearer" || bearer.BearerFormat != "JWT" {
		t.Errorf("Bearer security scheme = %#v", bearer)
	}
	assertSchemaFields(t, spec, "ApiErrorEnvelope", "code", "data", "msg")
	assertSchemaFields(t, spec, "RuntimeCapabilities", "hostProfile", "version", "desktop", "offline", "nativeDialogs")
	assertSchemaFields(t, spec, "OperationalStatus", "status")
	if len(spec.Components.Schemas["RuntimeCapabilities"].Properties) != 5 {
		t.Errorf("RuntimeCapabilities contains fields beyond the five public capability values")
	}
	menuContract := string(spec.Paths["/api/v1/menurole"])
	if !strings.Contains(menuContract, "ApiErrorEnvelope") || !strings.Contains(menuContract, "Bearer") {
		t.Errorf("protected business operation omitted Bearer or error-envelope contract: %s", menuContract)
	}
	publicOperations := map[string]struct{}{
		"GET /api/v1/app-config":           {},
		"GET /api/v1/captcha":              {},
		"GET /api/v1/health":               {},
		"GET /api/v1/job/remove/{id}":      {},
		"GET /api/v1/job/start/{id}":       {},
		"GET /api/v1/metrics":              {},
		"GET /api/v1/runtime/capabilities": {},
		"POST /api/v1/login":               {},
	}
	for path, raw := range spec.Paths {
		if !strings.HasPrefix(path, "/api/v1/") {
			continue
		}
		var item map[string]json.RawMessage
		if err := json.Unmarshal(raw, &item); err != nil {
			t.Fatalf("parse path item %s: %v", path, err)
		}
		for method, operation := range item {
			key := strings.ToUpper(method) + " " + path
			_, public := publicOperations[key]
			hasBearer := strings.Contains(string(operation), `"Bearer"`)
			if public == hasBearer {
				t.Errorf("%s security declaration does not match the registered authentication boundary", key)
			}
			if path != "/api/v1/health" && path != "/api/v1/metrics" && path != "/api/v1/runtime/capabilities" && countReference(t, operation, "#/components/schemas/ApiErrorEnvelope") != 1 {
				t.Errorf("%s %s must include exactly one error-envelope alternative", strings.ToUpper(method), path)
			}
		}
	}
}

func TestCanonicalOpenAPIMatchesRegisteredAPIRoutes(t *testing.T) {
	contents, err := os.ReadFile("openapi.json")
	if err != nil {
		t.Fatalf("read canonical OpenAPI: %v", err)
	}
	var spec document
	if err := json.Unmarshal(contents, &spec); err != nil {
		t.Fatalf("parse canonical OpenAPI: %v", err)
	}

	previousMode := config.ApplicationConfig.Mode
	previousSecret := config.JwtConfig.Secret
	previousEngine := sdk.Runtime.GetEngine()
	t.Cleanup(func() {
		config.ApplicationConfig.Mode = previousMode
		config.JwtConfig.Secret = previousSecret
		sdk.Runtime.SetEngine(previousEngine)
	})
	config.ApplicationConfig.Mode = "prod"
	config.JwtConfig.Secret = "openapi-contract-test"
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	sdk.Runtime.SetEngine(engine)
	if _, err := application.Build(
		application.Config{Name: "openapi-contract-test"},
		application.Dependencies{Handler: engine},
		modules.Default(),
	); err != nil {
		t.Fatalf("build application route inventory: %v", err)
	}

	registered := map[string]struct{}{
		"GET /api/v1/runtime/capabilities": {},
	}
	parameter := regexp.MustCompile(`:([A-Za-z0-9_]+)`)
	for _, route := range engine.Routes() {
		if strings.HasPrefix(route.Path, "/api/v1") {
			path := parameter.ReplaceAllString(route.Path, `{$1}`)
			registered[route.Method+" "+path] = struct{}{}
		}
	}

	documented := make(map[string]struct{})
	for path, raw := range spec.Paths {
		if !strings.HasPrefix(path, "/api/v1") {
			continue
		}
		var item map[string]json.RawMessage
		if err := json.Unmarshal(raw, &item); err != nil {
			t.Fatalf("parse path item %s: %v", path, err)
		}
		for method := range item {
			documented[strings.ToUpper(method)+" "+path] = struct{}{}
		}
	}

	if missing := setDifference(registered, documented); len(missing) != 0 {
		t.Errorf("registered API routes missing from OpenAPI:\n  %s", strings.Join(missing, "\n  "))
	}
	if stale := setDifference(documented, registered); len(stale) != 0 {
		t.Errorf("OpenAPI contains routes not registered by the Application:\n  %s", strings.Join(stale, "\n  "))
	}
}

func assertSchemaFields(t *testing.T, spec document, name string, fields ...string) {
	t.Helper()
	schema, exists := spec.Components.Schemas[name]
	if !exists {
		t.Fatalf("canonical OpenAPI omitted schema %s", name)
	}
	for _, field := range fields {
		if _, exists := schema.Properties[field]; !exists {
			t.Errorf("schema %s omitted field %s", name, field)
		}
	}
}

func countReference(t *testing.T, raw json.RawMessage, reference string) int {
	t.Helper()
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		t.Fatalf("parse operation JSON: %v", err)
	}
	var visit func(any) int
	visit = func(current any) int {
		switch typed := current.(type) {
		case []any:
			count := 0
			for _, child := range typed {
				count += visit(child)
			}
			return count
		case map[string]any:
			count := 0
			for key, child := range typed {
				if key == "$ref" && fmt.Sprint(child) == reference {
					count++
				}
				count += visit(child)
			}
			return count
		default:
			return 0
		}
	}
	return visit(value)
}

func setDifference(left, right map[string]struct{}) []string {
	result := make([]string, 0)
	for value := range left {
		if _, exists := right[value]; !exists {
			result = append(result, value)
		}
	}
	sort.Strings(result)
	return result
}
