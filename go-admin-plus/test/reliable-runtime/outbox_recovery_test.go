package reliableruntime_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"go-admin/internal/platform/database"
	"go-admin/internal/platform/outbox"
)

func TestOutboxBusinessKeyIsIdempotentAndImmutable(t *testing.T) {
	t.Parallel()
	db := openReliableSQLite(t)
	store := outbox.NewStore(db)
	event := reliableEvent("event-idempotent-1", "order:42", time.Date(2026, 8, 26, 13, 0, 0, 0, time.UTC))

	for attempt := 0; attempt < 2; attempt++ {
		err := db.WithinTx(context.Background(), func(ctx context.Context, tx database.Tx) error {
			created, err := store.Enqueue(ctx, tx, event)
			if err != nil {
				return err
			}
			if created != (attempt == 0) {
				return errors.New("unexpected enqueue result")
			}
			return nil
		})
		if err != nil {
			t.Fatalf("Enqueue attempt %d: %v", attempt, err)
		}
	}

	conflict := event
	conflict.ID = "event-idempotent-2"
	conflict.Payload = []byte(`{"revision":2}`)
	err := db.WithinTx(context.Background(), func(ctx context.Context, tx database.Tx) error {
		_, err := store.Enqueue(ctx, tx, conflict)
		return err
	})
	if !errors.Is(err, outbox.ErrBusinessKeyConflict) {
		t.Fatalf("conflicting Enqueue() error = %v", err)
	}
}

func TestConcurrentClaimHasOneWinner(t *testing.T) {
	t.Parallel()
	db := openReliableSQLite(t)
	store := outbox.NewStore(db)
	now := time.Date(2026, 8, 26, 13, 30, 0, 0, time.UTC)
	enqueueEvent(t, db, store, reliableEvent("event-concurrent-1", "order:concurrent", now))

	start := make(chan struct{})
	winners := make(chan string, 8)
	var workers sync.WaitGroup
	for worker := 0; worker < 8; worker++ {
		workers.Add(1)
		go func(owner string) {
			defer workers.Done()
			<-start
			var records []outbox.Record
			err := db.WithinTx(context.Background(), func(ctx context.Context, tx database.Tx) error {
				var err error
				records, err = store.Claim(ctx, tx, outbox.ClaimOptions{Owner: owner, Now: now, Lease: time.Minute, Limit: 1})
				return err
			})
			if err == nil && len(records) == 1 {
				winners <- owner
			}
		}(fmt.Sprintf("worker-%d", worker))
	}
	close(start)
	workers.Wait()
	close(winners)
	if len(winners) != 1 {
		t.Fatalf("claim winners = %d, want 1", len(winners))
	}
}

func TestOutboxDatabaseCancellationKeepsContextIdentity(t *testing.T) {
	t.Parallel()
	db := openReliableSQLite(t)
	store := outbox.NewStore(db)
	now := time.Date(2026, 8, 26, 13, 45, 0, 0, time.UTC)
	err := db.WithinTx(context.Background(), func(_ context.Context, tx database.Tx) error {
		canceled, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := store.Claim(canceled, tx, outbox.ClaimOptions{
			Owner: "worker-a", Now: now, Lease: time.Minute, Limit: 1,
		})
		return err
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Claim(canceled) error = %v", err)
	}
}

func TestExpiredClaimRecoversAndConsumerDeliveryIsTransactional(t *testing.T) {
	t.Parallel()
	db := openReliableSQLite(t)
	if _, err := db.SQL().Exec(`CREATE TABLE effects (business_key TEXT PRIMARY KEY, calls INTEGER NOT NULL)`); err != nil {
		t.Fatalf("create effects: %v", err)
	}
	store := outbox.NewStore(db)
	now := time.Date(2026, 8, 26, 14, 0, 0, 0, time.UTC)
	event := reliableEvent("event-recovery-1", "order:43", now)
	enqueueEvent(t, db, store, event)

	first := claimEvents(t, db, store, outbox.ClaimOptions{Owner: "worker-a", Now: now, Lease: time.Minute, Limit: 1})
	if len(first) != 1 || first[0].Attempts != 1 || first[0].State != outbox.StateClaimed {
		t.Fatalf("first claim = %#v", first)
	}
	if got := claimEvents(t, db, store, outbox.ClaimOptions{Owner: "worker-b", Now: now.Add(30 * time.Second), Lease: time.Minute, Limit: 1}); len(got) != 0 {
		t.Fatalf("live claim stolen: %#v", got)
	}
	second := claimEvents(t, db, store, outbox.ClaimOptions{Owner: "worker-b", Now: now.Add(2 * time.Minute), Lease: time.Minute, Limit: 1})
	if len(second) != 1 || second[0].Attempts != 2 || second[0].ClaimedBy != "worker-b" {
		t.Fatalf("recovered claim = %#v", second)
	}

	consumer := newEffectConsumer(t, true)
	err := db.WithinTx(context.Background(), func(ctx context.Context, tx database.Tx) error {
		_, err := store.Deliver(ctx, tx, second[0], consumer, now.Add(2*time.Minute))
		return err
	})
	if !errors.Is(err, outbox.ErrConsumerFailed) || strings.Contains(err.Error(), "missing_effect_target") {
		t.Fatalf("Deliver() error = %v", err)
	}
	assertEffectCalls(t, db, event.BusinessKey, 0)

	retryAt := now.Add(3 * time.Minute)
	if err := db.WithinTx(context.Background(), func(ctx context.Context, tx database.Tx) error {
		return store.Retry(ctx, tx, second[0], now.Add(2*time.Minute), retryAt, "consumer_failed")
	}); err != nil {
		t.Fatalf("Retry() error = %v", err)
	}
	if got := claimEvents(t, db, store, outbox.ClaimOptions{Owner: "worker-c", Now: retryAt.Add(-time.Second), Lease: time.Minute, Limit: 1}); len(got) != 0 {
		t.Fatalf("event claimed before retry time: %#v", got)
	}
	third := claimEvents(t, db, store, outbox.ClaimOptions{Owner: "worker-c", Now: retryAt, Lease: time.Minute, Limit: 1})
	consumer = newEffectConsumer(t, false)
	if err := db.WithinTx(context.Background(), func(ctx context.Context, tx database.Tx) error {
		delivered, err := store.Deliver(ctx, tx, third[0], consumer, retryAt)
		if err != nil || !delivered {
			return errors.New("first delivery did not commit")
		}
		return nil
	}); err != nil {
		t.Fatalf("Deliver() success error = %v", err)
	}
	assertEffectCalls(t, db, event.BusinessKey, 1)

	// A replay with another event ID but the same business key is skipped by the receipt.
	replay := reliableEvent("event-recovery-replay", event.BusinessKey, retryAt.Add(time.Second))
	replay.Topic = "orders.replayed"
	enqueueEvent(t, db, store, replay)
	replayed := claimEvents(t, db, store, outbox.ClaimOptions{Owner: "worker-c", Now: replay.OccurredAt, Lease: time.Minute, Limit: 1})
	if err := db.WithinTx(context.Background(), func(ctx context.Context, tx database.Tx) error {
		delivered, err := store.Deliver(ctx, tx, replayed[0], consumer, replay.OccurredAt)
		if err != nil || delivered {
			return errors.New("receipt did not suppress replay")
		}
		return nil
	}); err != nil {
		t.Fatalf("Deliver() replay error = %v", err)
	}
	assertEffectCalls(t, db, event.BusinessKey, 1)
}

func TestExpiredOwnerCannotSettleOrRunConsumer(t *testing.T) {
	t.Parallel()
	db := openReliableSQLite(t)
	if _, err := db.SQL().Exec(`CREATE TABLE effects (business_key TEXT PRIMARY KEY, calls INTEGER NOT NULL)`); err != nil {
		t.Fatalf("create effects: %v", err)
	}
	store := outbox.NewStore(db)
	now := time.Date(2026, 8, 26, 14, 30, 0, 0, time.UTC)
	event := reliableEvent("event-expired-1", "order:expired", now)
	enqueueEvent(t, db, store, event)
	record := claimEvents(t, db, store, outbox.ClaimOptions{Owner: "worker-a", Now: now, Lease: time.Minute, Limit: 1})[0]
	expiredAt := now.Add(time.Minute)
	consumer := newEffectConsumer(t, false)
	err := db.WithinTx(context.Background(), func(ctx context.Context, tx database.Tx) error {
		_, err := store.Deliver(ctx, tx, record, consumer, expiredAt)
		return err
	})
	if !errors.Is(err, outbox.ErrClaimLost) {
		t.Fatalf("Deliver(expired) error = %v", err)
	}
	assertEffectCalls(t, db, event.BusinessKey, 0)
	err = db.WithinTx(context.Background(), func(ctx context.Context, tx database.Tx) error {
		return store.Retry(ctx, tx, record, expiredAt, expiredAt.Add(time.Minute), "consumer_failed")
	})
	if !errors.Is(err, outbox.ErrClaimLost) {
		t.Fatalf("Retry(expired) error = %v", err)
	}
	stored, found, err := store.Lookup(context.Background(), event.ID)
	if err != nil || !found || stored.State != outbox.StateClaimed {
		t.Fatalf("expired record = %#v, found %v, err %v", stored, found, err)
	}
}

func TestClaimTokenFencesReusedOwner(t *testing.T) {
	t.Parallel()
	db := openReliableSQLite(t)
	if _, err := db.SQL().Exec(`CREATE TABLE effects (business_key TEXT PRIMARY KEY, calls INTEGER NOT NULL)`); err != nil {
		t.Fatalf("create effects: %v", err)
	}
	store := outbox.NewStore(db)
	now := time.Date(2026, 8, 26, 14, 40, 0, 0, time.UTC)
	event := reliableEvent("event-token-fence-1", "order:token-fence", now)
	enqueueEvent(t, db, store, event)
	stale := claimEvents(t, db, store, outbox.ClaimOptions{Owner: "worker-a", Now: now, Lease: time.Minute, Limit: 1})[0]
	current := claimEvents(t, db, store, outbox.ClaimOptions{Owner: "worker-a", Now: now.Add(2 * time.Minute), Lease: time.Minute, Limit: 1})[0]
	if current.Attempts != 2 || current.ClaimedBy != stale.ClaimedBy {
		t.Fatalf("same-owner reclaim = %#v", current)
	}
	err := db.WithinTx(context.Background(), func(ctx context.Context, tx database.Tx) error {
		_, err := store.Deliver(ctx, tx, stale, newEffectConsumer(t, false), now.Add(2*time.Minute))
		return err
	})
	if !errors.Is(err, outbox.ErrClaimLost) {
		t.Fatalf("Deliver(stale token) error = %v", err)
	}
	assertEffectCalls(t, db, event.BusinessKey, 0)
	err = db.WithinTx(context.Background(), func(ctx context.Context, tx database.Tx) error {
		return store.Retry(ctx, tx, stale, now.Add(2*time.Minute), now.Add(3*time.Minute), "consumer_failed")
	})
	if !errors.Is(err, outbox.ErrClaimLost) {
		t.Fatalf("Retry(stale token) error = %v", err)
	}
}

func TestOccurredAtUsesPostgresStableMicrosecondPrecision(t *testing.T) {
	t.Parallel()
	db := openReliableSQLite(t)
	store := outbox.NewStore(db)
	invalid := reliableEvent("event-nanos-1", "order:nanos", time.Date(2026, 8, 26, 14, 45, 0, 123456789, time.UTC))
	err := db.WithinTx(context.Background(), func(ctx context.Context, tx database.Tx) error {
		_, err := store.Enqueue(ctx, tx, invalid)
		return err
	})
	if !errors.Is(err, outbox.ErrInvalidEvent) {
		t.Fatalf("Enqueue(nanoseconds) error = %v", err)
	}
	valid := reliableEvent("event-micros-1", "order:micros", time.Date(2026, 8, 26, 14, 45, 0, 123456000, time.UTC))
	enqueueEvent(t, db, store, valid)
	if err := db.WithinTx(context.Background(), func(ctx context.Context, tx database.Tx) error {
		created, err := store.Enqueue(ctx, tx, valid)
		if err != nil || created {
			return errors.New("microsecond replay was not idempotent")
		}
		return nil
	}); err != nil {
		t.Fatalf("Enqueue(microsecond replay) error = %v", err)
	}
}

func TestEventPayloadRejectsNestedSensitiveFields(t *testing.T) {
	t.Parallel()
	db := openReliableSQLite(t)
	store := outbox.NewStore(db)
	now := time.Date(2026, 8, 26, 14, 50, 0, 0, time.UTC)
	payloads := [][]byte{
		[]byte(`{"password":"value"}`),
		[]byte(`{"profile":{"client_secret":"value"}}`),
		[]byte(`{"items":[{"raw_session":"value"}]}`),
		[]byte(`{"access-token":"value"}`),
		[]byte(`{"credential_bundle":"value"}`),
	}
	for index, payload := range payloads {
		event := reliableEvent(fmt.Sprintf("event-sensitive-%d", index), fmt.Sprintf("order:sensitive:%d", index), now)
		event.Payload = payload
		err := db.WithinTx(context.Background(), func(ctx context.Context, tx database.Tx) error {
			_, err := store.Enqueue(ctx, tx, event)
			return err
		})
		if !errors.Is(err, outbox.ErrInvalidEvent) {
			t.Fatalf("Enqueue(payload %s) error = %v", payload, err)
		}
	}
}

func newEffectConsumer(t *testing.T, fail bool) outbox.TransactionalConsumer {
	t.Helper()
	statements := []outbox.Statement{{
		Query: `INSERT INTO effects (business_key, calls) VALUES (?, 1)
			ON CONFLICT (business_key) DO UPDATE SET calls = effects.calls + 1`,
		Arguments: []outbox.EventField{outbox.FieldBusinessKey},
	}}
	if fail {
		statements = append(statements, outbox.Statement{Query: `INSERT INTO missing_effect_target (value) VALUES (1)`})
	}
	consumer, err := outbox.NewTransactionalConsumer("projector", statements...)
	if err != nil {
		t.Fatalf("NewTransactionalConsumer() error = %v", err)
	}
	return consumer
}

func reliableEvent(id, businessKey string, at time.Time) outbox.Event {
	return outbox.Event{ID: id, Topic: "orders.changed", BusinessKey: businessKey, Payload: []byte(`{"revision":1}`), OccurredAt: at}
}

func enqueueEvent(t *testing.T, db *database.Database, store *outbox.Store, event outbox.Event) {
	t.Helper()
	if err := db.WithinTx(context.Background(), func(ctx context.Context, tx database.Tx) error {
		_, err := store.Enqueue(ctx, tx, event)
		return err
	}); err != nil {
		t.Fatalf("Enqueue() error = %v", err)
	}
}

func claimEvents(t *testing.T, db *database.Database, store *outbox.Store, options outbox.ClaimOptions) []outbox.Record {
	t.Helper()
	var records []outbox.Record
	if err := db.WithinTx(context.Background(), func(ctx context.Context, tx database.Tx) error {
		var err error
		records, err = store.Claim(ctx, tx, options)
		return err
	}); err != nil {
		t.Fatalf("Claim() error = %v", err)
	}
	return records
}

func assertEffectCalls(t *testing.T, db *database.Database, key string, want int) {
	t.Helper()
	var got int
	err := db.SQL().QueryRow(`SELECT calls FROM effects WHERE business_key = ?`, key).Scan(&got)
	if want == 0 && errors.Is(err, sql.ErrNoRows) {
		return
	}
	if err != nil || got != want {
		t.Fatalf("effect calls = %d, err = %v, want %d", got, err, want)
	}
}
