package characterization_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-admin-team/go-admin-core/v2/sdk"
	coreconfig "github.com/go-admin-team/go-admin-core/v2/sdk/config"

	adminrouter "go-admin/app/admin/router"
	demorouter "go-admin/app/demo/router"
	jobsrouter "go-admin/app/jobs/router"
	"go-admin/common/middleware"
)

type routeKey struct {
	method string
	path   string
}

func TestAPIV1RouteManifest(t *testing.T) {
	engine := baselineRouter(t)
	routes := make(map[routeKey]struct{}, len(engine.Routes()))
	for _, route := range engine.Routes() {
		routes[routeKey{method: route.Method, path: route.Path}] = struct{}{}
	}

	want := []routeKey{
		{method: http.MethodPost, path: "/api/v1/login"},
		{method: http.MethodGet, path: "/api/v1/menurole"},
		{method: http.MethodGet, path: "/api/v1/demo-product"},
		{method: http.MethodGet, path: "/api/v1/demo-product/:id"},
		{method: http.MethodPost, path: "/api/v1/demo-product"},
		{method: http.MethodPut, path: "/api/v1/demo-product/:id"},
		{method: http.MethodDelete, path: "/api/v1/demo-product"},
		{method: http.MethodPost, path: "/api/v1/user/avatar"},
		{method: http.MethodGet, path: "/api/v1/sysjob"},
		{method: http.MethodGet, path: "/api/v1/job/start/:id"},
		{method: http.MethodGet, path: "/api/v1/job/remove/:id"},
	}

	for _, route := range want {
		if _, ok := routes[route]; !ok {
			t.Errorf("missing route %s %s", route.method, route.path)
		}
	}
}

func TestProtectedAPIsReturnTheAuthorizationEnvelope(t *testing.T) {
	engine := baselineRouter(t)
	tests := []struct {
		name   string
		method string
		path   string
	}{
		{name: "menu", method: http.MethodGet, path: "/api/v1/menurole"},
		{name: "product", method: http.MethodGet, path: "/api/v1/demo-product"},
		{name: "upload", method: http.MethodPost, path: "/api/v1/user/avatar"},
		{name: "jobs", method: http.MethodGet, path: "/api/v1/sysjob"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, tt.path, nil)
			response := httptest.NewRecorder()
			engine.ServeHTTP(response, request)

			if response.Code != http.StatusOK {
				t.Fatalf("HTTP status = %d, want 200", response.Code)
			}
			var envelope struct {
				Code int    `json:"code"`
				Msg  string `json:"msg"`
			}
			if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
				t.Fatalf("decode response %q: %v", response.Body.String(), err)
			}
			if envelope.Code != http.StatusUnauthorized {
				t.Fatalf("code = %d, want 401", envelope.Code)
			}
			if envelope.Msg == "" {
				t.Fatal("authorization failure omitted msg")
			}
		})
	}
}

func baselineRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	previousMode := coreconfig.ApplicationConfig.Mode
	previousSecret := coreconfig.JwtConfig.Secret
	coreconfig.ApplicationConfig.Mode = "prod"
	coreconfig.JwtConfig.Secret = "characterization-only-secret"
	t.Cleanup(func() {
		coreconfig.ApplicationConfig.Mode = previousMode
		coreconfig.JwtConfig.Secret = previousSecret
	})

	engine := gin.New()
	sdk.Runtime.SetEngine(engine)
	auth, err := middleware.AuthInit()
	if err != nil {
		t.Fatalf("initialize auth middleware: %v", err)
	}
	adminrouter.InitSysRouter(engine, auth)
	adminrouter.InitExamplesRouter(engine, auth)
	demorouter.InitBusinessRouter(engine, auth)
	jobsrouter.InitRouter()
	return engine
}
