package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	jwt "github.com/go-admin-team/go-admin-core/v2/jwtauth"
	"github.com/go-admin-team/go-admin-core/v2/sdk/config"

	"go-admin/common/middleware"
	"go-admin/common/middleware/handler"
)

func TestAuthInitUsesCanonicalLoginSuccessEnvelope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	previousSecret := config.JwtConfig.Secret
	previousTimeout := config.JwtConfig.Timeout
	config.JwtConfig.Secret = "test-secret"
	config.JwtConfig.Timeout = 3600
	t.Cleanup(func() {
		config.JwtConfig.Secret = previousSecret
		config.JwtConfig.Timeout = previousTimeout
	})

	auth, err := middleware.AuthInit()
	if err != nil {
		t.Fatalf("AuthInit: %v", err)
	}
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	auth.AntdLoginResponse(ctx, http.StatusOK, "session-token", time.Unix(1_800_000_000, 0))

	var body map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["code"] != float64(http.StatusOK) || body["msg"] != "登录成功" || body["token"] != "session-token" {
		t.Fatalf("login response = %#v", body)
	}
	if data, exists := body["data"]; !exists || data != nil {
		t.Fatalf("login data = %#v, exists %v; want explicit null", data, exists)
	}
}

func TestLogoutUsesCanonicalSuccessEnvelope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	previousEnabledDB := config.LoggerConfig.EnabledDB
	config.LoggerConfig.EnabledDB = false
	t.Cleanup(func() { config.LoggerConfig.EnabledDB = previousEnabledDB })

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/logout", nil)
	ctx.Set(jwt.JwtPayloadKey, jwt.MapClaims{"nice": "admin"})
	handler.LogOut(ctx)

	var body map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["code"] != float64(http.StatusOK) || body["msg"] != "退出成功" {
		t.Fatalf("logout response = %#v", body)
	}
	if data, exists := body["data"]; !exists || data != nil {
		t.Fatalf("logout data = %#v, exists %v; want explicit null", data, exists)
	}
}
