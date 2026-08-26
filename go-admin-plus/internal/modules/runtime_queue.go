package modules

import (
	"context"
	"errors"
	"net/http"
	"sync"

	"github.com/go-admin-team/go-admin-core/v2/sdk"
	corestorage "github.com/go-admin-team/go-admin-core/v2/storage"

	"go-admin/app/admin/models"
	"go-admin/common/global"
	"go-admin/internal/application"
)

type runtimeQueueModule struct {
	mu    sync.Mutex
	queue corestorage.AdapterQueue
	done  chan struct{}
}

func newRuntimeQueueModule() *runtimeQueueModule { return &runtimeQueueModule{} }

func (*runtimeQueueModule) ID() string                          { return "runtime-queue" }
func (*runtimeQueueModule) RegisterRoutes(http.Handler) error   { return nil }
func (*runtimeQueueModule) Migrations() []application.Migration { return nil }

func (m *runtimeQueueModule) Start(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if m.queue != nil {
		return errors.New("runtime queue module is already running")
	}

	queue := sdk.Runtime.GetQueuePrefix("")
	queue.Register(global.LoginLog, models.SaveLoginLog)
	queue.Register(global.OperateLog, models.SaveOperaLog)
	queue.Register(global.ApiCheck, models.SaveSysApi)
	done := make(chan struct{})
	m.queue = queue
	m.done = done
	go func() {
		defer close(done)
		queue.Run()
	}()
	return nil
}

func (m *runtimeQueueModule) Stop(ctx context.Context) error {
	m.mu.Lock()
	queue := m.queue
	done := m.done
	m.queue = nil
	m.done = nil
	m.mu.Unlock()

	if queue == nil {
		return nil
	}
	queue.Shutdown()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
