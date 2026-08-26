package jobs

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/go-admin-team/go-admin-core/v2/sdk"
	"github.com/go-admin-team/go-admin-core/v2/sdk/pkg/cronjob"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"

	jobmodels "go-admin/app/jobs/models"
	"go-admin/internal/application"
)

// ErrAlreadyRunning reports an attempt to start one Module twice.
var ErrAlreadyRunning = errors.New("jobs module is already running")

type runningCron struct {
	tenant string
	cron   *cron.Cron
}

type Module struct {
	mu        sync.Mutex
	databases map[string]*gorm.DB
	running   []runningCron
	started   bool
	runID     uint64
	cancel    context.CancelFunc
}

// NewModule uses the legacy Runtime database map until the expand phase ends.
func NewModule() *Module { return &Module{} }

// NewModuleWithDatabases copies an explicit tenant database map for a profile.
func NewModuleWithDatabases(databases map[string]*gorm.DB) *Module {
	return &Module{databases: copyDatabases(databases)}
}

func (*Module) ID() string { return "jobs" }

func (*Module) Migrations() []application.Migration { return nil }

func (m *Module) Start(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if m.started {
		return ErrAlreadyRunning
	}

	InitJob()
	runCtx, cancel := context.WithCancel(ctx)
	keepContext := false
	defer func() {
		if !keepContext {
			cancel()
		}
	}()
	databases := m.databases
	if databases == nil {
		databases = sdk.Runtime.GetAllDb()
	}
	tenants := make([]string, 0, len(databases))
	for tenant := range databases {
		tenants = append(tenants, tenant)
	}
	sort.Strings(tenants)

	for _, tenant := range tenants {
		if err := ctx.Err(); err != nil {
			cancel()
			m.stopLocked(context.WithoutCancel(ctx))
			return err
		}
		crontab := cronjob.NewWithSeconds()
		bindCronContext(crontab, runCtx)
		if err := configureCron(runCtx, databases[tenant], crontab); err != nil {
			unbindCronContext(crontab)
			cancel()
			m.stopLocked(context.WithoutCancel(ctx))
			return fmt.Errorf("configure jobs for tenant %q: %w", tenant, err)
		}
		sdk.Runtime.SetCrontabByTenant(tenant, crontab)
		crontab.Start()
		m.running = append(m.running, runningCron{tenant: tenant, cron: crontab})
	}
	if err := runCtx.Err(); err != nil {
		cancel()
		m.stopLocked(context.WithoutCancel(ctx))
		return err
	}
	m.runID++
	runID := m.runID
	m.cancel = cancel
	m.started = true
	keepContext = true
	go m.stopWhenCancelled(runCtx, runID)
	return nil
}

func (m *Module) stopWhenCancelled(ctx context.Context, runID uint64) {
	<-ctx.Done()
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.started || m.runID != runID {
		return
	}
	_ = m.stopLocked(context.Background())
}

func configureCron(ctx context.Context, database *gorm.DB, crontab *cron.Cron) error {
	if database == nil {
		return errors.New("database is required")
	}
	database = database.WithContext(ctx)

	model := jobmodels.SysJob{}
	jobList := make([]jobmodels.SysJob, 0)
	if err := model.GetList(database, &jobList); err != nil {
		return err
	}
	if _, err := model.RemoveAllEntryID(database); err != nil {
		return err
	}

	for index := range jobList {
		job := &jobList[index]
		var scheduled Job
		switch job.JobType {
		case 1:
			scheduled = &HttpJob{JobCore: JobCore{
				InvokeTarget: job.InvokeTarget, Name: job.JobName, JobId: job.JobId,
				CronExpression: job.CronExpression, ctx: ctx,
			}}
		case 2:
			scheduled = &ExecJob{JobCore: JobCore{
				InvokeTarget: job.InvokeTarget, Name: job.JobName, JobId: job.JobId,
				CronExpression: job.CronExpression, Args: job.Args, ctx: ctx,
			}}
		default:
			continue
		}

		entryID, err := AddJob(crontab, scheduled)
		if err != nil {
			return fmt.Errorf("add job %d: %w", job.JobId, err)
		}
		update := jobmodels.SysJob{EntryId: entryID}
		if err := update.Update(database, job.JobId); err != nil {
			return fmt.Errorf("persist job %d entry ID: %w", job.JobId, err)
		}
	}
	return nil
}

func (m *Module) Stop(ctx context.Context) error {
	if ctx == nil {
		return errors.New("stop context is required")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.stopLocked(ctx)
}

func (m *Module) stopLocked(ctx context.Context) error {
	if m.cancel != nil {
		m.cancel()
		m.cancel = nil
	}
	var stopErrors []error
	for index := len(m.running) - 1; index >= 0; index-- {
		running := m.running[index]
		done := running.cron.Stop()
		select {
		case <-done.Done():
		case <-ctx.Done():
			stopErrors = append(stopErrors, fmt.Errorf("stop jobs for tenant %q: %w", running.tenant, ctx.Err()))
		}
		unbindCronContext(running.cron)
	}
	m.running = nil
	m.started = false
	return errors.Join(stopErrors...)
}

func copyDatabases(databases map[string]*gorm.DB) map[string]*gorm.DB {
	if databases == nil {
		return nil
	}
	result := make(map[string]*gorm.DB, len(databases))
	for tenant, database := range databases {
		result[tenant] = database
	}
	return result
}
