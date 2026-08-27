//go:build desktop_native_e2e

package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNativeE2EControlRequestIsExactAndBounded(t *testing.T) {
	for _, action := range []string{"scope-self", "scope-all", "permissions-off", "permissions-on", "session-revoke"} {
		request := httptest.NewRequest(http.MethodPost, "/__desktop/test-control", strings.NewReader(`{"action":"`+action+`"}`))
		request.Header.Set("Content-Type", "application/json")
		got, err := decodeNativeE2EAction(request)
		if err != nil || got != action {
			t.Fatalf("action %q = %q, %v", action, got, err)
		}
	}
	for _, body := range []string{
		`{"action":"unknown"}`,
		`{"action":"scope-all","extra":true}`,
		`{"action":"scope-all"}{"action":"scope-self"}`,
		`{"action":"` + strings.Repeat("a", 130) + `"}`,
	} {
		request := httptest.NewRequest(http.MethodPost, "/__desktop/test-control", strings.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		if _, err := decodeNativeE2EAction(request); err == nil {
			t.Fatalf("accepted invalid body %q", body)
		}
	}
	request := httptest.NewRequest(http.MethodPost, "/__desktop/test-control", strings.NewReader(`{"action":"scope-all"}`))
	if _, err := decodeNativeE2EAction(request); err == nil {
		t.Fatal("accepted request without exact JSON content type")
	}
}
