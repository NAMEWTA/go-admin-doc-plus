package reliableruntime_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"go-admin/internal/platform/coordination"
	"go-admin/internal/platform/database"
	"go-admin/internal/platform/outbox"
)

func TestDispatcherBacksOffOnDependencyFailure(t *testing.T) {
	t.Parallel()
	db := openReliableSQLite(t)
	dispatcher, err := outbox.NewDispatcher(outbox.NewStore(db), outbox.DispatcherConfig{
		Owner: "worker-a", LeaseDuration: time.Minute, RetryDelay: time.Minute, BatchSize: 10,
	}, nil)
	if err != nil {
		t.Fatalf("NewDispatcher() error = %v", err)
	}
	dependencyFailure := errors.New("database unavailable")
	executor := failingExecutor{err: dependencyFailure}
	var waits []time.Duration
	stop := errors.New("stop loop")
	err = dispatcher.Run(context.Background(), &executor, outbox.LoopOptions{
		PollInterval: time.Second, FailureBackoff: 30 * time.Second,
		Now: func() time.Time { return time.Date(2026, 8, 26, 16, 0, 0, 0, time.UTC) },
		Wait: func(_ context.Context, delay time.Duration) error {
			waits = append(waits, delay)
			return stop
		},
	})
	if !errors.Is(err, stop) || len(waits) != 1 || waits[0] != 30*time.Second || executor.calls != 1 {
		t.Fatalf("Run() = err %v, waits %v, calls %d", err, waits, executor.calls)
	}
}

func TestDispatcherStopsImmediatelyWhenExecutorLeaseIsLost(t *testing.T) {
	t.Parallel()
	db := openReliableSQLite(t)
	dispatcher, err := outbox.NewDispatcher(outbox.NewStore(db), outbox.DispatcherConfig{
		Owner: "worker-a", LeaseDuration: time.Minute, RetryDelay: time.Minute, BatchSize: 10,
	}, nil)
	if err != nil {
		t.Fatalf("NewDispatcher() error = %v", err)
	}
	executor := failingExecutor{err: coordination.ErrLeaseLost}
	waits := 0
	err = dispatcher.Run(context.Background(), &executor, outbox.LoopOptions{
		PollInterval: time.Second, FailureBackoff: 30 * time.Second,
		Now: func() time.Time { return time.Date(2026, 8, 26, 16, 30, 0, 0, time.UTC) },
		Wait: func(context.Context, time.Duration) error {
			waits++
			return nil
		},
	})
	if !errors.Is(err, coordination.ErrLeaseLost) || waits != 0 || executor.calls != 1 {
		t.Fatalf("Run() = err %v, waits %d, calls %d", err, waits, executor.calls)
	}
}

func TestDispatcherRetriesConsumerFailureWithoutLeakingDiagnostic(t *testing.T) {
	t.Parallel()
	db := openReliableSQLite(t)
	if _, err := db.SQL().Exec(`CREATE TABLE effects (business_key TEXT PRIMARY KEY, calls INTEGER NOT NULL)`); err != nil {
		t.Fatalf("create effects: %v", err)
	}
	store := outbox.NewStore(db)
	now := time.Date(2026, 8, 26, 17, 0, 0, 0, time.UTC)
	event := reliableEvent("event-dispatch-1", "order:dispatch", now)
	enqueueEvent(t, db, store, event)
	consumer := effectConsumer{name: "projector", fail: errors.New("secret diagnostic")}
	dispatcher, err := outbox.NewDispatcher(store, outbox.DispatcherConfig{
		Owner: "worker-a", LeaseDuration: time.Minute, RetryDelay: 5 * time.Minute, BatchSize: 10,
	}, map[string]outbox.Consumer{event.Topic: consumer})
	if err != nil {
		t.Fatalf("NewDispatcher() error = %v", err)
	}
	result, err := dispatcher.RunOnce(context.Background(), directExecutor{db: db}, now)
	if err != nil || result.Claimed != 1 || result.Retried != 1 || result.Delivered != 0 {
		t.Fatalf("RunOnce() = %#v, %v", result, err)
	}
	record, found, err := store.Lookup(context.Background(), event.ID)
	if err != nil || !found || record.State != outbox.StateRetry || record.LastErrorCode != "consumer_failed" {
		t.Fatalf("retry record = %#v, found %v, err %v", record, found, err)
	}
	assertEffectCalls(t, db, event.BusinessKey, 0)
}

type directExecutor struct{ db *database.Database }

func (e directExecutor) WithinTx(ctx context.Context, fn func(context.Context, database.Tx) error) error {
	return e.db.WithinTx(ctx, fn)
}

type failingExecutor struct {
	err   error
	calls int
}

func (e *failingExecutor) WithinTx(context.Context, func(context.Context, database.Tx) error) error {
	e.calls++
	return e.err
}
