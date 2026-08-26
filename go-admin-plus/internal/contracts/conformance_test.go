package contracts

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type probeHandler struct{}

var _ StrictServerInterface = probeHandler{}

func (probeHandler) ProbeContract(_ context.Context, request ProbeContractRequestObject) (ProbeContractResponseObject, error) {
	if request.Body.Value == "fail" {
		return nil, errors.New("pq: secret=database-password at /var/lib/app/database.go:42")
	}
	return ProbeContract200JSONResponse{Value: request.Body.Value}, nil
}

func fixedTraceID(*http.Request) string { return "0123456789abcdef" }

func TestStableProblemCategories(t *testing.T) {
	tests := []struct {
		category ProblemCategory
		status   int
		code     string
		typeID   string
	}{
		{Validation, http.StatusBadRequest, "REQUEST_INVALID", "urn:go-admin-plus:problem:validation"},
		{Authentication, http.StatusUnauthorized, "SESSION_REQUIRED", "urn:go-admin-plus:problem:authentication"},
		{Authorization, http.StatusForbidden, "PERMISSION_DENIED", "urn:go-admin-plus:problem:authorization"},
		{NotFound, http.StatusNotFound, "RESOURCE_NOT_FOUND", "urn:go-admin-plus:problem:not-found"},
		{Conflict, http.StatusConflict, "RESOURCE_CONFLICT", "urn:go-admin-plus:problem:conflict"},
		{Internal, http.StatusInternalServerError, "INTERNAL_ERROR", "urn:go-admin-plus:problem:internal"},
	}

	for _, test := range tests {
		t.Run(string(test.category), func(t *testing.T) {
			problem, err := NewStableProblem(test.category, "unsafe trace value")
			if err != nil {
				t.Fatalf("construct stable problem: %v", err)
			}
			if problem.Status != test.status || problem.Code != test.code || problem.Type != test.typeID {
				t.Fatalf("problem = %#v, want status=%d code=%s type=%s", problem, test.status, test.code, test.typeID)
			}
			if problem.Category != test.category || problem.TraceId != fallbackTraceID {
				t.Fatalf("problem category/trace = %q/%q", problem.Category, problem.TraceId)
			}
		})
	}

	if _, err := NewStableProblem(ProblemCategory("database"), fixedTraceID(nil)); err == nil {
		t.Fatal("unsupported public category unexpectedly accepted")
	}
}

func TestGeneratedStrictHandlerUsesChiTransport(t *testing.T) {
	handler, err := NewHandler(probeHandler{}, fixedTraceID, nil)
	if err != nil {
		t.Fatalf("create handler: %v", err)
	}
	request := httptest.NewRequest(http.MethodPost, "/contract/probe?page=2&pageSize=40", strings.NewReader(`{"value":"roundtrip"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusOK, response.Body.String())
	}
	if contentType := response.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("content type = %q, want application/json", contentType)
	}
	var body ContractProbeResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Value != "roundtrip" {
		t.Fatalf("value = %q, want roundtrip", body.Value)
	}
}

func TestStrictTransportNormalizesRequestAndInternalFailures(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		body     string
		status   int
		category ProblemCategory
	}{
		{name: "invalid request", url: "/contract/probe", body: `{`, status: http.StatusBadRequest, category: Validation},
		{name: "invalid query", url: "/contract/probe?page=invalid", body: `{"value":"ok"}`, status: http.StatusBadRequest, category: Validation},
		{name: "internal failure", url: "/contract/probe", body: `{"value":"fail"}`, status: http.StatusInternalServerError, category: Internal},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler, err := NewHandler(probeHandler{}, fixedTraceID, nil)
			if err != nil {
				t.Fatalf("create handler: %v", err)
			}
			request := httptest.NewRequest(http.MethodPost, test.url, strings.NewReader(test.body))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			if response.Code != test.status {
				t.Fatalf("status = %d, want %d; body=%s", response.Code, test.status, response.Body.String())
			}
			if contentType := response.Header().Get("Content-Type"); contentType != "application/problem+json" {
				t.Fatalf("content type = %q, want application/problem+json", contentType)
			}
			var problem Problem
			if err := json.NewDecoder(response.Body).Decode(&problem); err != nil {
				t.Fatalf("decode problem: %v", err)
			}
			if problem.Category != test.category {
				t.Fatalf("category = %q, want %q", problem.Category, test.category)
			}
			lower := strings.ToLower(response.Body.String())
			for _, forbidden := range []string{"pq:", "secret", "password", "/var/", "database.go"} {
				if strings.Contains(lower, forbidden) {
					t.Fatalf("problem response leaked %q: %s", forbidden, response.Body.String())
				}
			}
		})
	}
}

func TestGeneratedProblemResponseKeepsStableCategory(t *testing.T) {
	response := httptest.NewRecorder()
	problem := ProbeContract500ApplicationProblemPlusJSONResponse{
		InternalProblemApplicationProblemPlusJSONResponse: InternalProblemApplicationProblemPlusJSONResponse{
			Type: "urn:go-admin-plus:problem:internal", Title: "Internal server error",
			Status: http.StatusInternalServerError, Category: Internal,
			Code: "INTERNAL_ERROR", TraceId: "0123456789abcdef",
		},
	}
	if err := problem.VisitProbeContractResponse(response); err != nil {
		t.Fatalf("write problem response: %v", err)
	}
	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusInternalServerError)
	}
	if strings.Contains(strings.ToLower(response.Body.String()), "sql") {
		t.Fatalf("problem response leaked internal detail: %s", response.Body.String())
	}
}
