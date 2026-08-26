package session_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go-admin/internal/modules/iam/account"
	sessionmigration "go-admin/internal/modules/iam/migrations/0010-session-schema"
	"go-admin/internal/modules/iam/session"
	"go-admin/internal/platform/config"
	"go-admin/internal/platform/database"
	"go-admin/internal/platform/migrations"
)

func TestHTTPLoginUsesHostCookieAndDoesNotReturnToken(t *testing.T) {
	_, service, _ := newFixture(t, session.Policy{IdleTimeout: time.Hour, AbsoluteTimeout: 8 * time.Hour, RotationInterval: 2 * time.Hour})
	handler, err := session.NewHTTPHandler(service, func(*http.Request) string { return "0123456789abcdef" })
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, "/iam/session/login", strings.NewReader(`{"username":"admin","password":"correct horse battery"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("login status=%d body=%s", response.Code, response.Body.String())
	}
	if response.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("authentication response is cacheable")
	}
	cookie := response.Header().Get("Set-Cookie")
	for _, required := range []string{"__Host-go-admin-session=", "Path=/", "HttpOnly", "Secure", "SameSite=Strict"} {
		if !strings.Contains(cookie, required) {
			t.Fatalf("cookie missing %q: %s", required, cookie)
		}
	}
	var body map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	encoded := response.Body.String()
	if strings.Contains(encoded, "sessionToken") || strings.Contains(encoded, "password") || strings.Contains(encoded, cookieValue(cookie)) {
		t.Fatal("sensitive session material reached JSON")
	}
	if body["csrfToken"] == "" || response.Header().Get("X-CSRF-Token") != body["csrfToken"] {
		t.Fatal("CSRF bootstrap contract is inconsistent")
	}
	change := httptest.NewRequest(http.MethodPut, "/iam/account/password", strings.NewReader(`{"currentPassword":"wrong current password","newPassword":"replacement password value"}`))
	change.Header.Set("Content-Type", "application/json")
	change.Header.Set("Cookie", strings.SplitN(cookie, ";", 2)[0])
	change.Header.Set("X-CSRF-Token", body["csrfToken"].(string))
	changeResponse := httptest.NewRecorder()
	handler.ServeHTTP(changeResponse, change)
	if changeResponse.Code != http.StatusBadRequest {
		t.Fatalf("wrong current password changed identity state: %d %s", changeResponse.Code, changeResponse.Body.String())
	}
	profile := httptest.NewRequest(http.MethodGet, "/iam/account/profile", nil)
	profile.Header.Set("Cookie", strings.SplitN(cookie, ";", 2)[0])
	profileResponse := httptest.NewRecorder()
	handler.ServeHTTP(profileResponse, profile)
	if profileResponse.Code != http.StatusOK {
		t.Fatalf("wrong current password logged out active identity: %d", profileResponse.Code)
	}

	wrongUser := loginFailure(t, handler, `{"username":"missing","password":"incorrect password"}`)
	wrongPassword := loginFailure(t, handler, `{"username":"admin","password":"incorrect password"}`)
	if wrongUser != wrongPassword {
		t.Fatalf("login failure enumerates account: %s != %s", wrongUser, wrongPassword)
	}
}

func TestHTTPInfrastructureFailureIsNotReportedAsLogout(t *testing.T) {
	service, err := session.NewService(failingDatabase{}, session.Policy{IdleTimeout: time.Hour, AbsoluteTimeout: 2 * time.Hour, RotationInterval: time.Hour})
	if err != nil {
		t.Fatal(err)
	}
	handler, err := session.NewHTTPHandler(service, func(*http.Request) string { return "0123456789abcdef" })
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, "/iam/session/login", strings.NewReader(`{"username":"admin","password":"correct horse battery"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusInternalServerError || !strings.Contains(response.Body.String(), `"category":"internal"`) {
		t.Fatalf("dependency failure was misclassified: %d %s", response.Code, response.Body.String())
	}
	if strings.Contains(response.Body.String(), "database") || strings.Contains(response.Body.String(), "secret") || strings.Contains(response.Body.String(), "path") {
		t.Fatal("internal detail escaped the stable problem")
	}
}

type failingDatabase struct{}

func (failingDatabase) WithinTx(context.Context, func(context.Context, database.Tx) error) error {
	return errors.New("database secret/path must not escape")
}
func (failingDatabase) Dialect() database.Dialect { return database.DialectSQLite }

func loginFailure(t *testing.T, handler http.Handler, body string) string {
	t.Helper()
	request := httptest.NewRequest(http.MethodPost, "/iam/session/login", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("failure status=%d body=%s", response.Code, response.Body.String())
	}
	return response.Body.String()
}

func cookieValue(header string) string {
	value := strings.TrimPrefix(strings.SplitN(header, ";", 2)[0], session.CookieName+"=")
	return value
}

func TestArgon2idPolicyRejectsLegacyAndWrongCredentials(t *testing.T) {
	hash, err := account.HashPassword("correct horse battery")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(hash, "$argon2id$v=19$m=65536,t=3,p=4$") {
		t.Fatalf("unexpected password format: %q", hash)
	}
	if !account.VerifyPassword(hash, "correct horse battery") {
		t.Fatal("correct password rejected")
	}
	if account.VerifyPassword(hash, "wrong password") || account.VerifyPassword("$2a$legacy", "correct horse battery") {
		t.Fatal("invalid password representation accepted")
	}
}

func TestSessionLifecyclePersistsOnlyDigests(t *testing.T) {
	db, service, clock := newFixture(t, session.Policy{IdleTimeout: 2 * time.Hour, AbsoluteTimeout: 8 * time.Hour, RotationInterval: time.Hour})
	issued, err := service.Login(context.Background(), "ADMIN", "correct horse battery")
	if err != nil {
		t.Fatal(err)
	}
	if len(issued.Token) != 43 || len(issued.CSRF) != 43 {
		t.Fatal("session material is not high entropy")
	}
	var tokenHash, csrfHash string
	if err := db.Bun().QueryRowContext(context.Background(), `SELECT token_hash, csrf_hash FROM iam_sessions`).Scan(&tokenHash, &csrfHash); err != nil {
		t.Fatal(err)
	}
	if tokenHash == issued.Token || csrfHash == issued.CSRF || len(tokenHash) != 64 || len(csrfHash) != 64 {
		t.Fatal("raw session material reached persistence")
	}
	serialized, err := json.Marshal(issued)
	if err != nil || strings.Contains(string(serialized), issued.Token) || strings.Contains(string(serialized), issued.CSRF) {
		t.Fatal("issued session serialization leaked secrets")
	}

	*clock = clock.Add(61 * time.Minute)
	rotated, err := service.Current(context.Background(), issued.Token)
	if err != nil || !rotated.Rotated || rotated.Token == issued.Token {
		t.Fatalf("rotation failed: %#v %v", rotated, err)
	}
	if _, err := service.Current(context.Background(), issued.Token); !errors.Is(err, session.ErrAuthentication) {
		t.Fatalf("rotated token recovered: %v", err)
	}
	refreshed, err := service.Current(context.Background(), rotated.Token)
	if err != nil {
		t.Fatalf("replacement token rejected: %v", err)
	}

	if err := service.Logout(context.Background(), rotated.Token, refreshed.CSRF); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Current(context.Background(), rotated.Token); !errors.Is(err, session.ErrAuthentication) {
		t.Fatalf("revoked token recovered: %v", err)
	}
}

func TestCSRFPasswordAndTimeoutFailuresArePermanentAndAtomic(t *testing.T) {
	db, service, clock := newFixture(t, session.Policy{IdleTimeout: 10 * time.Minute, AbsoluteTimeout: 30 * time.Minute, RotationInterval: 20 * time.Minute})
	issued, err := service.Login(context.Background(), "admin", "correct horse battery")
	if err != nil {
		t.Fatal(err)
	}
	before, err := service.Profile(context.Background(), issued.Token)
	if err != nil {
		t.Fatal(err)
	}
	var idleBefore time.Time
	if err := db.Bun().QueryRowContext(context.Background(), `SELECT idle_expires_at FROM iam_sessions WHERE state = 'active'`).Scan(&idleBefore); err != nil {
		t.Fatal(err)
	}
	*clock = clock.Add(time.Minute)
	if _, err := service.UpdateProfile(context.Background(), issued.Token, "wrong-csrf", "Changed", "changed@example.test", nil); !errors.Is(err, session.ErrCSRF) {
		t.Fatalf("expected CSRF rejection: %v", err)
	}
	var idleAfter time.Time
	if err := db.Bun().QueryRowContext(context.Background(), `SELECT idle_expires_at FROM iam_sessions WHERE state = 'active'`).Scan(&idleAfter); err != nil {
		t.Fatal(err)
	}
	if !idleAfter.Equal(idleBefore) {
		t.Fatal("CSRF failure refreshed idle timeout")
	}
	after, _ := service.Profile(context.Background(), issued.Token)
	if before != after {
		t.Fatal("CSRF failure mutated profile")
	}

	if err := service.ChangePassword(context.Background(), issued.Token, issued.CSRF, "correct horse battery", "replacement password value"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Current(context.Background(), issued.Token); !errors.Is(err, session.ErrAuthentication) {
		t.Fatal("password change did not revoke current session")
	}
	if _, err := service.Login(context.Background(), "admin", "correct horse battery"); !errors.Is(err, session.ErrCredentials) {
		t.Fatal("old password remains valid")
	}
	second, err := service.Login(context.Background(), "admin", "replacement password value")
	if err != nil {
		t.Fatal(err)
	}

	*clock = clock.Add(11 * time.Minute)
	if _, err := service.Current(context.Background(), second.Token); !errors.Is(err, session.ErrAuthentication) {
		t.Fatalf("idle expiry accepted: %v", err)
	}
	if _, err := service.Current(context.Background(), second.Token); !errors.Is(err, session.ErrAuthentication) {
		t.Fatalf("idle-expired token recovered: %v", err)
	}
}

func TestProtectedReadsTouchIdleButAbsoluteExpiryWins(t *testing.T) {
	db, service, clock := newFixture(t, session.Policy{IdleTimeout: 10 * time.Minute, AbsoluteTimeout: 25 * time.Minute, RotationInterval: 20 * time.Minute})
	issued, err := service.Login(context.Background(), "admin", "correct horse battery")
	if err != nil {
		t.Fatal(err)
	}
	var originalIdle time.Time
	if err := db.Bun().QueryRowContext(context.Background(), `SELECT idle_expires_at FROM iam_sessions WHERE state = 'active'`).Scan(&originalIdle); err != nil {
		t.Fatal(err)
	}
	*clock = clock.Add(8 * time.Minute)
	if _, err := service.Profile(context.Background(), issued.Token); err != nil {
		t.Fatal(err)
	}
	var touchedIdle time.Time
	if err := db.Bun().QueryRowContext(context.Background(), `SELECT idle_expires_at FROM iam_sessions WHERE state = 'active'`).Scan(&touchedIdle); err != nil {
		t.Fatal(err)
	}
	if !touchedIdle.After(originalIdle) {
		t.Fatal("successful protected read did not refresh idle timeout")
	}
	*clock = clock.Add(18 * time.Minute)
	if _, err := service.Profile(context.Background(), issued.Token); !errors.Is(err, session.ErrAuthentication) {
		t.Fatalf("absolute expiry accepted: %v", err)
	}
	var state string
	if err := db.Bun().QueryRowContext(context.Background(), `SELECT state FROM iam_sessions`).Scan(&state); err != nil {
		t.Fatal(err)
	}
	if state != "expired" {
		t.Fatalf("expired transition was not committed: %q", state)
	}
}

func TestProfilePersistenceAndAdministrativeRevoke(t *testing.T) {
	_, service, _ := newFixture(t, session.Policy{IdleTimeout: time.Hour, AbsoluteTimeout: 8 * time.Hour, RotationInterval: 2 * time.Hour})
	issued, err := service.Login(context.Background(), "admin", "correct horse battery")
	if err != nil {
		t.Fatal(err)
	}
	avatar := "files/avatar-1"
	profile, err := service.UpdateProfile(context.Background(), issued.Token, issued.CSRF, "New Name", "new@example.test", &avatar)
	if err != nil {
		t.Fatal(err)
	}
	persisted, err := service.Profile(context.Background(), issued.Token)
	if err != nil || persisted.DisplayName != profile.DisplayName || persisted.Email != "new@example.test" {
		t.Fatalf("profile was not persisted: %#v %v", persisted, err)
	}
	if err := service.RevokeAccount(context.Background(), profile.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Current(context.Background(), issued.Token); !errors.Is(err, session.ErrAuthentication) {
		t.Fatal("administrative revoke did not fence token")
	}
}

func TestDisabledAccountCannotAuthenticateOrMutate(t *testing.T) {
	db, service, clock := newFixture(t, session.Policy{IdleTimeout: time.Hour, AbsoluteTimeout: 8 * time.Hour, RotationInterval: 2 * time.Hour})
	issued, err := service.Login(context.Background(), "admin", "correct horse battery")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Bun().ExecContext(context.Background(), `UPDATE iam_accounts SET disabled_at = ? WHERE id = ?`, clock.UTC(), "account-00000001"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Login(context.Background(), "admin", "correct horse battery"); !errors.Is(err, session.ErrCredentials) {
		t.Fatalf("disabled account login was distinguishable or accepted: %v", err)
	}
	if _, err := service.UpdateProfile(context.Background(), issued.Token, issued.CSRF, "Changed", "changed@example.test", nil); !errors.Is(err, session.ErrAuthentication) {
		t.Fatalf("disabled account mutated profile: %v", err)
	}
}

func newFixture(t *testing.T, policy session.Policy) (*database.Database, *session.Service, *time.Time) {
	t.Helper()
	ctx := context.Background()
	process := database.NewProcess()
	db, err := process.Open(ctx, database.Config{Profile: config.ProfileServerSQLite, SQLitePath: t.TempDir() + "/iam.db"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	runner, err := migrations.NewRunner(sessionmigration.Provider{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := runner.Up(ctx, db); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 8, 26, 8, 0, 0, 0, time.UTC)
	hash, err := account.HashPassword("correct horse battery")
	if err != nil {
		t.Fatal(err)
	}
	repository := account.NewRepository(db.Dialect())
	err = db.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error {
		return repository.Create(ctx, tx, account.Credential{Profile: account.Profile{ID: "account-00000001", Username: "admin", DisplayName: "Administrator", Email: "admin@example.test"}, PasswordHash: hash}, now)
	})
	if err != nil {
		t.Fatal(err)
	}
	// The pointer controls the closure without exporting clock mutation into production APIs.
	clock := &now
	service, err := session.NewService(db, policy, session.WithClock(func() time.Time { return *clock }))
	if err != nil {
		t.Fatal(err)
	}
	return db, service, clock
}
