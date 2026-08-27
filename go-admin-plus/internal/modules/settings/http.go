package settings

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
	"go-admin/internal/contracts"
	transport "go-admin/internal/modules/settings/transport"
)

//go:embed transport/openapi.json
var openAPIDocument []byte

var (
	ErrAuthentication = errors.New("settings authentication required")
	ErrCSRF           = errors.New("settings csrf rejected")
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

func NewHTTPHandler(service *Service, auth RequestAuthenticator, traceID contracts.TraceIDProvider) (http.Handler, error) {
	if service == nil || auth == nil || auth.CookieName() == "" || traceID == nil {
		return nil, errors.New("settings HTTP dependencies are required")
	}
	server := &HTTPServer{service: service}
	strict := transport.NewStrictHandlerWithOptions(server, nil, transport.StrictHTTPServerOptions{RequestErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, _ error) {
		writeProblem(w, problem(transport.Validation, "REQUEST_INVALID", "Request validation failed", traceID(r), 400))
	}, ResponseErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, _ error) {
		writeProblem(w, problem(transport.Internal, "INTERNAL_ERROR", "Internal server error", traceID(r), 500))
	}})
	router := transport.HandlerWithOptions(strict, transport.ChiServerOptions{ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, _ error) {
		writeProblem(w, problem(transport.Validation, "REQUEST_INVALID", "Request validation failed", traceID(r), 400))
	}})
	validated, err := contracts.NewRequestValidator(openAPIDocument, router, traceID, contracts.RequestValidatorOptions{MaxBodyBytes: contracts.DefaultMaxRequestBodyBytes, AuthenticationFunc: openapi3filter.NoopAuthenticationFunc})
	if err != nil {
		return nil, err
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		cookie, _ := r.Cookie(auth.CookieName())
		token := ""
		if cookie != nil {
			token = cookie.Value
		}
		identity, authErr := auth.AuthorizeRequest(r.Context(), token, r.Header.Get("X-CSRF-Token"), r.Method != http.MethodGet && r.Method != http.MethodHead)
		if authErr != nil {
			if errors.Is(authErr, ErrCSRF) {
				writeProblem(w, problem(transport.Authorization, "CSRF_REJECTED", "Request authorization failed", traceID(r), 403))
			} else if errors.Is(authErr, ErrAuthentication) {
				writeProblem(w, problem(transport.Authentication, "SESSION_REQUIRED", "Authentication required", traceID(r), 401))
			} else {
				writeProblem(w, problem(transport.Internal, "INTERNAL_ERROR", "Internal server error", traceID(r), 500))
			}
			return
		}
		w.Header().Set("X-CSRF-Token", identity.CSRF)
		if identity.ReplacementCookie != nil {
			w.Header().Set("Set-Cookie", *identity.ReplacementCookie)
		}
		ctx := context.WithValue(r.Context(), requestContextKey{}, requestContext{actorID: identity.ActorID, csrf: identity.CSRF, cookie: identity.ReplacementCookie, trace: traceID(r)})
		validated.ServeHTTP(w, r.WithContext(ctx))
	}), nil
}
func requestValue(ctx context.Context) requestContext {
	v, _ := ctx.Value(requestContextKey{}).(requestContext)
	return v
}
func headers(ctx context.Context) (string, *string) { v := requestValue(ctx); return v.csrf, v.cookie }
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
func validation(ctx context.Context) transport.ValidationProblemApplicationProblemPlusJSONResponse {
	csrf, cookie := headers(ctx)
	return transport.ValidationProblemApplicationProblemPlusJSONResponse{Body: problem(transport.Validation, "REQUEST_INVALID", "Request validation failed", requestValue(ctx).trace, 400), Headers: transport.ValidationProblemResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}
}
func authorizationProblem(ctx context.Context) transport.AuthorizationProblemApplicationProblemPlusJSONResponse {
	csrf, cookie := headers(ctx)
	return transport.AuthorizationProblemApplicationProblemPlusJSONResponse{Body: problem(transport.Authorization, "PERMISSION_DENIED", "Request authorization failed", requestValue(ctx).trace, 403), Headers: transport.AuthorizationProblemResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}
}
func notFound(ctx context.Context) transport.NotFoundProblemApplicationProblemPlusJSONResponse {
	csrf, cookie := headers(ctx)
	return transport.NotFoundProblemApplicationProblemPlusJSONResponse{Body: problem(transport.NotFound, "RESOURCE_NOT_FOUND", "Resource not found", requestValue(ctx).trace, 404), Headers: transport.NotFoundProblemResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}
}
func conflict(ctx context.Context) transport.ConflictProblemApplicationProblemPlusJSONResponse {
	csrf, cookie := headers(ctx)
	return transport.ConflictProblemApplicationProblemPlusJSONResponse{Body: problem(transport.Conflict, "RESOURCE_CONFLICT", "Resource conflict", requestValue(ctx).trace, 409), Headers: transport.ConflictProblemResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}
}
func internal(ctx context.Context) transport.InternalProblemApplicationProblemPlusJSONResponse {
	csrf, cookie := headers(ctx)
	return transport.InternalProblemApplicationProblemPlusJSONResponse{Body: problem(transport.Internal, "INTERNAL_ERROR", "Internal server error", requestValue(ctx).trace, 500), Headers: transport.InternalProblemResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}
}
func query(search *string, page, pageSize *int) ListQuery {
	result := ListQuery{Page: 1, PerPage: 20}
	if search != nil {
		result.Search = *search
	}
	if page != nil {
		result.Page = *page
	}
	if pageSize != nil {
		result.PerPage = *pageSize
	}
	return result
}
func settingTransport(v Setting) transport.SettingValue {
	return transport.SettingValue{Id: uuid.MustParse(v.ID), Category: transport.SettingCategory(v.Category), Key: v.Key, Label: v.Label, Value: v.Value, Description: v.Description, Enabled: v.Enabled, Revision: int(v.Revision)}
}
func dictionaryTransport(v Dictionary) transport.DictionaryType {
	return transport.DictionaryType{Id: uuid.MustParse(v.ID), Key: v.Key, Name: v.Name, Description: v.Description, Enabled: v.Enabled, Revision: int(v.Revision)}
}
func itemTransport(v DictionaryItem) transport.DictionaryItem {
	return transport.DictionaryItem{Id: uuid.MustParse(v.ID), DictionaryId: uuid.MustParse(v.DictionaryID), Value: v.Value, Label: v.Label, SortOrder: v.SortOrder, Enabled: v.Enabled, Revision: int(v.Revision)}
}
