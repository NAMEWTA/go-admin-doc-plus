package audit

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"regexp"
	"strings"
	"time"

	"go-admin/internal/modules/iam/session"
	"go-admin/internal/platform/database"
)

var stableKeyPart = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,63}$`)
var stableActorRef = regexp.MustCompile(`^account:[a-z0-9][a-z0-9_-]{7,63}$`)
var opaqueLoginBusinessID = regexp.MustCompile(`^[a-f0-9]{32}$`)

type LoginFact struct {
	Outcome    Outcome
	ActorType  ActorType
	ActorRef   *string
	Source     Source
	OccurredAt time.Time
}

// LoginRecorder is the Audit-owned synchronous port for login facts that must commit with a caller transaction.
type LoginRecorder struct {
	newBusinessID func() (string, error)
}

func NewLoginRecorder(db *database.Database) (*LoginRecorder, error) {
	if db == nil {
		return nil, ErrInvalidArgument
	}
	return &LoginRecorder{newBusinessID: newLoginBusinessID}, nil
}

func (recorder *LoginRecorder) Record(ctx context.Context, tx database.Tx, fact LoginFact) (bool, error) {
	if recorder == nil || recorder.newBusinessID == nil {
		return false, ErrInvalidArgument
	}
	businessID, err := recorder.newBusinessID()
	if err != nil {
		return false, ErrInternal
	}
	return recorder.record(ctx, tx, businessID, fact)
}

func (recorder *LoginRecorder) record(ctx context.Context, tx database.Tx, businessID string, fact LoginFact) (bool, error) {
	if recorder == nil || tx == nil || !opaqueLoginBusinessID.MatchString(businessID) || fact.ActorType != ActorAccount ||
		fact.Outcome != OutcomeSucceeded && fact.Outcome != OutcomeFailed ||
		fact.Outcome == OutcomeSucceeded && (fact.ActorRef == nil || !validActorRef(*fact.ActorRef)) ||
		fact.Outcome == OutcomeFailed && fact.ActorRef != nil ||
		fact.Source != SourceWeb && fact.Source != SourceDesktop && fact.Source != SourceServer || fact.OccurredAt.IsZero() {
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
	row := &storedLoginFact{Topic: topic, BusinessKey: "login:" + businessID, ActorRef: cloneString(fact.ActorRef), Payload: payload, OccurredAt: fact.OccurredAt.UTC()}
	result, err := tx.NewInsert().Model(row).ModelTableExpr("audit_facts").Exec(ctx)
	if err != nil {
		return false, conceal(err)
	}
	created, err := result.RowsAffected()
	if err != nil {
		return false, ErrInternal
	}
	if created != 1 {
		return false, ErrInternal
	}
	return true, nil
}

// SessionLoginFactAdapter is the Audit-owned implementation of Session's module-neutral port.
// Session creates the opaque attempt ID and never imports Audit.
type SessionLoginFactAdapter struct{ recorder *LoginRecorder }

func NewSessionLoginFactAdapter(db *database.Database) (*SessionLoginFactAdapter, error) {
	recorder, err := NewLoginRecorder(db)
	if err != nil {
		return nil, err
	}
	return &SessionLoginFactAdapter{recorder: recorder}, nil
}

func (adapter *SessionLoginFactAdapter) RecordLoginFact(ctx context.Context, tx database.Tx, fact session.LoginFact) error {
	if adapter == nil || adapter.recorder == nil || !fact.AttemptID.Valid() {
		return ErrInvalidArgument
	}
	mapped := LoginFact{ActorType: ActorAccount, Source: Source(fact.Source), OccurredAt: fact.OccurredAt}
	switch fact.Outcome {
	case session.LoginSucceeded:
		actorRef := "account:" + fact.AccountID
		mapped.Outcome = OutcomeSucceeded
		mapped.ActorRef = &actorRef
	case session.LoginFailed:
		if fact.AccountID != "" {
			return ErrInvalidArgument
		}
		mapped.Outcome = OutcomeFailed
	default:
		return ErrInvalidArgument
	}
	_, err := adapter.recorder.record(ctx, tx, fact.AttemptID.Opaque(), mapped)
	return err
}

type storedLoginFact struct {
	Topic       string    `bun:"topic"`
	BusinessKey string    `bun:"business_key"`
	ActorRef    *string   `bun:"actor_ref"`
	Payload     []byte    `bun:"payload"`
	OccurredAt  time.Time `bun:"occurred_at"`
}

func validActorRef(value string) bool {
	if !stableActorRef.MatchString(value) {
		return false
	}
	lower := strings.ToLower(value)
	for _, sensitive := range []string{"password", "secret", "session", "token", "credential"} {
		if strings.Contains(lower, sensitive) {
			return false
		}
	}
	return true
}

func validOperationActorID(value string) bool {
	if !stableKeyPart.MatchString(value) {
		return false
	}
	lower := strings.ToLower(value)
	for _, sensitive := range []string{"password", "secret", "session", "token", "credential"} {
		if strings.Contains(lower, sensitive) {
			return false
		}
	}
	return true
}

func newLoginBusinessID() (string, error) {
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return hex.EncodeToString(raw), nil
}
