package contracts

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
)

const fallbackTraceID = "0000000000000000"

var validTraceID = regexp.MustCompile(`^[a-f0-9]{16,64}$`)

// TraceIDProvider returns the safe correlation identifier exposed in a Problem response.
type TraceIDProvider func(*http.Request) string

type stableProblemDefinition struct {
	status int
	typeID string
	title  string
	code   string
}

var stableProblems = map[ProblemCategory]stableProblemDefinition{
	Validation:     {http.StatusBadRequest, "urn:go-admin-plus:problem:validation", "Request validation failed", "REQUEST_INVALID"},
	Authentication: {http.StatusUnauthorized, "urn:go-admin-plus:problem:authentication", "Authentication required", "SESSION_REQUIRED"},
	Authorization:  {http.StatusForbidden, "urn:go-admin-plus:problem:authorization", "Permission denied", "PERMISSION_DENIED"},
	NotFound:       {http.StatusNotFound, "urn:go-admin-plus:problem:not-found", "Resource not found", "RESOURCE_NOT_FOUND"},
	Conflict:       {http.StatusConflict, "urn:go-admin-plus:problem:conflict", "Resource conflict", "RESOURCE_CONFLICT"},
	Internal:       {http.StatusInternalServerError, "urn:go-admin-plus:problem:internal", "Internal server error", "INTERNAL_ERROR"},
}

// NewStableProblem constructs a declared public failure without accepting raw internal detail.
func NewStableProblem(category ProblemCategory, traceID string) (Problem, error) {
	definition, ok := stableProblems[category]
	if !ok {
		return Problem{}, errors.New("unsupported public problem category")
	}
	return Problem{
		Type: definition.typeID, Title: definition.title, Status: definition.status,
		Category: category, Code: definition.code, TraceId: safeTraceID(traceID),
	}, nil
}

// NewHandler builds the generated Chi and strict layers with the stable Problem contract.
func NewHandler(
	implementation StrictServerInterface,
	traceID TraceIDProvider,
	middlewares []StrictMiddlewareFunc,
) (http.Handler, error) {
	if implementation == nil {
		return nil, errors.New("strict server implementation is required")
	}
	if traceID == nil {
		return nil, errors.New("trace ID provider is required")
	}

	strict := NewStrictHandlerWithOptions(implementation, middlewares, StrictHTTPServerOptions{
		RequestErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, _ error) {
			writeValidationProblem(w, safeTraceID(traceID(r)))
		},
		ResponseErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, _ error) {
			problem, _ := NewStableProblem(Internal, traceID(r))
			writeProblem(w, problem)
		},
	})

	return HandlerWithOptions(strict, ChiServerOptions{
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, _ error) {
			writeValidationProblem(w, safeTraceID(traceID(r)))
		},
	}), nil
}

func safeTraceID(traceID string) string {
	if validTraceID.MatchString(traceID) {
		return traceID
	}
	return fallbackTraceID
}

func writeValidationProblem(w http.ResponseWriter, traceID string) {
	problem, _ := NewStableProblem(Validation, traceID)
	detail := "Request does not match the API contract."
	problem.Detail = &detail
	writeProblem(w, problem)
}

func writeProblem(w http.ResponseWriter, problem Problem) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(problem.Status)
	_ = json.NewEncoder(w).Encode(problem)
}
