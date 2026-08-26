package outbox

import (
	"context"
	"errors"
	"time"

	"go-admin/internal/platform/coordination"
	"go-admin/internal/platform/database"
)

type transactionExecutor interface {
	WithinTx(context.Context, func(context.Context, database.Tx) error) error
}

type DispatcherConfig struct {
	Owner         string
	LeaseDuration time.Duration
	RetryDelay    time.Duration
	BatchSize     int
	Now           func() time.Time
}

type DispatchResult struct {
	Claimed   int
	Delivered int
	Replayed  int
	Retried   int
}

type Dispatcher struct {
	store     *Store
	config    DispatcherConfig
	consumers map[string]TransactionalConsumer
	now       func() time.Time
}

func NewDispatcher(store *Store, config DispatcherConfig, consumers map[string]TransactionalConsumer) (*Dispatcher, error) {
	if store == nil || store.db == nil || !validOwner(config.Owner) || config.LeaseDuration <= 0 ||
		config.RetryDelay <= 0 || config.BatchSize < 1 || config.BatchSize > 1000 {
		return nil, errors.New("outbox dispatcher config is invalid")
	}
	owned := make(map[string]TransactionalConsumer, len(consumers))
	for topic, consumer := range consumers {
		if !topicPattern.MatchString(topic) || !validOwner(consumer.Name()) || len(consumer.statements) == 0 {
			return nil, errors.New("outbox consumer registration is invalid")
		}
		owned[topic] = consumer
	}
	now := config.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return &Dispatcher{store: store, config: config, consumers: owned, now: now}, nil
}

// RunOnce claims and settles one bounded batch while the supplied executor lease is held.
func (d *Dispatcher) RunOnce(ctx context.Context, lease *coordination.Lease, now time.Time) (DispatchResult, error) {
	if lease == nil {
		return DispatchResult{}, errors.New("outbox coordination lease is required")
	}
	return d.runOnce(ctx, lease, now)
}

func (d *Dispatcher) runOnce(ctx context.Context, executor transactionExecutor, now time.Time) (DispatchResult, error) {
	if d == nil || executor == nil || now.IsZero() || now.Location() != time.UTC {
		return DispatchResult{}, errors.New("outbox dispatch input is invalid")
	}
	var records []Record
	err := executor.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error {
		var err error
		records, err = d.store.Claim(ctx, tx, ClaimOptions{
			Owner: d.config.Owner, Now: now, Lease: d.config.LeaseDuration, Limit: d.config.BatchSize,
		})
		return err
	})
	if err != nil {
		return DispatchResult{}, err
	}
	result := DispatchResult{Claimed: len(records)}
	for _, record := range records {
		consumer, ok := d.consumers[record.Event.Topic]
		if !ok {
			if err := d.retry(ctx, executor, record, "consumer_missing"); err != nil {
				return result, err
			}
			result.Retried++
			continue
		}
		var delivered bool
		err := executor.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error {
			var err error
			delivered, err = d.store.deliver(ctx, tx, record, consumer, d.now)
			return err
		})
		if err != nil {
			if errors.Is(err, coordination.ErrLeaseLost) || errors.Is(err, coordination.ErrNotLeader) {
				return result, err
			}
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return result, err
			}
			if err := d.retry(ctx, executor, record, "consumer_failed"); err != nil {
				return result, err
			}
			result.Retried++
			continue
		}
		if delivered {
			result.Delivered++
		} else {
			result.Replayed++
		}
	}
	return result, nil
}

func (d *Dispatcher) retry(ctx context.Context, executor transactionExecutor, record Record, code string) error {
	failedAt := d.now().UTC()
	return executor.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error {
		return d.store.Retry(ctx, tx, record, failedAt, failedAt.Add(d.config.RetryDelay), code)
	})
}

type LoopOutcome string

const (
	LoopIdle              LoopOutcome = "idle"
	LoopDispatched        LoopOutcome = "dispatched"
	LoopRetried           LoopOutcome = "retried"
	LoopDependencyFailure LoopOutcome = "dependency_failure"
	LoopLeaseLost         LoopOutcome = "lease_lost"
)

// LoopObservation is deliberately payload- and error-free so operational reporting cannot expose
// database diagnostics or event material.
type LoopObservation struct {
	Outcome LoopOutcome
	Result  DispatchResult
	Delay   time.Duration
}

type Observer interface {
	Observe(LoopObservation)
}

type ObserveFunc func(LoopObservation)

func (fn ObserveFunc) Observe(observation LoopObservation) { fn(observation) }

type LoopOptions struct {
	PollInterval   time.Duration
	FailureBackoff time.Duration
	Now            func() time.Time
	Wait           func(context.Context, time.Duration) error
	Observer       Observer
}

// Run dispatches until cancellation. Every completed or failed poll waits before another database
// operation, preventing dependency failure from becoming a busy loop.
func (d *Dispatcher) Run(ctx context.Context, lease *coordination.Lease, options LoopOptions) error {
	if lease == nil {
		return errors.New("outbox coordination lease is required")
	}
	return d.run(ctx, lease, options)
}

func (d *Dispatcher) run(ctx context.Context, executor transactionExecutor, options LoopOptions) error {
	if d == nil || executor == nil || options.PollInterval <= 0 || options.FailureBackoff <= 0 {
		return errors.New("outbox loop options are invalid")
	}
	now := options.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	wait := options.Wait
	if wait == nil {
		wait = waitFor
	}
	for {
		result, err := d.runOnce(ctx, executor, now())
		if errors.Is(err, coordination.ErrLeaseLost) || errors.Is(err, coordination.ErrNotLeader) {
			observe(options.Observer, LoopObservation{Outcome: LoopLeaseLost, Result: result})
			return err
		}
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}
		delay := options.PollInterval
		outcome := LoopIdle
		if result.Claimed > 0 {
			outcome = LoopDispatched
		}
		if result.Retried > 0 {
			outcome = LoopRetried
		}
		if err != nil {
			delay = options.FailureBackoff
			outcome = LoopDependencyFailure
		}
		observe(options.Observer, LoopObservation{Outcome: outcome, Result: result, Delay: delay})
		if err := wait(ctx, delay); err != nil {
			return err
		}
	}
}

func observe(observer Observer, observation LoopObservation) {
	if observer != nil {
		observer.Observe(observation)
	}
}

func waitFor(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
