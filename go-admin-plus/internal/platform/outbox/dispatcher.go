package outbox

import (
	"context"
	"errors"
	"time"

	"go-admin/internal/platform/coordination"
	"go-admin/internal/platform/database"
)

type Executor interface {
	WithinTx(context.Context, func(context.Context, database.Tx) error) error
}

type DispatcherConfig struct {
	Owner         string
	LeaseDuration time.Duration
	RetryDelay    time.Duration
	BatchSize     int
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
	consumers map[string]Consumer
}

func NewDispatcher(store *Store, config DispatcherConfig, consumers map[string]Consumer) (*Dispatcher, error) {
	if store == nil || store.db == nil || !validOwner(config.Owner) || config.LeaseDuration <= 0 ||
		config.RetryDelay <= 0 || config.BatchSize < 1 || config.BatchSize > 1000 {
		return nil, errors.New("outbox dispatcher config is invalid")
	}
	owned := make(map[string]Consumer, len(consumers))
	for topic, consumer := range consumers {
		if !topicPattern.MatchString(topic) || consumer == nil || !validOwner(consumer.Name()) {
			return nil, errors.New("outbox consumer registration is invalid")
		}
		owned[topic] = consumer
	}
	return &Dispatcher{store: store, config: config, consumers: owned}, nil
}

// RunOnce claims and settles one bounded batch while the supplied executor lease is held.
func (d *Dispatcher) RunOnce(ctx context.Context, executor Executor, now time.Time) (DispatchResult, error) {
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
			if err := d.retry(ctx, executor, record, now, "consumer_missing"); err != nil {
				return result, err
			}
			result.Retried++
			continue
		}
		var delivered bool
		err := executor.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error {
			var err error
			delivered, err = d.store.Deliver(ctx, tx, record, consumer, now)
			return err
		})
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return result, err
			}
			if err := d.retry(ctx, executor, record, now, "consumer_failed"); err != nil {
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

func (d *Dispatcher) retry(ctx context.Context, executor Executor, record Record, now time.Time, code string) error {
	return executor.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error {
		return d.store.Retry(ctx, tx, record, now.Add(d.config.RetryDelay), code)
	})
}

type LoopOptions struct {
	PollInterval   time.Duration
	FailureBackoff time.Duration
	Now            func() time.Time
	Wait           func(context.Context, time.Duration) error
}

// Run dispatches until cancellation. Every completed or failed poll waits before another database
// operation, preventing dependency failure from becoming a busy loop.
func (d *Dispatcher) Run(ctx context.Context, executor Executor, options LoopOptions) error {
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
		_, err := d.RunOnce(ctx, executor, now())
		if errors.Is(err, coordination.ErrLeaseLost) || errors.Is(err, coordination.ErrNotLeader) {
			return err
		}
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}
		delay := options.PollInterval
		if err != nil {
			delay = options.FailureBackoff
		}
		if err := wait(ctx, delay); err != nil {
			return err
		}
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
