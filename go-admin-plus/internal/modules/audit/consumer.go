package audit

import (
	"strings"

	"go-admin/internal/platform/outbox"
)

// TransactionalConsumers builds the only consumers allowed to mutate Audit-owned tables.
func TransactionalConsumers() (map[string]outbox.TransactionalConsumer, error) {
	consumers := make(map[string]outbox.TransactionalConsumer, len(topicDefinitions))
	for topic := range topicDefinitions {
		consumer, err := outbox.NewTransactionalConsumer(
			"audit",
			"audit-projector-"+strings.ReplaceAll(topic, ".", "-"),
			[]string{"audit_facts"},
			outbox.Mutation{
				Operation: outbox.OperationInsert,
				Table:     "audit_facts",
				Values: []outbox.ColumnBinding{
					{Column: "event_id", Field: outbox.FieldEventID},
					{Column: "topic", Field: outbox.FieldTopic},
					{Column: "business_key", Field: outbox.FieldBusinessKey},
					{Column: "payload", Field: outbox.FieldPayload},
					{Column: "occurred_at", Field: outbox.FieldOccurredAt},
				},
				ExpectExactly: 1,
			},
		)
		if err != nil {
			return nil, err
		}
		consumers[topic] = consumer
	}
	return consumers, nil
}
