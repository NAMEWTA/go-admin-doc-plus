package audit_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
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
		ID: "audit-event-0001", Topic: audit.TopicLoginSucceeded,
		BusinessKey: "login:attempt-0001", Payload: []byte(`{"source":"web","actorType":"account"}`), OccurredAt: now,
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
	if fact.ID != event.ID || fact.Kind != audit.KindLogin || fact.Action != "login" || fact.Outcome != audit.OutcomeSucceeded || fact.Source != audit.SourceWeb || fact.ActorType != audit.ActorAccount || fact.Subject != "login:attempt-0001" {
		t.Fatalf("projected fact = %#v", fact)
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
		"password":            `{"actorType":"account","source":"web","password":"private"}`,
		"raw session":         `{"actorType":"account","source":"web","rawSession":"private"}`,
		"nested request body": `{"actorType":"account","source":"web","requestBody":{"displayName":"private"}}`,
		"sensitive value":     `{"actorType":"account","source":"raw-session-secret"}`,
		"duplicate member":    `{"actorType":"system","actorType":"account","source":"web"}`,
		"array":               `{"actorType":"account","source":["web"]}`,
	}
	for name, payload := range tests {
		t.Run(name, func(t *testing.T) {
			event := outbox.Event{ID: "audit-rejected-0001", Topic: audit.TopicLoginSucceeded, BusinessKey: "login:rejected-0001", Payload: []byte(payload), OccurredAt: now}
			created, err := dbEnqueue(context.Background(), db, store, event)
			if created || !errors.Is(err, outbox.ErrInvalidEvent) {
				t.Fatalf("sensitive event = %v, %v", created, err)
			}
		})
	}
}

func TestFailedProjectionRetriesAndConvergesWithoutDuplicates(t *testing.T) {
	db := openSQLite(t)
	migrate(t, db, reliablemigration.Provider{})
	store := newAuditStore(t, db)
	consumers := mustConsumers(t)
	now := time.Date(2026, 8, 27, 2, 0, 0, 0, time.UTC)
	event := outbox.Event{ID: "audit-retry-0001", Topic: audit.TopicOperationUpdated, BusinessKey: "resource:demo:record-0001:revision-2", Payload: []byte(`{"source":"server","actorType":"system"}`), OccurredAt: now}
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
	enqueue(t, db, store, outbox.Event{ID: "audit-query-00001", Topic: audit.TopicLoginFailed, BusinessKey: "login:attempt-0002", Payload: []byte(`{"actorType":"account","source":"desktop"}`), OccurredAt: now.Add(-90 * 24 * time.Hour)})
	enqueue(t, db, store, outbox.Event{ID: "audit-query-00002", Topic: audit.TopicOperationCreated, BusinessKey: "resource:demo:record-0002:revision-1", Payload: []byte(`{"actorType":"system","source":"server"}`), OccurredAt: now.Add(-60 * 24 * time.Hour)})
	enqueue(t, db, store, outbox.Event{ID: "audit-query-00003", Topic: audit.TopicOperationUpdated, BusinessKey: "resource:demo:record-0003:revision-2", Payload: []byte(`{"actorType":"account","source":"web"}`), OccurredAt: now.Add(-24 * time.Hour)})
	dispatch(t, db, store, consumers, now)
	if _, err := db.Bun().ExecContext(context.Background(), "UPDATE audit_facts SET payload = ? WHERE event_id = ?", []byte(`{"source":"desktop","actorType":"account"}`), "audit-query-00001"); err != nil {
		t.Fatal(err)
	}

	observer := &recordingObserver{}
	service := mustServiceWithPolicy(t, db, allowAll{}, audit.RetentionPolicy{MinimumAge: 30 * 24 * time.Hour, CleanupLimit: 1, Now: func() time.Time { return now }, Observer: observer})
	principal := audit.Principal{ID: "auditor-00000001"}
	page, err := service.List(context.Background(), principal, audit.Filter{Page: 1, PageSize: 20, Kind: audit.KindLogin, Outcome: audit.OutcomeFailed, Source: audit.SourceDesktop})
	if err != nil || page.Total != 1 || page.Records[0].ID != "audit-query-00001" {
		t.Fatalf("filtered page = %#v, %v", page, err)
	}
	fact, err := service.Detail(context.Background(), principal, "audit-query-00002")
	if err != nil || fact.Action != "create" || fact.Source != audit.SourceServer {
		t.Fatalf("detail = %#v, %v", fact, err)
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
		created, err := recorder.Record(ctx, tx, audit.LoginFact{EventID: "audit-login-00001", AttemptID: "attempt-0003", Outcome: audit.OutcomeSucceeded, ActorType: audit.ActorAccount, ActorRef: stringPointer("account:account-0003"), Source: audit.SourceWeb, OccurredAt: now})
		if err != nil || !created {
			return errors.New("login record failed")
		}
		created, err = recorder.Record(ctx, tx, audit.LoginFact{EventID: "audit-login-00001", AttemptID: "attempt-0003", Outcome: audit.OutcomeSucceeded, ActorType: audit.ActorAccount, ActorRef: stringPointer("account:account-0003"), Source: audit.SourceWeb, OccurredAt: now})
		if err != nil || created {
			return errors.New("login record was not idempotent")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	page, err := mustService(t, db, allowAll{}).List(context.Background(), audit.Principal{ID: "auditor-00000001"}, audit.Filter{Page: 1, PageSize: 20})
	if err != nil || page.Total != 1 || page.Records[0].Outcome != audit.OutcomeSucceeded || page.Records[0].ActorRef == nil || *page.Records[0].ActorRef != "account:account-0003" {
		t.Fatalf("synchronous fact = %#v, %v", page, err)
	}
}

func TestHTTPTransportUsesGeneratedContractAndNeverReturnsStoredEnvelope(t *testing.T) {
	db := openSQLite(t)
	migrate(t, db, auditmigration.Provider{})
	recorder, _ := audit.NewLoginRecorder(db)
	_ = db.WithinTx(context.Background(), func(ctx context.Context, tx database.Tx) error {
		_, err := recorder.Record(ctx, tx, audit.LoginFact{EventID: "audit-http-000001", AttemptID: "attempt-http", Outcome: audit.OutcomeFailed, ActorType: audit.ActorAccount, Source: audit.SourceWeb, OccurredAt: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)})
		return err
	})
	service := mustServiceWithPolicy(t, db, allowAll{}, audit.RetentionPolicy{MinimumAge: 30 * 24 * time.Hour, CleanupLimit: 10, Now: func() time.Time { return time.Date(2026, 8, 27, 0, 0, 0, 0, time.UTC) }})
	authorized, err := audit.NewAuthorizedRequest(audit.Principal{ID: "auditor-00000001"}, strings.Repeat("c", 43), stringPointer("rotated-cookie-material"))
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
	if response.Code != http.StatusOK || response.Header().Get("Cache-Control") != "no-store" || response.Header().Get("X-CSRF-Token") != strings.Repeat("c", 43) || response.Header().Get("Set-Cookie") != "rotated-cookie-material" {
		t.Fatalf("list response = %d, %s", response.Code, response.Body.String())
	}
	body := response.Body.String()
	if strings.Contains(body, "payload") || strings.Contains(body, "businessKey") || !strings.Contains(body, `"subject":"login:attempt-http"`) {
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
		_, err := recorder.Record(ctx, tx, audit.LoginFact{EventID: "audit-authz-00001", AttemptID: "attempt-authz", Outcome: audit.OutcomeSucceeded, ActorType: audit.ActorAccount, Source: audit.SourceWeb, OccurredAt: time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)})
		return err
	})
	now := time.Date(2026, 8, 27, 0, 0, 0, 0, time.UTC)
	csrf := strings.Repeat("z", 43)
	cookie := "private-rotated-cookie"
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

type requestAuthorizer func(context.Context, *http.Request) (audit.AuthorizedRequest, audit.RequestFailure)

func (provider requestAuthorizer) AuthorizeRequest(ctx context.Context, request *http.Request) (audit.AuthorizedRequest, audit.RequestFailure) {
	return provider(ctx, request)
}

func stringPointer(value string) *string { return &value }

type allowAll struct{}

func (allowAll) Authorize(context.Context, database.Tx, audit.Principal, audit.Permission) error {
	return nil
}

type denyAll struct{}

func (denyAll) Authorize(context.Context, database.Tx, audit.Principal, audit.Permission) error {
	return errors.New("private policy material")
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
