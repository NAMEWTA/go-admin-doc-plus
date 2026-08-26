// Package cache provides an optional process-local acceleration boundary.
package cache

import (
	"context"
	"sync"
)

// Cache is deliberately small: absence, disablement, and eviction are all normal cache misses.
type Cache[K comparable, V any] interface {
	Get(context.Context, K) (V, bool)
	Set(context.Context, K, V)
}

// Resolve preserves source-of-truth behavior regardless of cache state.
func Resolve[K comparable, V any](ctx context.Context, store Cache[K, V], key K, loader func(context.Context) (V, error)) (V, error) {
	if value, ok := store.Get(ctx, key); ok {
		return value, nil
	}
	value, err := loader(ctx)
	if err != nil {
		var zero V
		return zero, err
	}
	store.Set(ctx, key, value)
	return value, nil
}

type disabled[K comparable, V any] struct{}

// Disabled returns a cache that performs no I/O and always misses.
func Disabled[K comparable, V any]() Cache[K, V] { return disabled[K, V]{} }

func (disabled[K, V]) Get(context.Context, K) (V, bool) {
	var zero V
	return zero, false
}
func (disabled[K, V]) Set(context.Context, K, V) {}

// Memory is a clearable, process-local cache. It is an optimization and owns no correctness state.
type Memory[K comparable, V any] struct {
	mu     sync.RWMutex
	values map[K]V
}

func NewMemory[K comparable, V any]() *Memory[K, V] {
	return &Memory[K, V]{values: make(map[K]V)}
}

func (m *Memory[K, V]) Get(_ context.Context, key K) (V, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	value, ok := m.values[key]
	return value, ok
}

func (m *Memory[K, V]) Set(_ context.Context, key K, value V) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.values[key] = value
}

func (m *Memory[K, V]) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.values = make(map[K]V)
}
