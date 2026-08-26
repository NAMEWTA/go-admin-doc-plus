package audit_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	audit "go-admin/internal/modules/audit"
	auditmigration "go-admin/internal/modules/audit/migrations/0011-audit"
	"go-admin/internal/platform/config"
	"go-admin/internal/platform/coordination"
	"go-admin/internal/platform/database"
	"go-admin/internal/platform/migrations"
	reliablemigration "go-admin/internal/platform/migrations/reliable-runtime"
	"go-admin/internal/platform/outbox"
)

func TestIntegrationEventProjectsOneCanonicalAuditFact(t *testing.T) {
	db := openSQLite(t)
	migrate(t, db, reliablemigration.Provider{}, auditmigration.Provider{})
	store := newAuditStore(t, db)
	consumers := mustConsumers(t)
	now := time.Date(2026, 8, 27, 2, 0, 0, 0, time.UTC)
	event := outbox.Event{
		ID: "raw-session-secret", Topic: audit.TopicOperationCreated,
		BusinessKey: "resource:demo:record-0001:revision-1:account:account-00000001", Payload: []byte(`{"source":"web"}`), OccurredAt: now,
	}
	enqueue(t, db, store, event)
	dispatch(t, db, store, consumers, now)

	service := mustService(t, db, allowAll{})
	page, err := service.List(context.Background(), audit.Principal{ID: "auditor-00000001"}, audit.Filter{Page: 1, PageSize: 20})
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 1 || len(page.Records) != 1 {
		t.Fatalf("audit page = %#v", page)
	}
	fact := page.Records[0]
	if fact.ID == event.ID || strings.Contains(fact.ID, "session") || fact.Kind != audit.KindOperation || fact.Action != "create" || fact.Outcome != audit.OutcomeSucceeded || fact.Source != audit.SourceWeb || fact.ActorType != audit.ActorAccount || fact.Subject != "demo:record-0001" || fact.ActorRef == nil || *fact.ActorRef != "account:account-00000001" {
		t.Fatalf("projected fact = %#v", fact)
	}
	var persisted string
	if err := db.Bun().QueryRowContext(context.Background(), "SELECT topic || ':' || business_key || ':' || CAST(payload AS TEXT) FROM audit_facts").Scan(&persisted); err != nil || strings.Contains(persisted, event.ID) {
		t.Fatalf("Audit persisted unsafe event id: %q, %v", persisted, err)
	}
	encoded, _ := json.Marshal(page)
	if strings.Contains(string(encoded), event.ID) {
		t.Fatalf("Audit JSON exposed unsafe event id: %s", encoded)
	}

	created, err := dbEnqueue(context.Background(), db, store, event)
	if err != nil || created {
		t.Fatalf("idempotent enqueue = %v, %v", created, err)
	}
	dispatch(t, db, store, consumers, now.Add(time.Second))
	page, err = service.List(context.Background(), audit.Principal{ID: "auditor-00000001"}, audit.Filter{Page: 1, PageSize: 20})
	if err != nil || page.Total != 1 {
		t.Fatalf("idempotent projection = %#v, %v", page, err)
	}
}

func TestAuditEventSchemaRejectsSensitiveAmbiguousAndUnboundedPayloads(t *testing.T) {
	db := openSQLite(t)
	migrate(t, db, reliablemigration.Provider{})
	store := newAuditStore(t, db)
	now := time.Date(2026, 8, 27, 2, 0, 0, 0, time.UTC)
	tests := map[string]string{
		"password":            `{"source":"web","password":"private"}`,
		"raw session":         `{"source":"web","rawSession":"private"}`,
		"nested request body": `{"source":"web","requestBody":{"displayName":"private"}}`,
		"sensitive value":     `{"source":"raw-session-secret"}`,
		"duplicate member":    `{"source":"server","source":"web"}`,
		"array":               `{"source":["web"]}`,
	}
	for name, payload := range tests {
		t.Run(name, func(t *testing.T) {
			event := outbox.Event{ID: "audit-rejected-0001", Topic: audit.TopicOperationUpdated, BusinessKey: "resource:demo:rejected-0001:revision-1:account:account-00000001", Payload: []byte(payload), OccurredAt: now}
			created, err := dbEnqueue(context.Background(), db, store, event)
			if created || !errors.Is(err, outbox.ErrInvalidEvent) {
				t.Fatalf("sensitive event = %v, %v", created, err)
			}
		})
	}
	loginEvent := outbox.Event{ID: "audit-login-bypass", Topic: audit.TopicLoginSucceeded, BusinessKey: "login:bypass", Payload: []byte(`{"actorType":"account","source":"web"}`), OccurredAt: now}
	if created, err := dbEnqueue(context.Background(), db, store, loginEvent); created || !errors.Is(err, outbox.ErrInvalidEvent) {
		t.Fatalf("async login bypass = %v, %v", created, err)
	}
}

func TestFailedProjectionRetriesAndConvergesWithoutDuplicates(t *testing.T) {
	db := openSQLite(t)
	migrate(t, db, reliablemigration.Provider{})
	store := newAuditStore(t, db)
	consumers := mustConsumers(t)
	now := time.Date(2026, 8, 27, 2, 0, 0, 0, time.UTC)
	event := outbox.Event{ID: "audit-retry-0001", Topic: audit.TopicOperationUpdated, BusinessKey: "resource:demo:record-0001:revision-2:system", Payload: []byte(`{"source":"server"}`), OccurredAt: now}
	enqueue(t, db, store, event)

	result := dispatch(t, db, store, consumers, now)
	if result.Retried != 1 || result.Delivered != 0 {
		t.Fatalf("first dispatch = %#v", result)
	}
	migrate(t, db, auditmigration.Provider{})
	result = dispatch(t, db, store, consumers, now.Add(time.Minute))
	if result.Delivered != 1 || result.Retried != 0 {
		t.Fatalf("recovery dispatch = %#v", result)
	}

	service := mustService(t, db, allowAll{})
	page, err := service.List(context.Background(), audit.Principal{ID: "auditor-00000001"}, audit.Filter{Page: 1, PageSize: 20})
	if err != nil || page.Total != 1 || page.Records[0].Action != "update" {
		t.Fatalf("recovered facts = %#v, %v", page, err)
	}
}

func TestQueryDetailAndCleanupAreAuthorizedFilteredAndBounded(t *testing.T) {
	db := openSQLite(t)
	migrate(t, db, reliablemigration.Provider{}, auditmigration.Provider{})
	store := newAuditStore(t, db)
	consumers := mustConsumers(t)
	now := time.Date(2026, 8, 27, 6, 0, 0, 0, time.UTC)
	recorder, err := audit.NewLoginRecorder(db)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.WithinTx(context.Background(), func(ctx context.Context, tx database.Tx) error {
		_, recordErr := recorder.Record(ctx, tx, audit.LoginFact{Outcome: audit.OutcomeFailed, ActorType: audit.ActorAccount, Source: audit.SourceDesktop, OccurredAt: now.Add(-90 * 24 * time.Hour)})
		return recordErr
	}); err != nil {
		t.Fatal(err)
	}
	enqueue(t, db, store, outbox.Event{ID: "audit-query-00002", Topic: audit.TopicOperationCreated, BusinessKey: "resource:demo:record-0002:revision-1:system", Payload: []byte(`{"source":"server"}`), OccurredAt: now.Add(-60 * 24 * time.Hour)})
	enqueue(t, db, store, outbox.Event{ID: "audit-query-00003", Topic: audit.TopicOperationUpdated, BusinessKey: "resource:demo:record-0003:revision-2:account:account-00000003", Payload: []byte(`{"source":"web"}`), OccurredAt: now.Add(-24 * time.Hour)})
	dispatch(t, db, store, consumers, now)
	if _, err := db.Bun().ExecContext(context.Background(), "UPDATE audit_facts SET payload = ? WHERE topic = ?", []byte(`{"source":"desktop","actorType":"account"}`), audit.TopicLoginFailed); err != nil {
		t.Fatal(err)
	}

	observer := &recordingObserver{}
	service := mustServiceWithPolicy(t, db, allowAll{}, audit.RetentionPolicy{MinimumAge: 30 * 24 * time.Hour, CleanupLimit: 1, Now: func() time.Time { return now }, Observer: observer})
	principal := audit.Principal{ID: "auditor-00000001"}
	page, err := service.List(context.Background(), principal, audit.Filter{Page: 1, PageSize: 20, Kind: audit.KindLogin, Outcome: audit.OutcomeFailed, Source: audit.SourceDesktop})
	if err != nil || page.Total != 1 || !regexp.MustCompile(`^login:[a-f0-9]{32}$`).MatchString(page.Records[0].Subject) {
		t.Fatalf("filtered page = %#v, %v", page, err)
	}
	all, err := service.List(context.Background(), principal, audit.Filter{Page: 1, PageSize: 20})
	if err != nil {
		t.Fatal(err)
	}
	createdID := findFactID(t, all, "demo:record-0002")
	updatedID := findFactID(t, all, "demo:record-0003")
	fact, err := service.Detail(context.Background(), principal, createdID)
	if err != nil || fact.Action != "create" || fact.Source != audit.SourceServer || fact.ActorType != audit.ActorSystem || fact.ActorRef != nil {
		t.Fatalf("detail = %#v, %v", fact, err)
	}
	accountFact, err := service.Detail(context.Background(), principal, updatedID)
	if err != nil || accountFact.ActorType != audit.ActorAccount || accountFact.ActorRef == nil || *accountFact.ActorRef != "account:account-00000003" {
		t.Fatalf("account operation actor = %#v, %v", accountFact, err)
	}
	if _, err := service.List(context.Background(), principal, audit.Filter{Page: 1, PageSize: 20, Kind: "unknown"}); !errors.Is(err, audit.ErrInvalidArgument) {
		t.Fatalf("unknown filter = %v", err)
	}

	denied := mustServiceWithPolicy(t, db, denyAll{}, audit.RetentionPolicy{MinimumAge: 30 * 24 * time.Hour, CleanupLimit: 1, Now: func() time.Time { return now }})
	if _, err := denied.Cleanup(context.Background(), principal, audit.CleanupCommand{Before: now.Add(-45 * 24 * time.Hour), Confirmation: audit.CleanupConfirmation}); !errors.Is(err, audit.ErrForbidden) {
		t.Fatalf("denied cleanup = %v", err)
	}
	if _, err := service.Cleanup(context.Background(), principal, audit.CleanupCommand{Before: now.Add(-45 * 24 * time.Hour), Confirmation: "delete everything"}); !errors.Is(err, audit.ErrInvalidArgument) {
		t.Fatalf("unconfirmed cleanup = %v", err)
	}
	if _, err := service.Cleanup(context.Background(), principal, audit.CleanupCommand{Before: now.Add(-24 * time.Hour), Confirmation: audit.CleanupConfirmation}); !errors.Is(err, audit.ErrRetentionProtected) {
		t.Fatalf("retention cleanup = %v", err)
	}
	result, err := service.Cleanup(context.Background(), principal, audit.CleanupCommand{Before: now.Add(-45 * 24 * time.Hour), Confirmation: audit.CleanupConfirmation})
	if err != nil || result.Deleted != 1 || !result.MoreEligible {
		t.Fatalf("bounded cleanup = %#v, %v", result, err)
	}
	page, err = service.List(context.Background(), principal, audit.Filter{Page: 1, PageSize: 20})
	if err != nil || page.Total != 2 {
		t.Fatalf("after cleanup = %#v, %v", page, err)
	}
	if observations := observer.snapshot(); len(observations) == 0 {
		t.Fatal("audit lifecycle was not observable")
	} else if raw, _ := json.Marshal(observations); strings.Contains(string(raw), "audit-query") || strings.Contains(string(raw), "record-000") {
		t.Fatalf("observer leaked identity: %s", raw)
	}
}

func TestSynchronousLoginPortPersistsTypedFactInCallerTransaction(t *testing.T) {
	db := openSQLite(t)
	migrate(t, db, auditmigration.Provider{})
	recorder, err := audit.NewLoginRecorder(db)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 8, 27, 7, 0, 0, 0, time.UTC)
	err = db.WithinTx(context.Background(), func(ctx context.Context, tx database.Tx) error {
		created, err := recorder.Record(ctx, tx, audit.LoginFact{Outcome: audit.OutcomeSucceeded, ActorType: audit.ActorAccount, ActorRef: stringPointer("account:account-00000003"), Source: audit.SourceWeb, OccurredAt: now})
		if err != nil || !created {
			return errors.New("login record failed")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	page, err := mustService(t, db, allowAll{}).List(context.Background(), audit.Principal{ID: "auditor-00000001"}, audit.Filter{Page: 1, PageSize: 20})
	if err != nil || page.Total != 1 || page.Records[0].Outcome != audit.OutcomeSucceeded || page.Records[0].ActorRef == nil || *page.Records[0].ActorRef != "account:account-00000003" ||
		!regexp.MustCompile(`^a1\.ls\.[A-Za-z0-9_-]+$`).MatchString(page.Records[0].ID) || !regexp.MustCompile(`^login:[a-f0-9]{32}$`).MatchString(page.Records[0].Subject) {
		t.Fatalf("synchronous fact = %#v, %v", page, err)
	}
}

func TestSynchronousLoginPortRejectsSensitiveActorReferencesWithoutWriting(t *testing.T) {
	db := openSQLite(t)
	migrate(t, db, auditmigration.Provider{})
	recorder, err := audit.NewLoginRecorder(db)
	if err != nil {
		t.Fatal(err)
	}
	for _, actorRef := range []string{
		"account:password000000", "account:secret00000000", "account:session0000000",
		"account:token000000000", "account:credential0000", "account:admin",
	} {
		err := db.WithinTx(context.Background(), func(ctx context.Context, tx database.Tx) error {
			_, recordErr := recorder.Record(ctx, tx, audit.LoginFact{Outcome: audit.OutcomeSucceeded, ActorType: audit.ActorAccount, ActorRef: &actorRef, Source: audit.SourceWeb, OccurredAt: time.Now().UTC()})
			return recordErr
		})
		if !errors.Is(err, audit.ErrInvalidArgument) {
			t.Fatalf("actor ref %q error = %v", actorRef, err)
		}
	}
	validActor := "account:account-00000003"
	for _, fact := range []audit.LoginFact{
		{Outcome: audit.OutcomeSucceeded, ActorType: audit.ActorAccount, Source: audit.SourceWeb, OccurredAt: time.Now().UTC()},
		{Outcome: audit.OutcomeFailed, ActorType: audit.ActorAccount, ActorRef: &validActor, Source: audit.SourceWeb, OccurredAt: time.Now().UTC()},
		{Outcome: audit.OutcomeSucceeded, ActorType: audit.ActorSystem, ActorRef: &validActor, Source: audit.SourceServer, OccurredAt: time.Now().UTC()},
		{Outcome: audit.OutcomeFailed, ActorType: audit.ActorSystem, Source: audit.SourceServer, OccurredAt: time.Now().UTC()},
	} {
		err := db.WithinTx(context.Background(), func(ctx context.Context, tx database.Tx) error {
			_, recordErr := recorder.Record(ctx, tx, fact)
			return recordErr
		})
		if !errors.Is(err, audit.ErrInvalidArgument) {
			t.Fatalf("invalid login actor invariant error = %v", err)
		}
	}
	if _, err := db.Bun().ExecContext(context.Background(), `INSERT INTO audit_facts (topic, business_key, actor_ref, payload, occurred_at) VALUES (?, ?, ?, ?, ?)`, audit.TopicLoginSucceeded, "login:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "account:x", []byte(`{"actorType":"account","source":"web"}`), time.Now().UTC()); err == nil {
		t.Fatal("SQLite accepted a short non-opaque actor reference")
	}
	if _, err := db.Bun().ExecContext(context.Background(), `INSERT INTO audit_facts (topic, business_key, payload, occurred_at) VALUES (?, ?, ?, ?)`, audit.TopicOperationUpdated, "resource:demo:invalid:revision-1:system:account-00000001", []byte(`{"source":"web"}`), time.Now().UTC()); err == nil {
		t.Fatal("SQLite accepted an ambiguous operation actor")
	}
	var count int
	if err := db.Bun().QueryRowContext(context.Background(), "SELECT COUNT(*) FROM audit_facts").Scan(&count); err != nil || count != 0 {
		t.Fatalf("rejected login facts = %d, %v", count, err)
	}
	page, err := mustService(t, db, allowAll{}).List(context.Background(), audit.Principal{ID: "auditor-00000001"}, audit.Filter{Page: 1, PageSize: 20})
	encoded, _ := json.Marshal(page)
	if err != nil || page.Total != 0 || regexp.MustCompile(`(?i)password|secret|session|token|credential`).Match(encoded) {
		t.Fatalf("rejected login facts reached list JSON: %s, %v", encoded, err)
	}
}

func TestHTTPTransportUsesGeneratedContractAndNeverReturnsStoredEnvelope(t *testing.T) {
	db := openSQLite(t)
	migrate(t, db, auditmigration.Provider{})
	recorder, _ := audit.NewLoginRecorder(db)
	_ = db.WithinTx(context.Background(), func(ctx context.Context, tx database.Tx) error {
		_, err := recorder.Record(ctx, tx, audit.LoginFact{Outcome: audit.OutcomeFailed, ActorType: audit.ActorAccount, Source: audit.SourceWeb, OccurredAt: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)})
		return err
	})
	service := mustServiceWithPolicy(t, db, allowAll{}, audit.RetentionPolicy{MinimumAge: 30 * 24 * time.Hour, CleanupLimit: 10, Now: func() time.Time { return time.Date(2026, 8, 27, 0, 0, 0, 0, time.UTC) }})
	cookie := sessionCookie(strings.Repeat("r", 43))
	authorized, err := audit.NewAuthorizedRequest(audit.Principal{ID: "auditor-00000001"}, strings.Repeat("c", 43), &cookie)
	if err != nil {
		t.Fatal(err)
	}
	handler, err := audit.NewHTTPHandler(service, requestAuthorizer(func(_ context.Context, request *http.Request) (audit.AuthorizedRequest, audit.RequestFailure) {
		if request.Header.Get("X-Test-Identity") != "auditor" {
			return audit.AuthorizedRequest{}, audit.RequestAuthenticationFailed
		}
		return authorized, audit.RequestAuthorized
	}), func(*http.Request) string { return "0123456789abcdef" })
	if err != nil {
		t.Fatal(err)
	}

	request := httptest.NewRequest(http.MethodGet, "/audit/records?page=1&pageSize=20&source=web", nil)
	request.Header.Set("X-Test-Identity", "auditor")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK || response.Header().Get("Cache-Control") != "no-store" || response.Header().Get("X-CSRF-Token") != strings.Repeat("c", 43) || response.Header().Get("Set-Cookie") != cookie {
		t.Fatalf("list response = %d, %s", response.Code, response.Body.String())
	}
	body := response.Body.String()
	if strings.Contains(body, "payload") || strings.Contains(body, "businessKey") || strings.Contains(body, "attempt-http") || !strings.Contains(body, `"subject":"login:`) {
		t.Fatalf("HTTP response leaked stored envelope: %s", body)
	}

	request = httptest.NewRequest(http.MethodGet, "/audit/records?page=1&pageSize=20", nil)
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized || strings.Contains(response.Body.String(), "private session material") {
		t.Fatalf("authentication response = %d, %s", response.Code, response.Body.String())
	}
}

func TestHTTPAuthorizationFailuresAreCategorizedAndRotationMaterialIsRedacted(t *testing.T) {
	db := openSQLite(t)
	migrate(t, db, auditmigration.Provider{})
	recorder, _ := audit.NewLoginRecorder(db)
	_ = db.WithinTx(context.Background(), func(ctx context.Context, tx database.Tx) error {
		_, err := recorder.Record(ctx, tx, audit.LoginFact{Outcome: audit.OutcomeSucceeded, ActorType: audit.ActorAccount, ActorRef: stringPointer("account:account-00000001"), Source: audit.SourceWeb, OccurredAt: time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)})
		return err
	})
	now := time.Date(2026, 8, 27, 0, 0, 0, 0, time.UTC)
	csrf := strings.Repeat("z", 43)
	cookie := sessionCookie(strings.Repeat("p", 43))
	authorized, err := audit.NewAuthorizedRequest(audit.Principal{ID: "auditor-00000001"}, csrf, &cookie)
	if err != nil {
		t.Fatal(err)
	}
	serialized, _ := json.Marshal(authorized)
	var logOutput bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logOutput, nil))
	logger.Info("authorized", "request", authorized)
	formatted := fmt.Sprintf("%s %#v %s", serialized, authorized, logOutput.String())
	if strings.Contains(formatted, csrf) || strings.Contains(formatted, cookie) {
		t.Fatalf("authorized request leaked rotation material: %s", formatted)
	}

	provider := requestAuthorizer(func(_ context.Context, request *http.Request) (audit.AuthorizedRequest, audit.RequestFailure) {
		switch request.Header.Get("X-Test-Failure") {
		case "authentication":
			return audit.AuthorizedRequest{}, audit.RequestAuthenticationFailed
		case "authorization":
			return authorized, audit.RequestAuthorizationFailed
		case "internal":
			return audit.AuthorizedRequest{}, audit.RequestInternalFailed
		default:
			return authorized, audit.RequestAuthorized
		}
	})
	service := mustServiceWithPolicy(t, db, allowAll{}, audit.RetentionPolicy{MinimumAge: 30 * 24 * time.Hour, CleanupLimit: 10, Now: func() time.Time { return now }})
	handler, err := audit.NewHTTPHandler(service, provider, func(*http.Request) string { return "0123456789abcdef" })
	if err != nil {
		t.Fatal(err)
	}

	assertResponse := func(method, target, body, failure string, status int, expectRotation bool) string {
		t.Helper()
		request := httptest.NewRequest(method, target, strings.NewReader(body))
		request.Header.Set("X-Test-Failure", failure)
		if body != "" {
			request.Header.Set("Content-Type", "application/json")
		}
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != status {
			t.Fatalf("%s response = %d, %s", failure, response.Code, response.Body.String())
		}
		if got := response.Header().Get("X-CSRF-Token"); expectRotation && got != csrf || !expectRotation && got != "" {
			t.Fatalf("%s CSRF header = %q", failure, got)
		}
		if strings.Contains(response.Body.String(), cookie) || strings.Contains(response.Body.String(), "private") {
			t.Fatalf("%s leaked private material: %s", failure, response.Body.String())
		}
		return response.Body.String()
	}
	assertResponse(http.MethodGet, "/audit/records?page=1&pageSize=20", "", "authentication", http.StatusUnauthorized, false)
	assertResponse(http.MethodGet, "/audit/records?page=1&pageSize=20", "", "internal", http.StatusInternalServerError, false)
	csrfBody := assertResponse(http.MethodPost, "/audit/records/cleanup", `{"before":"2026-06-01T00:00:00Z","confirmation":"delete-expired-audit-records"}`, "authorization", http.StatusForbidden, true)
	if !strings.Contains(csrfBody, `"code":"CSRF_REJECTED"`) {
		t.Fatalf("CSRF rejection code = %s", csrfBody)
	}
	page, err := service.List(context.Background(), audit.Principal{ID: "auditor-00000001"}, audit.Filter{Page: 1, PageSize: 20})
	if err != nil || page.Total != 1 {
		t.Fatalf("CSRF rejection mutated audit facts: %#v, %v", page, err)
	}
	assertResponse(http.MethodPost, "/audit/records/cleanup", `{"before":"2026-08-26T00:00:00Z","confirmation":"delete-expired-audit-records"}`, "", http.StatusConflict, true)

	deniedHandler, err := audit.NewHTTPHandler(mustServiceWithPolicy(t, db, denyAll{}, audit.RetentionPolicy{MinimumAge: 30 * 24 * time.Hour, CleanupLimit: 10, Now: func() time.Time { return now }}), provider, func(*http.Request) string { return "0123456789abcdef" })
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodGet, "/audit/records?page=1&pageSize=20", nil)
	response := httptest.NewRecorder()
	deniedHandler.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden || response.Header().Get("X-CSRF-Token") != csrf {
		t.Fatalf("service authorization response = %d, %q", response.Code, response.Header().Get("X-CSRF-Token"))
	}

	if _, err := db.Bun().ExecContext(context.Background(), "DROP TABLE audit_facts"); err != nil {
		t.Fatal(err)
	}
	assertResponse(http.MethodGet, "/audit/records?page=1&pageSize=20", "", "", http.StatusInternalServerError, true)
}

func TestPolicyDenyIsForbiddenButPolicyFaultsAreInternal(t *testing.T) {
	db := openSQLite(t)
	migrate(t, db, auditmigration.Provider{})
	now := time.Date(2026, 8, 27, 0, 0, 0, 0, time.UTC)
	principal := audit.Principal{ID: "auditor-00000001"}
	operations := []struct {
		name string
		run  func(*audit.Service) error
	}{
		{name: "list", run: func(service *audit.Service) error {
			_, err := service.List(context.Background(), principal, audit.Filter{Page: 1, PageSize: 20})
			return err
		}},
		{name: "detail", run: func(service *audit.Service) error {
			_, err := service.Detail(context.Background(), principal, testFactID("oc", "resource:demo:policy-0001:revision-1:system"))
			return err
		}},
		{name: "cleanup", run: func(service *audit.Service) error {
			_, err := service.Cleanup(context.Background(), principal, audit.CleanupCommand{Before: now.Add(-60 * 24 * time.Hour), Confirmation: audit.CleanupConfirmation})
			return err
		}},
	}
	for _, operation := range operations {
		t.Run(operation.name+" deny", func(t *testing.T) {
			service := mustServiceWithPolicy(t, db, denyAll{}, audit.RetentionPolicy{MinimumAge: 30 * 24 * time.Hour, CleanupLimit: 10, Now: func() time.Time { return now }})
			if err := operation.run(service); !errors.Is(err, audit.ErrForbidden) {
				t.Fatalf("deny error = %v", err)
			}
		})
		t.Run(operation.name+" fault", func(t *testing.T) {
			service := mustServiceWithPolicy(t, db, faultyAuthorizer{err: errors.New("private policy database detail")}, audit.RetentionPolicy{MinimumAge: 30 * 24 * time.Hour, CleanupLimit: 10, Now: func() time.Time { return now }})
			if err := operation.run(service); !errors.Is(err, audit.ErrInternal) || strings.Contains(err.Error(), "private") {
				t.Fatalf("fault error = %v", err)
			}
		})
		t.Run(operation.name+" timeout", func(t *testing.T) {
			service := mustServiceWithPolicy(t, db, faultyAuthorizer{err: context.DeadlineExceeded}, audit.RetentionPolicy{MinimumAge: 30 * 24 * time.Hour, CleanupLimit: 10, Now: func() time.Time { return now }})
			if err := operation.run(service); !errors.Is(err, audit.ErrInternal) || !errors.Is(err, context.DeadlineExceeded) {
				t.Fatalf("timeout error = %v", err)
			}
		})
	}

	authorized, err := audit.NewAuthorizedRequest(principal, strings.Repeat("c", 43), nil)
	if err != nil {
		t.Fatal(err)
	}
	provider := requestAuthorizer(func(context.Context, *http.Request) (audit.AuthorizedRequest, audit.RequestFailure) {
		return authorized, audit.RequestAuthorized
	})
	service := mustServiceWithPolicy(t, db, faultyAuthorizer{err: errors.New("private policy database detail")}, audit.RetentionPolicy{MinimumAge: 30 * 24 * time.Hour, CleanupLimit: 10, Now: func() time.Time { return now }})
	handler, err := audit.NewHTTPHandler(service, provider, func(*http.Request) string { return "0123456789abcdef" })
	if err != nil {
		t.Fatal(err)
	}
	for _, request := range []*http.Request{
		httptest.NewRequest(http.MethodGet, "/audit/records?page=1&pageSize=20", nil),
		httptest.NewRequest(http.MethodGet, "/audit/records/"+testFactID("oc", "resource:demo:policy-0001:revision-1:system"), nil),
		httptest.NewRequest(http.MethodPost, "/audit/records/cleanup", strings.NewReader(`{"before":"2026-06-01T00:00:00Z","confirmation":"delete-expired-audit-records"}`)),
	} {
		if request.Method == http.MethodPost {
			request.Header.Set("Content-Type", "application/json")
		}
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusInternalServerError || strings.Contains(response.Body.String(), "private") {
			t.Fatalf("policy fault response = %d, %s", response.Code, response.Body.String())
		}
	}
	invalidIDRequest := httptest.NewRequest(http.MethodGet, "/audit/records/a1xocxcmVzb3VyY2U", nil)
	invalidIDResponse := httptest.NewRecorder()
	handler.ServeHTTP(invalidIDResponse, invalidIDRequest)
	if invalidIDResponse.Code != http.StatusBadRequest {
		t.Fatalf("invalid fact ID response = %d, %s", invalidIDResponse.Code, invalidIDResponse.Body.String())
	}
}

func TestAuthorizedRequestRejectsMalformedReplacementCookie(t *testing.T) {
	principal := audit.Principal{ID: "auditor-00000001"}
	csrf := strings.Repeat("c", 43)
	valid := sessionCookie(strings.Repeat("r", 43))
	if _, err := audit.NewAuthorizedRequest(principal, csrf, &valid); err != nil {
		t.Fatalf("valid replacement cookie = %v", err)
	}
	invalid := []string{
		"go-admin-session=" + strings.Repeat("r", 43) + "; Path=/; HttpOnly; Secure; SameSite=Strict",
		"__Host-go-admin-session=" + strings.Repeat("r", 43) + "; HttpOnly; Secure; SameSite=Strict",
		"__Host-go-admin-session=" + strings.Repeat("r", 43) + "; Path=/; Secure; SameSite=Strict",
		"__Host-go-admin-session=" + strings.Repeat("r", 43) + "; Path=/; HttpOnly; SameSite=Strict",
		"__Host-go-admin-session=" + strings.Repeat("r", 43) + "; Path=/; HttpOnly; Secure; SameSite=Lax",
		"__Host-go-admin-session=" + strings.Repeat("r", 43) + "; Path=/; HttpOnly; Secure; SameSite=Strict; Domain=example.test",
		"__Host-go-admin-session=" + strings.Repeat("r", 43) + "; Path=/; Path=/; HttpOnly; Secure; SameSite=Strict",
		"__Host-go-admin-session=" + strings.Repeat("r", 43) + "; Path=/; HttpOnly; HttpOnly; Secure; SameSite=Strict",
		"__Host-go-admin-session=" + strings.Repeat("r", 43) + "; Path=/; HttpOnly; Secure; SameSite=Strict\r\nX-Injected: yes",
	}
	for _, cookie := range invalid {
		if _, err := audit.NewAuthorizedRequest(principal, csrf, &cookie); !errors.Is(err, audit.ErrInvalidArgument) {
			t.Fatalf("invalid cookie accepted: %q, %v", cookie, err)
		}
	}
}

type requestAuthorizer func(context.Context, *http.Request) (audit.AuthorizedRequest, audit.RequestFailure)

func (provider requestAuthorizer) AuthorizeRequest(ctx context.Context, request *http.Request) (audit.AuthorizedRequest, audit.RequestFailure) {
	return provider(ctx, request)
}

func stringPointer(value string) *string { return &value }

func testFactID(alias, businessKey string) string {
	return "a1." + alias + "." + base64.RawURLEncoding.EncodeToString([]byte(businessKey))
}

func findFactID(t *testing.T, page audit.Page, subject string) string {
	t.Helper()
	for _, fact := range page.Records {
		if fact.Subject == subject {
			return fact.ID
		}
	}
	t.Fatalf("Audit subject %q not found in %#v", subject, page)
	return ""
}

type allowAll struct{}

func (allowAll) Authorize(context.Context, database.Tx, audit.Principal, audit.Permission) (audit.AuthorizationDecision, error) {
	return audit.AuthorizationGranted, nil
}

type denyAll struct{}

func (denyAll) Authorize(context.Context, database.Tx, audit.Principal, audit.Permission) (audit.AuthorizationDecision, error) {
	return audit.AuthorizationDenied, nil
}

type faultyAuthorizer struct{ err error }

func (authorizer faultyAuthorizer) Authorize(context.Context, database.Tx, audit.Principal, audit.Permission) (audit.AuthorizationDecision, error) {
	return 0, authorizer.err
}

func sessionCookie(value string) string {
	return "__Host-go-admin-session=" + value + "; Path=/; HttpOnly; Secure; SameSite=Strict"
}

type recordingObserver struct {
	mu     sync.Mutex
	events []audit.Observation
}

func (observer *recordingObserver) Observe(observation audit.Observation) {
	observer.mu.Lock()
	defer observer.mu.Unlock()
	observer.events = append(observer.events, observation)
}

func (observer *recordingObserver) snapshot() []audit.Observation {
	observer.mu.Lock()
	defer observer.mu.Unlock()
	return append([]audit.Observation(nil), observer.events...)
}

func openSQLite(t *testing.T) *database.Database {
	t.Helper()
	db, err := database.NewProcess().Open(context.Background(), database.Config{Profile: config.ProfileServerSQLite, SQLitePath: filepath.Join(t.TempDir(), "audit.sqlite3")})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func migrate(t *testing.T, db *database.Database, providers ...migrations.Provider) {
	t.Helper()
	runner, err := migrations.NewRunner(providers...)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := runner.Up(context.Background(), db); err != nil {
		t.Fatal(err)
	}
}

func newAuditStore(t *testing.T, db *database.Database) *outbox.Store {
	t.Helper()
	store, err := outbox.NewStore(db, audit.TopicSchemas()...)
	if err != nil {
		t.Fatal(err)
	}
	return store
}

func mustConsumers(t *testing.T) map[string]outbox.TransactionalConsumer {
	t.Helper()
	consumers, err := audit.TransactionalConsumers()
	if err != nil {
		t.Fatal(err)
	}
	return consumers
}

func enqueue(t *testing.T, db *database.Database, store *outbox.Store, event outbox.Event) {
	t.Helper()
	created, err := dbEnqueue(context.Background(), db, store, event)
	if err != nil || !created {
		t.Fatalf("enqueue = %v, %v", created, err)
	}
}

func dbEnqueue(ctx context.Context, db *database.Database, store *outbox.Store, event outbox.Event) (bool, error) {
	var created bool
	err := db.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error {
		var err error
		created, err = store.Enqueue(ctx, tx, event)
		return err
	})
	return created, err
}

func dispatch(t *testing.T, db *database.Database, store *outbox.Store, consumers map[string]outbox.TransactionalConsumer, now time.Time) outbox.DispatchResult {
	t.Helper()
	lease, err := coordination.Acquire(context.Background(), db, coordination.Config{Owner: "audit-test-worker"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = lease.Close(context.Background()) })
	dispatcher, err := outbox.NewDispatcher(store, outbox.DispatcherConfig{Owner: "audit-test-worker", LeaseDuration: 30 * time.Second, RetryDelay: time.Minute, BatchSize: 20, Now: func() time.Time { return now }}, consumers)
	if err != nil {
		t.Fatal(err)
	}
	result, err := dispatcher.RunOnce(context.Background(), lease, now)
	if err != nil {
		t.Fatal(err)
	}
	if err := lease.Close(context.Background()); err != nil {
		t.Fatal(err)
	}
	return result
}

func mustService(t *testing.T, db *database.Database, authorizer audit.Authorizer) *audit.Service {
	t.Helper()
	return mustServiceWithPolicy(t, db, authorizer, audit.RetentionPolicy{MinimumAge: 30 * 24 * time.Hour, CleanupLimit: 500})
}

func mustServiceWithPolicy(t *testing.T, db *database.Database, authorizer audit.Authorizer, policy audit.RetentionPolicy) *audit.Service {
	t.Helper()
	service, err := audit.NewService(db, authorizer, policy)
	if err != nil {
		t.Fatal(err)
	}
	return service
}
