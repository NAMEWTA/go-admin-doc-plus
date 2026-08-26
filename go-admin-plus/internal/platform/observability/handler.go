// Package observability exposes host-neutral operational HTTP semantics.
package observability

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"go-admin/internal/app/kernel"
)

const (
	LivePath         = "/health/live"
	ReadyPath        = "/health/ready"
	MetricsPath      = "/metrics"
	CapabilitiesPath = "/api/v1/runtime/capabilities"
	StatusPath       = "/api/v1/runtime/status"
)

// StatusSource supplies a redacted point-in-time lifecycle snapshot.
type StatusSource interface {
	Snapshot() kernel.Snapshot
}

// Capabilities is the non-sensitive runtime feature contract. It intentionally
// has no connection, credential, token, or local-path fields.
type Capabilities struct {
	Profile       string `json:"profile"`
	Version       string `json:"version"`
	Database      string `json:"database"`
	Desktop       bool   `json:"desktop"`
	Offline       bool   `json:"offline"`
	NativeDialogs bool   `json:"nativeDialogs"`
}

// Probe checks one dependency required to accept business requests. Probe names
// and errors are retained in-process and never included in HTTP responses.
type Probe struct {
	Name  string
	Check func(context.Context) error
}

// Handler owns the five operational paths and delegates every other path.
type Handler struct {
	source       StatusSource
	capabilities Capabilities
	probes       []Probe
}

// New validates and snapshots the operational contract.
func New(source StatusSource, capabilities Capabilities, probes ...Probe) (*Handler, error) {
	if source == nil {
		return nil, errors.New("observability status source is required")
	}
	switch capabilities.Profile {
	case "server-postgres":
		if capabilities.Database != "postgres" || capabilities.Desktop || capabilities.Offline || capabilities.NativeDialogs {
			return nil, errors.New("observability server-postgres capabilities are inconsistent")
		}
	case "server-sqlite":
		if capabilities.Database != "sqlite" || capabilities.Desktop || capabilities.Offline || capabilities.NativeDialogs {
			return nil, errors.New("observability server-sqlite capabilities are inconsistent")
		}
	case "desktop-sqlite":
		if capabilities.Database != "sqlite" || !capabilities.Desktop || !capabilities.Offline {
			return nil, errors.New("observability desktop-sqlite capabilities are inconsistent")
		}
	default:
		return nil, errors.New("observability profile is unsupported")
	}
	if strings.TrimSpace(capabilities.Version) == "" {
		return nil, errors.New("observability version is required")
	}
	owned := append([]Probe(nil), probes...)
	seen := make(map[string]struct{}, len(owned))
	for index := range owned {
		owned[index].Name = strings.TrimSpace(owned[index].Name)
		if owned[index].Name == "" || owned[index].Check == nil {
			return nil, fmt.Errorf("observability probe at index %d requires name and check", index)
		}
		if _, exists := seen[owned[index].Name]; exists {
			return nil, fmt.Errorf("duplicate observability probe %q", owned[index].Name)
		}
		seen[owned[index].Name] = struct{}{}
	}
	return &Handler{source: source, capabilities: capabilities, probes: owned}, nil
}

// Wrap preserves business routing while reserving all operational paths.
func (handler *Handler) Wrap(next http.Handler) http.Handler {
	if next == nil {
		next = http.NotFoundHandler()
	}
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if isOperationalPath(request.URL.Path) && request.Method != http.MethodGet {
			serveMethodNotAllowed(response)
			return
		}
		switch request.URL.Path {
		case LivePath:
			handler.serveJSON(response, http.StatusOK, stateBody{Status: "live"})
		case ReadyPath:
			snapshot := handler.source.Snapshot()
			if handler.isReady(request.Context(), snapshot) {
				handler.serveJSON(response, http.StatusOK, stateBody{Status: "ready"})
				return
			}
			handler.serveJSON(response, http.StatusServiceUnavailable, stateBody{Status: "not_ready"})
		case MetricsPath:
			handler.serveMetrics(response, request)
		case CapabilitiesPath:
			handler.serveJSON(response, http.StatusOK, handler.capabilities)
		case StatusPath:
			snapshot := handler.source.Snapshot()
			handler.serveJSON(response, http.StatusOK, statusBody{
				Profile: handler.capabilities.Profile,
				Version: handler.capabilities.Version,
				State:   snapshot.State,
				Failure: snapshot.Failure,
				Ready:   handler.isReady(request.Context(), snapshot),
			})
		default:
			next.ServeHTTP(response, request)
		}
	})
}

func isOperationalPath(path string) bool {
	switch path {
	case LivePath, ReadyPath, MetricsPath, CapabilitiesPath, StatusPath:
		return true
	default:
		return false
	}
}

func serveMethodNotAllowed(response http.ResponseWriter) {
	setNoStore(response)
	response.Header().Set("Content-Type", "application/json; charset=utf-8")
	response.Header().Set("Allow", http.MethodGet)
	response.WriteHeader(http.StatusMethodNotAllowed)
	_ = json.NewEncoder(response).Encode(stateBody{Status: "method_not_allowed"})
}

func (handler *Handler) isReady(ctx context.Context, snapshot kernel.Snapshot) bool {
	if snapshot.State != kernel.StateReady {
		return false
	}
	for _, probe := range handler.probes {
		if err := probe.Check(ctx); err != nil {
			return false
		}
	}
	return true
}

func (handler *Handler) serveJSON(response http.ResponseWriter, status int, body any) {
	setNoStore(response)
	response.Header().Set("Content-Type", "application/json; charset=utf-8")
	response.WriteHeader(status)
	_ = json.NewEncoder(response).Encode(body)
}

func (handler *Handler) serveMetrics(response http.ResponseWriter, request *http.Request) {
	setNoStore(response)
	response.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	snapshot := handler.source.Snapshot()
	ready := 0
	if handler.isReady(request.Context(), snapshot) {
		ready = 1
	}
	_, _ = fmt.Fprintf(response,
		"# TYPE go_admin_runtime_info gauge\n"+
			"go_admin_runtime_info{profile=\"%s\",version=\"%s\",database=\"%s\"} 1\n"+
			"# TYPE go_admin_runtime_ready gauge\n"+
			"go_admin_runtime_ready %d\n"+
			"# TYPE go_admin_runtime_state gauge\n"+
			"go_admin_runtime_state{state=\"%s\"} 1\n",
		metricLabel(handler.capabilities.Profile), metricLabel(handler.capabilities.Version),
		metricLabel(handler.capabilities.Database), ready, metricLabel(string(snapshot.State)))
}

func setNoStore(response http.ResponseWriter) {
	response.Header().Set("Cache-Control", "no-store")
	response.Header().Set("Pragma", "no-cache")
}

func metricLabel(value string) string {
	replacer := strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\"", "\\\"")
	return replacer.Replace(value)
}

type stateBody struct {
	Status string `json:"status"`
}

type statusBody struct {
	Profile string         `json:"profile"`
	Version string         `json:"version"`
	State   kernel.State   `json:"state"`
	Failure kernel.Failure `json:"failure,omitempty"`
	Ready   bool           `json:"ready"`
}
