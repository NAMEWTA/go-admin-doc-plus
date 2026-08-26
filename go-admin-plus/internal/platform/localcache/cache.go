// Package localcache provides an optional bounded process cache. It never owns correctness state.
package localcache

import (
	"container/list"
	"context"
	"errors"
	"sync"
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
	mu       sync.Mutex
	items    map[K]*list.Element
	order    *list.List
	capacity int
	ttl      time.Duration
	disabled bool
	now      func() time.Time
}

func New[K comparable, V any](options Options) (*Cache[K, V], error) {
	if !options.Disabled && (options.Capacity < 1 || options.TTL < 0) {
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
	if c == nil || c.disabled {
		return zero, false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	element, ok := c.items[key]
	if !ok {
		return zero, false
	}
	item := element.Value.(entry[K, V])
	if !item.expiresAt.IsZero() && !c.now().Before(item.expiresAt) {
		c.remove(element)
		return zero, false
	}
	c.order.MoveToFront(element)
	return item.value, true
}

func (c *Cache[K, V]) Set(key K, value V) {
	if c == nil || c.disabled {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	expiresAt := time.Time{}
	if c.ttl > 0 {
		expiresAt = c.now().Add(c.ttl)
	}
	if element, ok := c.items[key]; ok {
		element.Value = entry[K, V]{key: key, value: value, expiresAt: expiresAt}
		c.order.MoveToFront(element)
		return
	}
	element := c.order.PushFront(entry[K, V]{key: key, value: value, expiresAt: expiresAt})
	c.items[key] = element
	for c.order.Len() > c.capacity {
		c.remove(c.order.Back())
	}
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
	value, err := loader(ctx, key)
	if err != nil {
		return zero, err
	}
	c.Set(key, value)
	return value, nil
}

func (c *Cache[K, V]) Clear() {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
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
			}
			element = previous
		}
	}
	return c.order.Len()
}

func (c *Cache[K, V]) remove(element *list.Element) {
	if element == nil {
		return
	}
	item := element.Value.(entry[K, V])
	delete(c.items, item.key)
	c.order.Remove(element)
}
