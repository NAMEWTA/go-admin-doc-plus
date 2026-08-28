package organization

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const organizationTestCookie = "organization-test-session"

type sessionAuthorizerStub struct {
	identity RequestIdentity
	err      error
	mutation bool
	csrf     string
	token    string
}

func (*sessionAuthorizerStub) CookieName() string { return organizationTestCookie }

func (stub *sessionAuthorizerStub) AuthorizeRequest(_ context.Context, token, csrf string, mutation bool) (RequestIdentity, error) {
	stub.token, stub.csrf, stub.mutation = token, csrf, mutation
	return stub.identity, stub.err
}

func TestOrganizationHTTPStrictBoundaryAndStableProblems(t *testing.T) {
	db := organizationDatabase(t)
	authorizer := &authorizerStub{scope: ScopeAll}
	service, err := NewService(db, authorizer)
	if err != nil {
		t.Fatal(err)
	}
	replacement := "organization-test-session=rotated-token; Path=/; HttpOnly; Secure; SameSite=Strict"
	sessions := &sessionAuthorizerStub{identity: RequestIdentity{ActorID: "account-admin-001", CSRF: "csrf-next", ReplacementCookie: &replacement}}
	handler, err := NewHTTPHandler(service, sessions, func(*http.Request) string { return "0123456789abcdef" })
	if err != nil {
		t.Fatal(err)
	}

	response := organizationRequest(handler, http.MethodGet, "/organization/departments", "", "session-token", "")
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"key":"root"`) {
		t.Fatalf("department list status=%d body=%s", response.Code, response.Body.String())
	}
	if sessions.mutation || sessions.token != "session-token" || response.Header().Get("Cache-Control") != "no-store" || response.Header().Get("X-CSRF-Token") != "csrf-next" {
		t.Fatalf("read boundary mutation=%t token=%q headers=%v", sessions.mutation, sessions.token, response.Header())
	}
	setCookie := response.Header().Get("Set-Cookie")
	if !strings.Contains(setCookie, organizationTestCookie+"=rotated-token") || !strings.Contains(setCookie, "HttpOnly") || !strings.Contains(setCookie, "Secure") || !strings.Contains(setCookie, "SameSite=Strict") {
		t.Fatalf("rotation cookie attributes missing: %q", setCookie)
	}

	body := `{"key":"operations","name":"Operations","parentId":"department-root-001","sortOrder":10}`
	response = organizationRequest(handler, http.MethodPost, "/organization/departments", body, "session-token", "csrf-current")
	if response.Code != http.StatusCreated || !sessions.mutation || sessions.csrf != "csrf-current" {
		t.Fatalf("create status=%d body=%s mutation=%t csrf=%q", response.Code, response.Body.String(), sessions.mutation, sessions.csrf)
	}
	response = organizationRequest(handler, http.MethodPost, "/organization/departments", body, "session-token", "csrf-current")
	assertStableProblem(t, response, http.StatusConflict, "conflict")
	response = organizationRequest(handler, http.MethodPost, "/organization/departments", `{"key":"invalid"}`, "session-token", "csrf-current")
	assertStableProblem(t, response, http.StatusBadRequest, "validation")

	authorizer.err = errors.New("private SQL /var/db/organization password=secret")
	response = organizationRequest(handler, http.MethodGet, "/organization/departments", "", "session-token", "")
	assertStableProblem(t, response, http.StatusInternalServerError, "internal")
	for _, secret := range []string{"private SQL", "/var/db", "password", "secret", "rotated-token", "csrf-next"} {
		if strings.Contains(response.Body.String(), secret) {
			t.Fatalf("response leaked %q: %s", secret, response.Body.String())
		}
	}
}

func TestOrganizationHTTPAuthenticationFailsBeforeService(t *testing.T) {
	db := organizationDatabase(t)
	authorizer := &authorizerStub{scope: ScopeAll}
	service, err := NewService(db, authorizer)
	if err != nil {
		t.Fatal(err)
	}
	sessions := &sessionAuthorizerStub{err: ErrAuthentication}
	handler, err := NewHTTPHandler(service, sessions, func(*http.Request) string { return "0123456789abcdef" })
	if err != nil {
		t.Fatal(err)
	}
	response := organizationRequest(handler, http.MethodGet, "/organization/departments", "", "", "")
	assertStableProblem(t, response, http.StatusUnauthorized, "authentication")
	if len(authorizer.permissions) != 0 {
		t.Fatalf("authentication failure reached authorizer: %v", authorizer.permissions)
	}
}

func organizationRequest(handler http.Handler, method, path, body, token, csrf string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		request.AddCookie(&http.Cookie{Name: organizationTestCookie, Value: token})
	}
	if csrf != "" {
		request.Header.Set("X-CSRF-Token", csrf)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func assertStableProblem(t *testing.T, response *httptest.ResponseRecorder, status int, category string) {
	t.Helper()
	if response.Code != status || !strings.Contains(response.Header().Get("Content-Type"), "application/problem+json") || !strings.Contains(response.Body.String(), `"category":"`+category+`"`) || !strings.Contains(response.Body.String(), `"traceId":"0123456789abcdef"`) {
		t.Fatalf("problem status=%d body=%s headers=%v", response.Code, response.Body.String(), response.Header())
	}
}
