package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	jwt "github.com/go-admin-team/go-admin-core/v2/jwtauth"
	"github.com/go-admin-team/go-admin-core/v2/sdk/config"
	"go-admin/common/middleware/handler"
)

// AuthInit jwt验证new
func AuthInit() (*jwt.GinJWTMiddleware, error) {
	timeout := time.Hour
	if config.ApplicationConfig.Mode == "dev" {
		timeout = time.Duration(876010) * time.Hour
	} else {
		if config.JwtConfig.Timeout != 0 {
			timeout = time.Duration(config.JwtConfig.Timeout) * time.Second
		}
	}
	loginResponse := func(c *gin.Context, _ int, token string, expire time.Time) {
		c.JSON(http.StatusOK, gin.H{
			"code":             http.StatusOK,
			"data":             nil,
			"msg":              "登录成功",
			"success":          true,
			"token":            token,
			"currentAuthority": token,
			"expire":           expire.Format(time.RFC3339),
		})
	}
	return jwt.New(&jwt.GinJWTMiddleware{
		Realm:             "test zone",
		Key:               []byte(config.JwtConfig.Secret),
		Timeout:           timeout,
		MaxRefresh:        time.Hour,
		PayloadFunc:       handler.PayloadFunc,
		IdentityHandler:   handler.IdentityHandler,
		Authenticator:     handler.Authenticator,
		Authorizator:      handler.Authorizator,
		Unauthorized:      handler.Unauthorized,
		LoginResponse:     loginResponse,
		AntdLoginResponse: loginResponse,
		TokenLookup:       "header: Authorization, query: token, cookie: jwt",
		TokenHeadName:     "Bearer",
		TimeFunc:          time.Now,
	})

}
