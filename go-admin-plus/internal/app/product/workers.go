package product

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/scheduler"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/platform/coordination"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/platform/database"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/platform/outbox"
)

type workerGroup struct {
	mu         sync.Mutex
	db         *database.Database
	owner      string
	interval   time.Duration
	executor   *scheduler.Executor
	dispatcher *outbox.Dispatcher
	lease      *coordination.Lease
	cancel     context.CancelFunc
	done       chan struct{}
	failure    error
	started    bool
}

func newWorkerGroup(db *database.Database, owner string, interval time.Duration, executor *scheduler.Executor, dispatcher *outbox.Dispatcher) *workerGroup {
	return &workerGroup{db: db, owner: owner, interval: interval, executor: executor, dispatcher: dispatcher}
}

func (group *workerGroup) Start(ctx context.Context) error {
	group.mu.Lock()
	defer group.mu.Unlock()
	if group.started {
		return errors.New("product workers already started")
	}
	lease, err := coordination.Acquire(ctx, group.db, coordination.Config{Owner: group.owner})
	if errors.Is(err, coordination.ErrNotLeader) && group.db.Dialect() == database.DialectPostgres {
		group.started = true
		return nil
	}
	if err != nil {
		return errors.New("product worker lease failed")
	}
	workerContext, cancel := context.WithCancel(ctx)
	group.lease = lease
	group.cancel = cancel
	group.done = make(chan struct{})
	group.started = true
	go group.run(workerContext, group.done)
	return nil
}

func (group *workerGroup) run(ctx context.Context, done chan struct{}) {
	defer close(done)
	ticker := time.NewTicker(group.interval)
	defer ticker.Stop()
	for {
		if err := group.runOnce(ctx); err != nil {
			if ctx.Err() == nil {
				group.mu.Lock()
				group.failure = errors.New("product worker execution failed")
				group.mu.Unlock()
			}
			return
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (group *workerGroup) runOnce(ctx context.Context) error {
	if _, err := group.executor.RunOnce(ctx, group.lease); err != nil {
		return err
	}
	_, err := group.dispatcher.RunOnce(ctx, group.lease, time.Now().UTC())
	return err
}

func (group *workerGroup) Stop(ctx context.Context) error {
	group.mu.Lock()
	if !group.started {
		group.mu.Unlock()
		return nil
	}
	cancel, done, lease := group.cancel, group.done, group.lease
	group.started = false
	group.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	if done != nil {
		select {
		case <-done:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	group.mu.Lock()
	group.cancel, group.done, group.lease = nil, nil, nil
	group.mu.Unlock()
	if lease != nil {
		return lease.Close(ctx)
	}
	return nil
}

func (group *workerGroup) Check(context.Context) error {
	group.mu.Lock()
	defer group.mu.Unlock()
	return group.failure
}
