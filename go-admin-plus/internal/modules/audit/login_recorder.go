package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"regexp"
	"time"

	"go-admin/internal/platform/database"
)

var stableKeyPart = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,63}$`)
var stableActorRef = regexp.MustCompile(`^account:[a-z0-9][a-z0-9_-]{0,63}$`)

type LoginFact struct {
	EventID    string
	AttemptID  string
	Outcome    Outcome
	ActorType  ActorType
	ActorRef   *string
	Source     Source
	OccurredAt time.Time
}

// LoginRecorder is the Audit-owned synchronous port for login facts that must commit with a caller transaction.
type LoginRecorder struct{}

func NewLoginRecorder(db *database.Database) (*LoginRecorder, error) {
	if db == nil {
		return nil, ErrInvalidArgument
	}
	return &LoginRecorder{}, nil
}

func (*LoginRecorder) Record(ctx context.Context, tx database.Tx, fact LoginFact) (bool, error) {
	if tx == nil || !validFactID(fact.EventID) || !stableKeyPart.MatchString(fact.AttemptID) ||
		fact.Outcome != OutcomeSucceeded && fact.Outcome != OutcomeFailed || !validEnvelope(fact.ActorType, fact.Source) ||
		fact.ActorRef != nil && (fact.ActorType != ActorAccount || !validActorRef(*fact.ActorRef)) || fact.OccurredAt.IsZero() {
		return false, ErrInvalidArgument
	}
	topic := TopicLoginSucceeded
	if fact.Outcome == OutcomeFailed {
		topic = TopicLoginFailed
	}
	payload, err := json.Marshal(struct {
		ActorType ActorType `json:"actorType"`
		Source    Source    `json:"source"`
	}{ActorType: fact.ActorType, Source: fact.Source})
	if err != nil {
		return false, ErrInternal
	}
	row := &storedLoginFact{EventID: fact.EventID, Topic: topic, BusinessKey: "login:" + fact.AttemptID, ActorRef: cloneString(fact.ActorRef), Payload: payload, OccurredAt: fact.OccurredAt.UTC()}
	result, err := tx.NewInsert().Model(row).ModelTableExpr("audit_facts").On("CONFLICT (event_id) DO NOTHING").Exec(ctx)
	if err != nil {
		return false, conceal(err)
	}
	created, err := result.RowsAffected()
	if err != nil {
		return false, ErrInternal
	}
	if created == 1 {
		return true, nil
	}
	var existing storedLoginFact
	if err := tx.NewSelect().Table("audit_facts").Column("event_id", "topic", "business_key", "actor_ref", "payload", "occurred_at").Where("event_id = ?", fact.EventID).Scan(ctx, &existing); err != nil {
		return false, conceal(err)
	}
	if existing.Topic != row.Topic || existing.BusinessKey != row.BusinessKey || !equalStrings(existing.ActorRef, row.ActorRef) || !bytes.Equal(existing.Payload, row.Payload) || !existing.OccurredAt.Equal(row.OccurredAt) {
		return false, ErrInvalidArgument
	}
	return false, nil
}

type storedLoginFact struct {
	EventID     string    `bun:"event_id"`
	Topic       string    `bun:"topic"`
	BusinessKey string    `bun:"business_key"`
	ActorRef    *string   `bun:"actor_ref"`
	Payload     []byte    `bun:"payload"`
	OccurredAt  time.Time `bun:"occurred_at"`
}

func validActorRef(value string) bool { return stableActorRef.MatchString(value) }

func equalStrings(left, right *string) bool {
	return left == nil && right == nil || left != nil && right != nil && *left == *right
}
