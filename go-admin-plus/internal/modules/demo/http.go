package demo

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"

	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/google/uuid"

	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/contracts"
	transport "github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/demo/transport"
)

//go:embed transport/openapi.json
var openAPIDocument []byte

var (
	ErrAuthentication = errors.New("demo authentication required")
	ErrCSRF           = errors.New("demo csrf rejected")
	tracePattern      = regexp.MustCompile(`^[a-f0-9]{16,64}$`)
)

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
type HTTPServer struct{ service *Service }

func NewHTTPHandler(service *Service, authenticator RequestAuthenticator, traceID contracts.TraceIDProvider) (http.Handler, error) {
	if service == nil || authenticator == nil || authenticator.CookieName() == "" || traceID == nil {
		return nil, errors.New("demo HTTP dependencies are required")
	}
	server := &HTTPServer{service: service}
	strict := transport.NewStrictHandlerWithOptions(server, nil, transport.StrictHTTPServerOptions{
		RequestErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, _ error) {
			writeProblem(w, makeProblem(transport.Validation, "REQUEST_INVALID", "Request validation failed", traceID(r), 400))
		},
		ResponseErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, _ error) {
			writeProblem(w, makeProblem(transport.Internal, "INTERNAL_ERROR", "Internal server error", traceID(r), 500))
		},
	})
	router := transport.HandlerWithOptions(strict, transport.ChiServerOptions{ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, _ error) {
		writeProblem(w, makeProblem(transport.Validation, "REQUEST_INVALID", "Request validation failed", traceID(r), 400))
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
			switch {
			case errors.Is(authErr, ErrCSRF):
				writeProblem(w, makeProblem(transport.Authorization, "CSRF_REJECTED", "Request authorization failed", traceID(r), 403))
			case errors.Is(authErr, ErrAuthentication):
				writeProblem(w, makeProblem(transport.Authentication, "SESSION_REQUIRED", "Authentication required", traceID(r), 401))
			default:
				writeProblem(w, makeProblem(transport.Internal, "INTERNAL_ERROR", "Internal server error", traceID(r), 500))
			}
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
func headers(ctx context.Context) (string, *string) {
	value := requestValue(ctx)
	return value.csrf, value.cookie
}
func makeProblem(category transport.ProblemCategory, code, title, trace string, status int) transport.Problem {
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
func transportProduct(value Product) transport.Product {
	return transport.Product{Id: uuid.MustParse(value.ID), Sku: value.SKU, Name: value.Name, Description: value.Description, PriceCents: int(value.PriceCents), Status: transport.ProductStatus(value.Status), Revision: int(value.Revision), CreatedAt: value.CreatedAt, UpdatedAt: value.UpdatedAt}
}

func validationResponse(ctx context.Context) transport.ValidationProblemApplicationProblemPlusJSONResponse {
	csrf, cookie := headers(ctx)
	return transport.ValidationProblemApplicationProblemPlusJSONResponse{Body: makeProblem(transport.Validation, "REQUEST_INVALID", "Request validation failed", requestValue(ctx).trace, 400), Headers: transport.ValidationProblemResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}
}
func authorizationResponse(ctx context.Context) transport.AuthorizationProblemApplicationProblemPlusJSONResponse {
	csrf, cookie := headers(ctx)
	return transport.AuthorizationProblemApplicationProblemPlusJSONResponse{Body: makeProblem(transport.Authorization, "PERMISSION_DENIED", "Request authorization failed", requestValue(ctx).trace, 403), Headers: transport.AuthorizationProblemResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}
}
func notFoundResponse(ctx context.Context) transport.NotFoundProblemApplicationProblemPlusJSONResponse {
	csrf, cookie := headers(ctx)
	return transport.NotFoundProblemApplicationProblemPlusJSONResponse{Body: makeProblem(transport.NotFound, "RESOURCE_NOT_FOUND", "Resource not found", requestValue(ctx).trace, 404), Headers: transport.NotFoundProblemResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}
}
func conflictResponse(ctx context.Context) transport.ConflictProblemApplicationProblemPlusJSONResponse {
	csrf, cookie := headers(ctx)
	return transport.ConflictProblemApplicationProblemPlusJSONResponse{Body: makeProblem(transport.Conflict, "RESOURCE_CONFLICT", "Resource conflict", requestValue(ctx).trace, 409), Headers: transport.ConflictProblemResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}
}
func internalResponse(ctx context.Context) transport.InternalProblemApplicationProblemPlusJSONResponse {
	csrf, cookie := headers(ctx)
	return transport.InternalProblemApplicationProblemPlusJSONResponse{Body: makeProblem(transport.Internal, "INTERNAL_ERROR", "Internal server error", requestValue(ctx).trace, 500), Headers: transport.InternalProblemResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}
}
