package generator

import (
	"strings"
	"sync"
	"testing"
)

func TestMinimalCommandEnvironmentForcesOfflineGoPolicy(t *testing.T) {
	for key, value := range map[string]string{
		"GOENV":       "/tmp/untrusted-go-env",
		"GOTOOLCHAIN": "auto",
		"GOPROXY":     "https://proxy.invalid",
		"GOSUMDB":     "sum.invalid",
	} {
		t.Setenv(key, value)
	}
	environmentOnce = sync.Once{}
	baseEnvironment = nil
	t.Cleanup(func() {
		environmentOnce = sync.Once{}
		baseEnvironment = nil
	})

	environment := minimalCommandEnvironment(t.TempDir())
	values := make(map[string]string, len(environment))
	for _, item := range environment {
		key, value, found := strings.Cut(item, "=")
		if found {
			values[key] = value
		}
	}

	for key, expected := range map[string]string{
		"GOENV":       "off",
		"GOTOOLCHAIN": "local",
		"GOPROXY":     "off",
		"GOSUMDB":     "off",
	} {
		if values[key] != expected {
			t.Fatalf("%s = %q, want %q", key, values[key], expected)
		}
	}
}
