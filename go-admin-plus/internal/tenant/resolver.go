package tenant

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
)

var (
	ErrInvalidTenant = errors.New("invalid tenant configuration")
	ErrUnknownHost   = errors.New("unknown tenant host")
)

// Resolver maps an inbound request to a configured tenant. Implementations
// must not derive tenant IDs directly from untrusted request data.
type Resolver interface {
	Resolve(request *http.Request) (string, error)
}

// FixedResolver always selects one configured tenant and is used by Desktop.
type FixedResolver struct {
	tenantID string
}

// Fixed builds a single-tenant resolver. An empty ID is reported by Resolve.
func Fixed(tenantID string) FixedResolver {
	return FixedResolver{tenantID: strings.TrimSpace(tenantID)}
}

func (r FixedResolver) Resolve(_ *http.Request) (string, error) {
	if r.tenantID == "" {
		return "", ErrInvalidTenant
	}
	return r.tenantID, nil
}

// ServerResolver accepts only constructor-provided host mappings. It does not
// trust forwarding headers; proxy trust must be established by ServerHost.
type ServerResolver struct {
	byHost       map[string]string
	defaultValue string
}

// NewServerResolver validates and copies host-to-tenant mappings. A "*" host
// is an explicit default and is the only way unknown hosts are accepted.
func NewServerResolver(hosts map[string]string) (*ServerResolver, error) {
	if len(hosts) == 0 {
		return nil, fmt.Errorf("%w: at least one host mapping is required", ErrInvalidTenant)
	}
	resolver := &ServerResolver{byHost: make(map[string]string, len(hosts))}
	for rawHost, rawTenant := range hosts {
		tenantID := strings.TrimSpace(rawTenant)
		if tenantID == "" {
			return nil, fmt.Errorf("%w: tenant ID is required", ErrInvalidTenant)
		}
		if strings.TrimSpace(rawHost) == "*" {
			if resolver.defaultValue != "" {
				return nil, fmt.Errorf("%w: duplicate default host", ErrInvalidTenant)
			}
			resolver.defaultValue = tenantID
			continue
		}
		host, err := normalizeHost(rawHost)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInvalidTenant, err)
		}
		if _, exists := resolver.byHost[host]; exists {
			return nil, fmt.Errorf("%w: duplicate host %q", ErrInvalidTenant, host)
		}
		resolver.byHost[host] = tenantID
	}
	return resolver, nil
}

func (r *ServerResolver) Resolve(request *http.Request) (string, error) {
	if request == nil {
		return "", errors.New("request is required")
	}
	host, err := normalizeHost(request.Host)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrUnknownHost, err)
	}
	if tenantID, exists := r.byHost[host]; exists {
		return tenantID, nil
	}
	if r.defaultValue != "" {
		return r.defaultValue, nil
	}
	return "", fmt.Errorf("%w: %s", ErrUnknownHost, host)
}

func normalizeHost(value string) (string, error) {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return "", errors.New("host is required")
	}
	if host, _, err := net.SplitHostPort(value); err == nil {
		value = host
	} else if strings.Contains(value, ":") && !strings.HasPrefix(value, "[") {
		return "", errors.New("host has an invalid port")
	}
	value = strings.Trim(value, "[]")
	if value == "" || strings.ContainsAny(value, "/\\\x00") {
		return "", errors.New("host is invalid")
	}
	return value, nil
}
