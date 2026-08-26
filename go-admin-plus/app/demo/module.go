package demo

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-admin-team/go-admin-core/v2/sdk"

	"go-admin/app/demo/router"
	common "go-admin/common/middleware"
	"go-admin/internal/application"
)

type Module struct{}

func NewModule() *Module { return &Module{} }

func (*Module) ID() string { return "demo" }

func (*Module) RegisterRoutes(handler http.Handler) error {
	engine, ok := handler.(*gin.Engine)
	if !ok {
		return fmt.Errorf("demo module requires *gin.Engine, got %T", handler)
	}
	sdk.Runtime.SetEngine(engine)
	authMiddleware, err := common.AuthInit()
	if err != nil {
		return fmt.Errorf("initialize demo authentication: %w", err)
	}
	router.InitBusinessRouter(engine, authMiddleware)
	return nil
}

func (*Module) Migrations() []application.Migration { return nil }
func (*Module) Start(context.Context) error         { return nil }
func (*Module) Stop(context.Context) error          { return nil }
