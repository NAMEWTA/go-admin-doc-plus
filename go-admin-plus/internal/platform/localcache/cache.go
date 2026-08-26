// Package localcache provides an optional bounded process cache. It never owns correctness state.
package localcache

import (
	"container/list"
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

type Loader[K comparable, V any] func(context.Context, K) (V, error)

type Options struct {
	Capacity int
	TTL      time.Duration
	Disabled bool
	Now      func() time.Time
}

type entry[K comparable, V any] struct {
	key       K
	value     V
	expiresAt time.Time
}

type Cache[K comparable, V any] struct {
	mu         sync.Mutex
	items      map[K]*list.Element
	order      *list.List
	capacity   int
	ttl        time.Duration
	disabled   bool
	now        func() time.Time
	generation uint64
	hits       atomic.Uint64
	misses     atomic.Uint64
	loads      atomic.Uint64
	evictions  atomic.Uint64
	clears     atomic.Uint64
}

type Snapshot struct {
	Hits      uint64
	Misses    uint64
	Loads     uint64
	Evictions uint64
	Clears    uint64
	Entries   int
}

func New[K comparable, V any](options Options) (*Cache[K, V], error) {
	if !options.Disabled && (options.Capacity < 1 || options.TTL <= 0) {
		return nil, errors.New("local cache options are invalid")
	}
	now := options.Now
	if now == nil {
		now = time.Now
	}
	return &Cache[K, V]{
		items: make(map[K]*list.Element), order: list.New(), capacity: options.Capacity,
		ttl: options.TTL, disabled: options.Disabled, now: now,
	}, nil
}

func (c *Cache[K, V]) Get(key K) (V, bool) {
	var zero V
	if c == nil {
		return zero, false
	}
	if c.disabled {
		c.misses.Add(1)
		return zero, false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	element, ok := c.items[key]
	if !ok {
		c.misses.Add(1)
		return zero, false
	}
	item := element.Value.(entry[K, V])
	if !item.expiresAt.IsZero() && !c.now().Before(item.expiresAt) {
		c.remove(element)
		c.evictions.Add(1)
		c.misses.Add(1)
		return zero, false
	}
	c.order.MoveToFront(element)
	c.hits.Add(1)
	return item.value, true
}

func (c *Cache[K, V]) Set(key K, value V) {
	if c == nil || c.disabled {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.setLocked(key, value)
}

// GetOrLoad makes the source loader the correctness path. Disabled, empty, expired, and cleared
// caches all execute the same loader and differ only in whether a later call can be served locally.
func (c *Cache[K, V]) GetOrLoad(ctx context.Context, key K, loader Loader[K, V]) (V, error) {
	if value, ok := c.Get(key); ok {
		return value, nil
	}
	var zero V
	if loader == nil {
		return zero, errors.New("local cache loader is required")
	}
	if c == nil {
		return loader(ctx, key)
	}
	c.mu.Lock()
	generation := c.generation
	c.mu.Unlock()
	c.loads.Add(1)
	value, err := loader(ctx, key)
	if err != nil {
		return zero, err
	}
	if c.disabled {
		return value, nil
	}
	c.mu.Lock()
	if c.generation == generation {
		c.setLocked(key, value)
	}
	c.mu.Unlock()
	return value, nil
}

func (c *Cache[K, V]) Clear() {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.generation++
	c.clears.Add(1)
	clear(c.items)
	c.order.Init()
}

func (c *Cache[K, V]) Len() int {
	if c == nil || c.disabled {
		return 0
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.ttl > 0 {
		now := c.now()
		for element := c.order.Back(); element != nil; {
			previous := element.Prev()
			item := element.Value.(entry[K, V])
			if !now.Before(item.expiresAt) {
				c.remove(element)
				c.evictions.Add(1)
			}
			element = previous
		}
	}
	return c.order.Len()
}

// Snapshot returns aggregate performance counters without retaining or exposing cache keys/values.
func (c *Cache[K, V]) Snapshot() Snapshot {
	if c == nil {
		return Snapshot{}
	}
	return Snapshot{
		Hits: c.hits.Load(), Misses: c.misses.Load(), Loads: c.loads.Load(),
		Evictions: c.evictions.Load(), Clears: c.clears.Load(), Entries: c.Len(),
	}
}

func (c *Cache[K, V]) setLocked(key K, value V) {
	expiresAt := c.now().Add(c.ttl)
	if element, ok := c.items[key]; ok {
		element.Value = entry[K, V]{key: key, value: value, expiresAt: expiresAt}
		c.order.MoveToFront(element)
		return
	}
	element := c.order.PushFront(entry[K, V]{key: key, value: value, expiresAt: expiresAt})
	c.items[key] = element
	for c.order.Len() > c.capacity {
		c.remove(c.order.Back())
		c.evictions.Add(1)
	}
}

func (c *Cache[K, V]) remove(element *list.Element) {
	if element == nil {
		return
	}
	item := element.Value.(entry[K, V])
	delete(c.items, item.key)
	c.order.Remove(element)
}
