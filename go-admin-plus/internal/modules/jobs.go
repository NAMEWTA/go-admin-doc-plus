package modules

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-admin-team/go-admin-core/v2/sdk"

	"go-admin/app/jobs"
	jobrouter "go-admin/app/jobs/router"
	common "go-admin/common/middleware"
)

type jobsModule struct {
	*jobs.Module
}

func newJobsModule() *jobsModule {
	return &jobsModule{Module: jobs.NewModule()}
}

func (*jobsModule) RegisterRoutes(handler http.Handler) error {
	engine, ok := handler.(*gin.Engine)
	if !ok {
		return fmt.Errorf("jobs module requires *gin.Engine, got %T", handler)
	}
	sdk.Runtime.SetEngine(engine)
	if _, err := common.AuthInit(); err != nil {
		return fmt.Errorf("initialize jobs authentication: %w", err)
	}
	jobrouter.InitRouter()
	return nil
}
