package jobs_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-admin-team/go-admin-core/v2/sdk"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"go-admin/app/jobs"
	jobmodels "go-admin/app/jobs/models"
)

func TestModuleStopsWhenStartContextIsCancelled(t *testing.T) {
	database, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	if err := database.AutoMigrate(&jobmodels.SysJob{}); err != nil {
		t.Fatalf("AutoMigrate jobs: %v", err)
	}
	module := jobs.NewModuleWithDatabases(map[string]*gorm.DB{"local": database})
	ctx, cancel := context.WithCancel(context.Background())
	if err := module.Start(ctx); err != nil {
		t.Fatalf("first Start: %v", err)
	}
	cancel()

	deadline := time.Now().Add(2 * time.Second)
	for {
		err := module.Start(context.Background())
		if err == nil {
			break
		}
		if !errors.Is(err, jobs.ErrAlreadyRunning) {
			t.Fatalf("second Start: %v", err)
		}
		if time.Now().After(deadline) {
			t.Fatal("jobs did not stop after context cancellation")
		}
		time.Sleep(10 * time.Millisecond)
	}
	if err := module.Stop(context.Background()); err != nil {
		t.Fatalf("Stop restarted module: %v", err)
	}
}

func TestLegacySetupReturnsAClosableModule(t *testing.T) {
	database, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	if err := database.AutoMigrate(&jobmodels.SysJob{}); err != nil {
		t.Fatalf("AutoMigrate jobs: %v", err)
	}

	module, err := jobs.Setup(map[string]*gorm.DB{"local": database})
	if err != nil {
		t.Fatalf("Setup: %v", err)
	}
	if err := module.Stop(context.Background()); err != nil {
		t.Fatalf("Stop: %v", err)
	}
}

func TestDynamicallyAddedHTTPJobUsesModuleContext(t *testing.T) {
	database, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	if err := database.AutoMigrate(&jobmodels.SysJob{}); err != nil {
		t.Fatalf("AutoMigrate jobs: %v", err)
	}
	requestStarted := make(chan struct{})
	requestStopped := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, request *http.Request) {
		close(requestStarted)
		<-request.Context().Done()
		close(requestStopped)
	}))
	t.Cleanup(func() {
		server.CloseClientConnections()
		server.Close()
	})

	ctx, cancel := context.WithCancel(context.Background())
	module := jobs.NewModuleWithDatabases(map[string]*gorm.DB{"local": database})
	if err := module.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	crontab := sdk.Runtime.GetCrontabByTenant("local")
	if _, err := jobs.AddJob(crontab, &jobs.HttpJob{JobCore: jobs.JobCore{
		InvokeTarget: server.URL, Name: "dynamic", CronExpression: "@every 10ms",
	}}); err != nil {
		t.Fatalf("AddJob: %v", err)
	}
	select {
	case <-requestStarted:
	case <-time.After(2 * time.Second):
		t.Fatal("dynamic HTTP job did not start")
	}
	cancel()
	select {
	case <-requestStopped:
	case <-time.After(2 * time.Second):
		t.Fatal("dynamic HTTP job did not receive module cancellation")
	}
	if err := module.Stop(context.Background()); err != nil {
		t.Fatalf("Stop: %v", err)
	}
}
