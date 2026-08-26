package cache_test

import (
	"context"
	"testing"

	"go-admin/internal/platform/cache"
)

func TestDisabledCachePreservesSourceOfTruthSemantics(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	loads := 0
	loader := func(context.Context) (string, error) {
		loads++
		return "authoritative", nil
	}
	disabled := cache.Disabled[string, string]()

	for range 2 {
		value, err := cache.Resolve(ctx, disabled, "key", loader)
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}
		if value != "authoritative" {
			t.Fatalf("Resolve() = %q, want authoritative", value)
		}
	}
	if loads != 2 {
		t.Fatalf("loader calls = %d, want 2", loads)
	}
}

func TestMemoryCacheMayBeClearedWithoutChangingCorrectness(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	loads := 0
	loader := func(context.Context) (int, error) {
		loads++
		return 42, nil
	}
	memory := cache.NewMemory[string, int]()

	for range 2 {
		value, err := cache.Resolve(ctx, memory, "answer", loader)
		if err != nil || value != 42 {
			t.Fatalf("Resolve() = (%d, %v), want (42, nil)", value, err)
		}
	}
	if loads != 1 {
		t.Fatalf("warm loader calls = %d, want 1", loads)
	}
	memory.Clear()
	value, err := cache.Resolve(ctx, memory, "answer", loader)
	if err != nil || value != 42 || loads != 2 {
		t.Fatalf("Resolve() after Clear = (%d, %v), loads %d", value, err, loads)
	}
}
