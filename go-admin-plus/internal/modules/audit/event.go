package audit

import "go-admin/internal/platform/outbox"

const (
	TopicLoginSucceeded   = "iam.login.succeeded"
	TopicLoginFailed      = "iam.login.failed"
	TopicOperationCreated = "operation.created"
	TopicOperationUpdated = "operation.updated"
	TopicOperationDeleted = "operation.deleted"
)

type Kind string

const (
	KindLogin     Kind = "login"
	KindOperation Kind = "operation"
)

type Outcome string

const (
	OutcomeSucceeded Outcome = "succeeded"
	OutcomeFailed    Outcome = "failed"
)

type ActorType string

const (
	ActorAccount ActorType = "account"
	ActorSystem  ActorType = "system"
)

type Source string

const (
	SourceWeb     Source = "web"
	SourceDesktop Source = "desktop"
	SourceServer  Source = "server"
)

type topicDefinition struct {
	kind      Kind
	action    string
	outcome   Outcome
	keyPrefix string
	minParts  int
	maxParts  int
}

var topicDefinitions = map[string]topicDefinition{
	TopicLoginSucceeded:   {kind: KindLogin, action: "login", outcome: OutcomeSucceeded, keyPrefix: "login", minParts: 1, maxParts: 2},
	TopicLoginFailed:      {kind: KindLogin, action: "login", outcome: OutcomeFailed, keyPrefix: "login", minParts: 1, maxParts: 2},
	TopicOperationCreated: {kind: KindOperation, action: "create", outcome: OutcomeSucceeded, keyPrefix: "resource", minParts: 4, maxParts: 5},
	TopicOperationUpdated: {kind: KindOperation, action: "update", outcome: OutcomeSucceeded, keyPrefix: "resource", minParts: 4, maxParts: 5},
	TopicOperationDeleted: {kind: KindOperation, action: "delete", outcome: OutcomeSucceeded, keyPrefix: "resource", minParts: 4, maxParts: 5},
}

// TopicSchemas returns a closed, non-sensitive contract for every event Audit consumes.
func TopicSchemas() []outbox.TopicSchema {
	result := make([]outbox.TopicSchema, 0, 3)
	for _, topic := range []string{TopicOperationCreated, TopicOperationUpdated, TopicOperationDeleted} {
		definition := topicDefinitions[topic]
		payload := []outbox.PayloadFieldSchema{
			{Name: "source", Kind: outbox.PayloadString, Required: true, AllowedStrings: []string{string(SourceWeb), string(SourceDesktop), string(SourceServer)}},
		}
		result = append(result, outbox.TopicSchema{
			Topic:       topic,
			Payload:     payload,
			BusinessKey: outbox.BusinessKeySchema{Prefix: definition.keyPrefix, MinParts: definition.minParts, MaxParts: definition.maxParts},
		})
	}
	return result
}
