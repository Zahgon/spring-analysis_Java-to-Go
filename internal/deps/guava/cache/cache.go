// Package cache reproduces the part of com.google.common.cache the
// application uses: a size-bounded LoadingCache built through a
// CacheBuilder, whose misses are filled by a CacheLoader.
//
// Only the observable behaviour is reproduced — bounded size, load on miss,
// explicit put winning over the loader, and eviction once the bound is
// exceeded. Guava's segmentation, reference types, refresh and statistics are
// not here, because nothing in the application observes them.
package cache

import (
	"container/list"
	"sync"
)

// CacheLoader computes the value for a key that is not resident.
type CacheLoader[K comparable, V any] interface {
	Load(key K) (V, error)
}

// LoaderFunc adapts a function to CacheLoader, the way Guava's
// CacheLoader.from does.
type LoaderFunc[K comparable, V any] func(key K) (V, error)

// Load calls f.
func (f LoaderFunc[K, V]) Load(key K) (V, error) { return f(key) }

// Builder mirrors com.google.common.cache.CacheBuilder.
type Builder[K comparable, V any] struct {
	maximumSize int64
}

// NewBuilder returns an unbounded builder, as CacheBuilder.newBuilder does.
func NewBuilder[K comparable, V any]() *Builder[K, V] {
	return &Builder[K, V]{maximumSize: -1}
}

// MaximumSize bounds the number of entries the cache retains. Guava treats
// this as a ceiling it may enforce a little early, never late; here eviction
// happens exactly when the bound would be exceeded.
func (b *Builder[K, V]) MaximumSize(n int64) *Builder[K, V] {
	b.maximumSize = n
	return b
}

// Build returns a LoadingCache backed by loader.
func (b *Builder[K, V]) Build(loader CacheLoader[K, V]) *LoadingCache[K, V] {
	return &LoadingCache[K, V]{
		loader:      loader,
		maximumSize: b.maximumSize,
		entries:     make(map[K]*list.Element),
		order:       list.New(),
	}
}

type entry[K comparable, V any] struct {
	key   K
	value V
}

// LoadingCache is a bounded cache that fills misses from its CacheLoader.
// It is safe for concurrent use, as Guava's is.
type LoadingCache[K comparable, V any] struct {
	mu          sync.Mutex
	loader      CacheLoader[K, V]
	maximumSize int64
	entries     map[K]*list.Element
	order       *list.List // least-recently-used at the back
}

// Get returns the value for key, loading it if it is not resident.
func (c *LoadingCache[K, V]) Get(key K) (V, error) {
	c.mu.Lock()
	if el, ok := c.entries[key]; ok {
		c.order.MoveToFront(el)
		v := el.Value.(*entry[K, V]).value
		c.mu.Unlock()
		return v, nil
	}
	c.mu.Unlock()

	// The loader runs outside the lock: Guava does not hold a segment lock
	// across a user-supplied load either.
	value, err := c.loader.Load(key)
	if err != nil {
		var zero V
		return zero, err
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if el, ok := c.entries[key]; ok {
		// Another caller loaded it first; that value wins, as Guava's
		// loading-value future does.
		c.order.MoveToFront(el)
		return el.Value.(*entry[K, V]).value, nil
	}
	c.putLocked(key, value)
	return value, nil
}

// GetIfPresent returns the resident value for key without invoking the loader.
func (c *LoadingCache[K, V]) GetIfPresent(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	el, ok := c.entries[key]
	if !ok {
		var zero V
		return zero, false
	}
	c.order.MoveToFront(el)
	return el.Value.(*entry[K, V]).value, true
}

// Put stores value for key, overriding whatever the loader would produce.
func (c *LoadingCache[K, V]) Put(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.putLocked(key, value)
}

// Invalidate discards the entry for key.
func (c *LoadingCache[K, V]) Invalidate(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if el, ok := c.entries[key]; ok {
		c.order.Remove(el)
		delete(c.entries, key)
	}
}

// Size returns the number of resident entries.
func (c *LoadingCache[K, V]) Size() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return int64(len(c.entries))
}

func (c *LoadingCache[K, V]) putLocked(key K, value V) {
	if el, ok := c.entries[key]; ok {
		el.Value.(*entry[K, V]).value = value
		c.order.MoveToFront(el)
		return
	}
	c.entries[key] = c.order.PushFront(&entry[K, V]{key: key, value: value})
	c.evictLocked()
}

// evictLocked drops least-recently-used entries until the bound is respected.
func (c *LoadingCache[K, V]) evictLocked() {
	if c.maximumSize < 0 {
		return
	}
	for int64(len(c.entries)) > c.maximumSize {
		oldest := c.order.Back()
		if oldest == nil {
			return
		}
		c.order.Remove(oldest)
		delete(c.entries, oldest.Value.(*entry[K, V]).key)
	}
}
