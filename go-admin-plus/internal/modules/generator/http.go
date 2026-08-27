package generator

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"

	"github.com/getkin/kin-openapi/openapi3filter"

	"go-admin/internal/contracts"
	transport "go-admin/internal/modules/generator/transport"
)

//go:embed transport/openapi.json
var openAPIDocument []byte

var tracePattern = regexp.MustCompile(`^[a-f0-9]{16,64}$`)

type RequestIdentity struct {
	ActorID, CSRF     string
	ReplacementCookie *string
}

type RequestAuthenticator interface {
	CookieName() string
	AuthorizeRequest(context.Context, string, string, bool) (RequestIdentity, error)
}

type requestContextKey struct{}
type requestContext struct {
	actorID, csrf, trace string
	cookie               *string
}
type HTTPServer struct{ generator *Generator }

func NewHTTPHandler(generator *Generator, authenticator RequestAuthenticator, traceID contracts.TraceIDProvider) (http.Handler, error) {
	if generator == nil || authenticator == nil || authenticator.CookieName() == "" || traceID == nil {
		return nil, ErrInvalid
	}
	server := &HTTPServer{generator: generator}
	strict := transport.NewStrictHandlerWithOptions(server, nil, transport.StrictHTTPServerOptions{
		RequestErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, _ error) {
			writeProblem(w, problem(transport.Validation, "REQUEST_INVALID", "Request validation failed", traceID(r), 400))
		},
		ResponseErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, _ error) {
			writeProblem(w, problem(transport.Internal, "INTERNAL_ERROR", "Internal server error", traceID(r), 500))
		},
	})
	router := transport.HandlerWithOptions(strict, transport.ChiServerOptions{ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, _ error) {
		writeProblem(w, problem(transport.Validation, "REQUEST_INVALID", "Request validation failed", traceID(r), 400))
	}})
	validated, err := contracts.NewRequestValidator(openAPIDocument, router, traceID, contracts.RequestValidatorOptions{MaxBodyBytes: contracts.DefaultMaxRequestBodyBytes, AuthenticationFunc: openapi3filter.NoopAuthenticationFunc})
	if err != nil {
		return nil, err
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		cookie, _ := r.Cookie(authenticator.CookieName())
		token := ""
		if cookie != nil {
			token = cookie.Value
		}
		identity, authErr := authenticator.AuthorizeRequest(r.Context(), token, r.Header.Get("X-CSRF-Token"), r.Method != http.MethodGet && r.Method != http.MethodHead)
		if authErr != nil {
			status, category, code, title := 401, transport.Authentication, "SESSION_REQUIRED", "Authentication required"
			if errors.Is(authErr, ErrCSRF) {
				status, category, code, title = 403, transport.Authorization, "CSRF_REJECTED", "Request authorization failed"
			}
			writeProblem(w, problem(category, code, title, traceID(r), status))
			return
		}
		w.Header().Set("X-CSRF-Token", identity.CSRF)
		if identity.ReplacementCookie != nil {
			w.Header().Set("Set-Cookie", *identity.ReplacementCookie)
		}
		value := requestContext{actorID: identity.ActorID, csrf: identity.CSRF, cookie: identity.ReplacementCookie, trace: traceID(r)}
		validated.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), requestContextKey{}, value)))
	}), nil
}

func requestValue(ctx context.Context) requestContext {
	value, _ := ctx.Value(requestContextKey{}).(requestContext)
	return value
}
func responseHeaders(ctx context.Context) (string, *string) {
	value := requestValue(ctx)
	return value.csrf, value.cookie
}
func problem(category transport.ProblemCategory, code, title, trace string, status int) transport.Problem {
	if !tracePattern.MatchString(trace) {
		trace = "0000000000000000"
	}
	return transport.Problem{Type: "urn:go-admin-plus:problem:" + strings.ToLower(strings.ReplaceAll(code, "_", "-")), Title: title, Status: status, Category: category, Code: code, TraceId: trace}
}
func writeProblem(w http.ResponseWriter, value transport.Problem) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(value.Status)
	_ = json.NewEncoder(w).Encode(value)
}

func authorizationProblem(ctx context.Context) transport.AuthorizationProblemApplicationProblemPlusJSONResponse {
	csrf, cookie := responseHeaders(ctx)
	return transport.AuthorizationProblemApplicationProblemPlusJSONResponse{Body: problem(transport.Authorization, "PERMISSION_DENIED", "Request authorization failed", requestValue(ctx).trace, 403), Headers: transport.AuthorizationProblemResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}
}
func validationProblem(ctx context.Context) transport.ValidationProblemApplicationProblemPlusJSONResponse {
	csrf, cookie := responseHeaders(ctx)
	return transport.ValidationProblemApplicationProblemPlusJSONResponse{Body: problem(transport.Validation, "REQUEST_INVALID", "Request validation failed", requestValue(ctx).trace, 400), Headers: transport.ValidationProblemResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}
}
func notFoundProblem(ctx context.Context) transport.NotFoundProblemApplicationProblemPlusJSONResponse {
	csrf, cookie := responseHeaders(ctx)
	return transport.NotFoundProblemApplicationProblemPlusJSONResponse{Body: problem(transport.NotFound, "RESOURCE_NOT_FOUND", "Resource not found", requestValue(ctx).trace, 404), Headers: transport.NotFoundProblemResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}
}
func conflictProblem(ctx context.Context) transport.ConflictProblemApplicationProblemPlusJSONResponse {
	csrf, cookie := responseHeaders(ctx)
	return transport.ConflictProblemApplicationProblemPlusJSONResponse{Body: problem(transport.Conflict, "OUTPUT_CONFLICT", "Output already exists", requestValue(ctx).trace, 409), Headers: transport.ConflictProblemResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}
}
func gateProblem(ctx context.Context) transport.GateProblemApplicationProblemPlusJSONResponse {
	csrf, cookie := responseHeaders(ctx)
	return transport.GateProblemApplicationProblemPlusJSONResponse{Body: problem(transport.Validation, "OUTPUT_GATE_FAILED", "Generated output failed validation", requestValue(ctx).trace, 422), Headers: transport.GateProblemResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}
}
func internalProblem(ctx context.Context) transport.InternalProblemApplicationProblemPlusJSONResponse {
	csrf, cookie := responseHeaders(ctx)
	return transport.InternalProblemApplicationProblemPlusJSONResponse{Body: problem(transport.Internal, "INTERNAL_ERROR", "Internal server error", requestValue(ctx).trace, 500), Headers: transport.InternalProblemResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}
}
