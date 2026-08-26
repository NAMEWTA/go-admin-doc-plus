package reliableruntime_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"go-admin/internal/platform/localcache"
)

func TestLocalCacheIsBoundedClearableAndDisableable(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 8, 26, 15, 0, 0, 0, time.UTC)
	clock := func() time.Time { return now }
	cache, err := localcache.New[string, int](localcache.Options{Capacity: 2, TTL: time.Minute, Now: clock})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	cache.Set("a", 1)
	cache.Set("b", 2)
	if value, ok := cache.Get("a"); !ok || value != 1 {
		t.Fatalf("Get(a) = %v, %v", value, ok)
	}
	cache.Set("c", 3)
	if _, ok := cache.Get("b"); ok || cache.Len() != 2 {
		t.Fatalf("least-recent entry was not evicted, len=%d", cache.Len())
	}
	now = now.Add(2 * time.Minute)
	if _, ok := cache.Get("a"); ok {
		t.Fatal("expired entry returned")
	}
	cache.Clear()
	if cache.Len() != 0 {
		t.Fatalf("Len() after Clear = %d", cache.Len())
	}

	disabled, err := localcache.New[string, int](localcache.Options{Disabled: true})
	if err != nil {
		t.Fatalf("New(disabled) error = %v", err)
	}
	disabled.Set("truth", 42)
	if _, ok := disabled.Get("truth"); ok || disabled.Len() != 0 {
		t.Fatal("disabled cache retained correctness state")
	}
}

func TestLocalCacheDisabledAndClearedUseTheSameSourceOfTruth(t *testing.T) {
	t.Parallel()
	var sourceCalls atomic.Int64
	loader := func(_ context.Context, key string) (string, error) {
		sourceCalls.Add(1)
		if key != "permission:42" {
			return "", errors.New("unknown key")
		}
		return "allowed", nil
	}
	enabled, err := localcache.New[string, string](localcache.Options{Capacity: 4})
	if err != nil {
		t.Fatalf("New(enabled) error = %v", err)
	}
	disabled, err := localcache.New[string, string](localcache.Options{Disabled: true})
	if err != nil {
		t.Fatalf("New(disabled) error = %v", err)
	}
	for _, cache := range []*localcache.Cache[string, string]{enabled, disabled} {
		for attempt := 0; attempt < 2; attempt++ {
			value, err := cache.GetOrLoad(context.Background(), "permission:42", loader)
			if err != nil || value != "allowed" {
				t.Fatalf("GetOrLoad() = %q, %v", value, err)
			}
		}
		cache.Clear()
		value, err := cache.GetOrLoad(context.Background(), "permission:42", loader)
		if err != nil || value != "allowed" {
			t.Fatalf("GetOrLoad() after Clear = %q, %v", value, err)
		}
	}
	if sourceCalls.Load() != 5 {
		t.Fatalf("source calls = %d, want 5", sourceCalls.Load())
	}
	if _, err := enabled.GetOrLoad(context.Background(), "missing", nil); err == nil {
		t.Fatal("nil loader did not return an error")
	}
}

func TestLocalCacheConcurrentAccessRemainsBounded(t *testing.T) {
	t.Parallel()
	cache, err := localcache.New[int, int](localcache.Options{Capacity: 8})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	var workers sync.WaitGroup
	for worker := 0; worker < 16; worker++ {
		workers.Add(1)
		go func(offset int) {
			defer workers.Done()
			for i := 0; i < 100; i++ {
				key := offset*100 + i
				cache.Set(key, key)
				cache.Get(key)
			}
		}(worker)
	}
	workers.Wait()
	if cache.Len() > 8 {
		t.Fatalf("concurrent cache Len() = %d, want <= 8", cache.Len())
	}
}
