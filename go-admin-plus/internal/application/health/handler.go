// Package health exposes host-neutral operational HTTP semantics around an
// Application without adding listener or process lifecycle concerns to it.
package health

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"go-admin/internal/application"
)

const (
	LivePath         = "/health/live"
	ReadyPath        = "/health/ready"
	CapabilitiesPath = "/api/v1/runtime/capabilities"
)

// Capabilities is the non-sensitive runtime feature contract returned to
// clients. Connection details and local paths deliberately have no fields.
type Capabilities struct {
	HostProfile   string `json:"hostProfile"`
	Version       string `json:"version"`
	Desktop       bool   `json:"desktop"`
	Offline       bool   `json:"offline"`
	NativeDialogs bool   `json:"nativeDialogs"`
}

// Checker verifies one required runtime dependency. Name is retained for
// diagnostics inside the host but is not exposed by the public handler.
type Checker struct {
	Name  string
	Check func(context.Context) error
}

// Handler wraps the business handler with stable operational endpoints.
type Handler struct {
	state        func() application.State
	capabilities Capabilities
	checkers     []Checker
}

// New validates and snapshots the operational contract.
func New(state func() application.State, capabilities Capabilities, checkers ...Checker) (*Handler, error) {
	if state == nil {
		return nil, errors.New("application state function is required")
	}
	if strings.TrimSpace(capabilities.HostProfile) == "" {
		return nil, errors.New("host profile is required")
	}
	owned := append([]Checker(nil), checkers...)
	for index, checker := range owned {
		if strings.TrimSpace(checker.Name) == "" {
			return nil, errors.New("readiness checker name is required")
		}
		if checker.Check == nil {
			return nil, errors.New("readiness checker function is required")
		}
		owned[index].Name = strings.TrimSpace(checker.Name)
	}
	return &Handler{state: state, capabilities: capabilities, checkers: owned}, nil
}

// Wrap preserves every business route while owning the three operational
// paths, independent of the HTTP router selected by a host.
func (handler *Handler) Wrap(next http.Handler) http.Handler {
	if next == nil {
		next = http.NotFoundHandler()
	}
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case LivePath:
			handler.serveOperation(response, request, func() (int, any) {
				return http.StatusOK, statusResponse{Status: "live"}
			})
		case ReadyPath:
			handler.serveOperation(response, request, func() (int, any) {
				if !handler.ready(request.Context()) {
					return http.StatusServiceUnavailable, statusResponse{Status: "not_ready"}
				}
				return http.StatusOK, statusResponse{Status: "ready"}
			})
		case CapabilitiesPath:
			handler.serveOperation(response, request, func() (int, any) {
				return http.StatusOK, handler.capabilities
			})
		default:
			next.ServeHTTP(response, request)
		}
	})
}

func (handler *Handler) ready(ctx context.Context) bool {
	if handler.state() != application.StateReady {
		return false
	}
	for _, checker := range handler.checkers {
		if err := checker.Check(ctx); err != nil {
			return false
		}
	}
	return true
}

func (handler *Handler) serveOperation(response http.ResponseWriter, request *http.Request, result func() (int, any)) {
	response.Header().Set("Cache-Control", "no-store")
	response.Header().Set("Content-Type", "application/json; charset=utf-8")
	if request.Method != http.MethodGet {
		response.Header().Set("Allow", http.MethodGet)
		response.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(response).Encode(statusResponse{Status: "method_not_allowed"})
		return
	}
	status, body := result()
	response.WriteHeader(status)
	_ = json.NewEncoder(response).Encode(body)
}

type statusResponse struct {
	Status string `json:"status"`
}
