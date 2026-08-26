package desktop

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strings"
)

var desktopOrigins = map[string]struct{}{
	"wails://wails":          {},
	"http://wails.localhost": {},
}

type Gateway struct {
	token string
}

func NewGateway(token string) (*Gateway, error) {
	if strings.TrimSpace(token) == "" {
		return nil, errors.New("desktop launch token is required")
	}
	return &Gateway{token: token}, nil
}

func (gateway *Gateway) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if !isLoopbackRemoteAddress(request.RemoteAddr) {
			writeGatewayError(writer, http.StatusForbidden, "desktop request source is not loopback")
			return
		}
		origin := request.Header.Get("Origin")
		if _, allowed := desktopOrigins[origin]; !allowed {
			writeGatewayError(writer, http.StatusForbidden, "desktop request origin is not allowed")
			return
		}
		setDesktopCORS(writer.Header(), origin)
		if request.Method == http.MethodOptions {
			writer.WriteHeader(http.StatusNoContent)
			return
		}
		provided := request.Header.Get(LaunchTokenHeader)
		if subtle.ConstantTimeCompare([]byte(provided), []byte(gateway.token)) != 1 {
			writeGatewayError(writer, http.StatusUnauthorized, "desktop launch token is invalid")
			return
		}
		next.ServeHTTP(writer, request)
	})
}

func isLoopbackRemoteAddress(address string) bool {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return false
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func setDesktopCORS(headers http.Header, origin string) {
	headers.Set("Access-Control-Allow-Origin", origin)
	headers.Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
	headers.Set("Access-Control-Allow-Headers", "Authorization, Content-Type, "+LaunchTokenHeader)
	headers.Add("Vary", "Origin")
}

func writeGatewayError(writer http.ResponseWriter, status int, message string) {
	writer.Header().Set("Content-Type", "application/json")
	writer.Header().Set("Cache-Control", "no-store")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(map[string]any{"code": status, "msg": message})
}
