package audit

import (
	"testing"
	"time"
)

func TestPresentAcceptsLegacyOfflineBootstrapAndRecoveryFacts(t *testing.T) {
	occurredAt := time.Date(2026, 9, 4, 10, 0, 0, 0, time.UTC)
	bootstrap, err := present(storedFact{
		Topic:       TopicOperationCreated,
		BusinessKey: "resource:iam_bootstrap:account-00000001",
		Payload:     []byte(`{"action":"bootstrap"}`),
		OccurredAt:  occurredAt,
	})
	if err != nil {
		t.Fatalf("legacy bootstrap fact was rejected: %v", err)
	}
	if bootstrap.Source != SourceServer || bootstrap.ActorType != ActorSystem || bootstrap.Subject != "iam_bootstrap:account-00000001" {
		t.Fatalf("legacy bootstrap fact = %#v", bootstrap)
	}

	recovery, err := present(storedFact{
		Topic:       TopicOperationUpdated,
		BusinessKey: "resource:iam_recovery:account-00000001:1756980000000000000",
		ActorRef:    stringPtr("account:account-00000001"),
		Payload:     []byte(`{"action":"recover-admin","reason":"lost-access"}`),
		OccurredAt:  occurredAt,
	})
	if err != nil {
		t.Fatalf("legacy recovery fact was rejected: %v", err)
	}
	if recovery.Source != SourceServer || recovery.ActorType != ActorAccount || recovery.Subject != "iam_recovery:account-00000001" || recovery.ActorRef == nil || *recovery.ActorRef != "account:account-00000001" {
		t.Fatalf("legacy recovery fact = %#v", recovery)
	}
}

func stringPtr(value string) *string { return &value }
