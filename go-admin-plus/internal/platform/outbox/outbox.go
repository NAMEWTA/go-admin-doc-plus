// Package outbox provides transactional integration-event persistence and recovery.
package outbox

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"time"

	"go-admin/internal/platform/database"
)

type State string

const (
	StatePending   State = "pending"
	StateClaimed   State = "claimed"
	StateDelivered State = "delivered"
	StateRetry     State = "retry"
)

var (
	ErrInvalidEvent        = errors.New("integration event is invalid")
	ErrBusinessKeyConflict = errors.New("integration event business key conflicts with an existing event")
	ErrInvalidClaim        = errors.New("outbox claim input is invalid")
	ErrClaimLost           = errors.New("outbox claim is no longer owned")
	ErrConsumerFailed      = errors.New("outbox consumer failed")
	topicPattern           = regexp.MustCompile(`^[a-z](?:[a-z0-9.-]{0,126}[a-z0-9])?$`)
	keyPattern             = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:/-]{0,254}$`)
)

type Event struct {
	ID          string
	Topic       string
	BusinessKey string
	Payload     []byte
	OccurredAt  time.Time
}

type Record struct {
	Event         Event
	State         State
	Attempts      int
	AvailableAt   time.Time
	ClaimedBy     string
	ClaimUntil    time.Time
	DeliveredAt   time.Time
	LastErrorCode string
}

type Store struct {
	db *database.Database
}

type ClaimOptions struct {
	Owner string
	Now   time.Time
	Lease time.Duration
	Limit int
}

type Consumer interface {
	Name() string
	// Handle writes durable consumer effects through tx. External effects require their own
	// idempotency contract because only database work can share the receipt transaction.
	Handle(context.Context, database.Tx, Event) error
}

func NewStore(db *database.Database) *Store { return &Store{db: db} }

// Enqueue writes through the caller's transaction so domain state and the event share one commit.
// A repeated topic/business key is accepted only when the immutable envelope is identical.
func (s *Store) Enqueue(ctx context.Context, tx database.Tx, event Event) (bool, error) {
	if s == nil || s.db == nil || tx == nil || validateEvent(event) != nil {
		return false, ErrInvalidEvent
	}
	result, err := tx.ExecContext(ctx, `
		INSERT INTO reliable_outbox (
			id, topic, business_key, payload, state, attempts,
			available_at, occurred_at, created_at
		) VALUES (?, ?, ?, ?, ?, 0, ?, ?, ?)
		ON CONFLICT (topic, business_key) DO NOTHING`,
		event.ID, event.Topic, event.BusinessKey, event.Payload, StatePending,
		event.OccurredAt, event.OccurredAt, event.OccurredAt,
	)
	if err != nil {
		return false, stageError(ctx, "integration event enqueue failed", err)
	}
	created, err := result.RowsAffected()
	if err != nil {
		return false, stageError(ctx, "integration event enqueue result failed", err)
	}
	if created == 1 {
		return true, nil
	}
	var existingID string
	var existingPayload []byte
	var existingOccurred time.Time
	err = tx.QueryRowContext(ctx, `
		SELECT id, payload, occurred_at
		FROM reliable_outbox
		WHERE topic = ? AND business_key = ?`, event.Topic, event.BusinessKey,
	).Scan(&existingID, &existingPayload, &existingOccurred)
	if err != nil {
		return false, stageError(ctx, "integration event idempotency check failed", err)
	}
	if existingID != event.ID || !bytes.Equal(existingPayload, event.Payload) || !existingOccurred.Equal(event.OccurredAt) {
		return false, ErrBusinessKeyConflict
	}
	return false, nil
}

// Claim recovers expired leases and exclusively moves a bounded ready batch to claimed.
// It must run in the executor transaction so a lost PostgreSQL lock connection aborts the claim.
func (s *Store) Claim(ctx context.Context, tx database.Tx, options ClaimOptions) ([]Record, error) {
	if s == nil || s.db == nil || tx == nil || !validOwner(options.Owner) ||
		options.Now.IsZero() || options.Now.Location() != time.UTC || options.Lease <= 0 || options.Limit < 1 || options.Limit > 1000 {
		return nil, ErrInvalidClaim
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE reliable_outbox
		SET state = ?, available_at = ?, claimed_by = NULL, claim_until = NULL,
			last_error_code = ?
		WHERE state = ? AND claim_until <= ?`,
		StateRetry, options.Now, "claim_expired", StateClaimed, options.Now,
	); err != nil {
		return nil, stageError(ctx, "outbox expired claim recovery failed", err)
	}

	query := `
		SELECT id, topic, business_key, payload, occurred_at, state, attempts,
			available_at, claimed_by, claim_until, delivered_at, last_error_code
		FROM reliable_outbox
		WHERE state IN (?, ?) AND available_at <= ?
		ORDER BY available_at, created_at, id
		LIMIT ?`
	if s.db.Dialect() == database.DialectPostgres {
		query += ` FOR UPDATE SKIP LOCKED`
	}
	rows, err := tx.QueryContext(ctx, query, StatePending, StateRetry, options.Now, options.Limit)
	if err != nil {
		return nil, stageError(ctx, "outbox claim selection failed", err)
	}
	var candidates []Record
	for rows.Next() {
		record, scanErr := scanRecord(rows)
		if scanErr != nil {
			_ = rows.Close()
			return nil, stageError(ctx, "outbox claim row failed", scanErr)
		}
		candidates = append(candidates, record)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return nil, stageError(ctx, "outbox claim rows failed", err)
	}
	if err := rows.Close(); err != nil {
		return nil, stageError(ctx, "outbox claim rows close failed", err)
	}

	claimed := make([]Record, 0, len(candidates))
	until := options.Now.Add(options.Lease)
	for _, record := range candidates {
		result, updateErr := tx.ExecContext(ctx, `
			UPDATE reliable_outbox
			SET state = ?, attempts = attempts + 1, claimed_by = ?, claim_until = ?
			WHERE id = ? AND state IN (?, ?) AND available_at <= ?`,
			StateClaimed, options.Owner, until, record.Event.ID, StatePending, StateRetry, options.Now,
		)
		if updateErr != nil {
			return nil, stageError(ctx, "outbox claim update failed", updateErr)
		}
		changed, updateErr := result.RowsAffected()
		if updateErr != nil {
			return nil, stageError(ctx, "outbox claim result failed", updateErr)
		}
		if changed == 0 {
			continue
		}
		record.State = StateClaimed
		record.Attempts++
		record.ClaimedBy = options.Owner
		record.ClaimUntil = until
		claimed = append(claimed, record)
	}
	return claimed, nil
}

// Deliver applies a consumer's database effect, durable receipt, and delivered state atomically.
// The returned bool is false when an existing receipt suppressed a replay.
func (s *Store) Deliver(ctx context.Context, tx database.Tx, record Record, consumer Consumer, now time.Time) (bool, error) {
	if s == nil || s.db == nil || tx == nil || consumer == nil || !validOwner(consumer.Name()) ||
		record.State != StateClaimed || !validOwner(record.ClaimedBy) || now.IsZero() || now.Location() != time.UTC {
		return false, ErrInvalidClaim
	}
	var receiptID string
	err := tx.QueryRowContext(ctx, `
		SELECT event_id FROM reliable_consumer_receipt
		WHERE consumer_name = ? AND business_key = ?`, consumer.Name(), record.Event.BusinessKey,
	).Scan(&receiptID)
	if err == nil {
		if err := s.markDelivered(ctx, tx, record, now); err != nil {
			return false, err
		}
		return false, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return false, stageError(ctx, "outbox receipt check failed", err)
	}
	if err := consumer.Handle(ctx, tx, cloneEvent(record.Event)); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return false, stageError(ctx, "outbox consumer canceled", err)
		}
		return false, ErrConsumerFailed
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO reliable_consumer_receipt (consumer_name, business_key, event_id, processed_at)
		VALUES (?, ?, ?, ?)`, consumer.Name(), record.Event.BusinessKey, record.Event.ID, now,
	); err != nil {
		return false, stageError(ctx, "outbox receipt write failed", err)
	}
	if err := s.markDelivered(ctx, tx, record, now); err != nil {
		return false, err
	}
	return true, nil
}

func (s *Store) Retry(ctx context.Context, tx database.Tx, record Record, availableAt time.Time, errorCode string) error {
	if s == nil || s.db == nil || tx == nil || record.State != StateClaimed || !validOwner(record.ClaimedBy) ||
		availableAt.IsZero() || availableAt.Location() != time.UTC || !errorCodePattern.MatchString(errorCode) {
		return ErrInvalidClaim
	}
	result, err := tx.ExecContext(ctx, `
		UPDATE reliable_outbox
		SET state = ?, available_at = ?, claimed_by = NULL, claim_until = NULL, last_error_code = ?
		WHERE id = ? AND state = ? AND claimed_by = ?`,
		StateRetry, availableAt, errorCode, record.Event.ID, StateClaimed, record.ClaimedBy,
	)
	if err != nil {
		return stageError(ctx, "outbox retry update failed", err)
	}
	return requireOneRow(result)
}

func (s *Store) markDelivered(ctx context.Context, tx database.Tx, record Record, now time.Time) error {
	result, err := tx.ExecContext(ctx, `
		UPDATE reliable_outbox
		SET state = ?, delivered_at = ?, claimed_by = NULL, claim_until = NULL, last_error_code = NULL
		WHERE id = ? AND state = ? AND claimed_by = ?`,
		StateDelivered, now, record.Event.ID, StateClaimed, record.ClaimedBy,
	)
	if err != nil {
		return stageError(ctx, "outbox delivery update failed", err)
	}
	return requireOneRow(result)
}

func requireOneRow(result sql.Result) error {
	changed, err := result.RowsAffected()
	if err != nil {
		return errors.New("outbox state result failed")
	}
	if changed != 1 {
		return ErrClaimLost
	}
	return nil
}

func (s *Store) Lookup(ctx context.Context, id string) (Record, bool, error) {
	if s == nil || s.db == nil || id == "" {
		return Record{}, false, errors.New("outbox lookup input is invalid")
	}
	row := s.db.Bun().QueryRowContext(ctx, `
		SELECT id, topic, business_key, payload, occurred_at, state, attempts,
			available_at, claimed_by, claim_until, delivered_at, last_error_code
		FROM reliable_outbox
		WHERE id = ?`, id)
	record, err := scanRecord(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Record{}, false, nil
	}
	if err != nil {
		return Record{}, false, stageError(ctx, "outbox lookup failed", err)
	}
	return record, true, nil
}

var errorCodePattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,62}$`)

func validOwner(value string) bool { return keyPattern.MatchString(value) }

func cloneEvent(event Event) Event {
	event.Payload = bytes.Clone(event.Payload)
	return event
}

func stageError(ctx context.Context, stage string, err error) error {
	if contextErr := ctx.Err(); contextErr != nil {
		return fmt.Errorf("%s: %w", stage, contextErr)
	}
	if errors.Is(err, context.Canceled) {
		return fmt.Errorf("%s: %w", stage, context.Canceled)
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("%s: %w", stage, context.DeadlineExceeded)
	}
	return errors.New(stage)
}

type rowScanner interface {
	Scan(...any) error
}

func scanRecord(row rowScanner) (Record, error) {
	var record Record
	var claimedBy, lastError sql.NullString
	var claimUntil, deliveredAt sql.NullTime
	err := row.Scan(
		&record.Event.ID, &record.Event.Topic, &record.Event.BusinessKey, &record.Event.Payload,
		&record.Event.OccurredAt, &record.State, &record.Attempts, &record.AvailableAt,
		&claimedBy, &claimUntil, &deliveredAt, &lastError,
	)
	if err != nil {
		return Record{}, err
	}
	record.Event.Payload = bytes.Clone(record.Event.Payload)
	record.ClaimedBy = claimedBy.String
	record.ClaimUntil = claimUntil.Time
	record.DeliveredAt = deliveredAt.Time
	record.LastErrorCode = lastError.String
	return record, nil
}

func validateEvent(event Event) error {
	if !keyPattern.MatchString(event.ID) || !topicPattern.MatchString(event.Topic) || !keyPattern.MatchString(event.BusinessKey) {
		return ErrInvalidEvent
	}
	if event.OccurredAt.IsZero() || event.OccurredAt.Location() != time.UTC {
		return ErrInvalidEvent
	}
	if len(event.Payload) == 0 || len(event.Payload) > 1<<20 || !json.Valid(event.Payload) {
		return ErrInvalidEvent
	}
	return nil
}
