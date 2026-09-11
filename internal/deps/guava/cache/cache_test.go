package cache_test

import (
	"errors"
	"sync"
	"testing"

	"github.com/seaswalker/spring-analysis/internal/deps/guava/cache"
)

func greeter() *cache.LoadingCache[string, string] {
	return cache.NewBuilder[string, string]().MaximumSize(2).
		Build(cache.LoaderFunc[string, string](func(key string) (string, error) {
			return "Hello: " + key, nil
		}))
}

// TestLoadOnMiss covers the loader filling a miss, and not running again for
// a resident key.
func TestLoadOnMiss(t *testing.T) {
	loads := 0
	c := cache.NewBuilder[string, string]().MaximumSize(2).
		Build(cache.LoaderFunc[string, string](func(key string) (string, error) {
			loads++
			return "Hello: " + key, nil
		}))

	for i := 0; i < 3; i++ {
		got, err := c.Get("China")
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		if got != "Hello: China" {
			t.Errorf("Get = %q, want %q", got, "Hello: China")
		}
	}
	if loads != 1 {
		t.Errorf("the loader ran %d times, want 1", loads)
	}
}

// TestMaximumSizeEvicts covers the bound: a third distinct key does not
// enlarge the cache, and the least recently used entry is the one that goes.
func TestMaximumSizeEvicts(t *testing.T) {
	c := greeter()
	if _, err := c.Get("China"); err != nil {
		t.Fatal(err)
	}
	c.Put("US", "US")
	if _, err := c.Get("US"); err != nil {
		t.Fatal(err)
	}
	c.Put("UK", "UK")

	if got := c.Size(); got != 2 {
		t.Errorf("size = %d, want 2", got)
	}
	if _, present := c.GetIfPresent("China"); present {
		t.Error("the least recently used key survived")
	}
	for _, key := range []string{"US", "UK"} {
		if _, present := c.GetIfPresent(key); !present {
			t.Errorf("%q was evicted", key)
		}
	}
}

// TestUnboundedCacheKeepsEverything covers the default builder.
func TestUnboundedCacheKeepsEverything(t *testing.T) {
	c := cache.NewBuilder[string, int]().Build(
		cache.LoaderFunc[string, int](func(string) (int, error) { return 0, nil }))
	for _, k := range []string{"a", "b", "c", "d"} {
		c.Put(k, len(k))
	}
	if got := c.Size(); got != 4 {
		t.Errorf("size = %d, want 4", got)
	}
}

// TestInvalidate removes an entry.
func TestInvalidate(t *testing.T) {
	c := greeter()
	c.Put("US", "US")
	c.Invalidate("US")
	if _, present := c.GetIfPresent("US"); present {
		t.Error("Invalidate left the entry in place")
	}
	c.Invalidate("absent")
}

// TestLoaderError propagates a loader failure and caches nothing.
func TestLoaderError(t *testing.T) {
	boom := errors.New("boom")
	c := cache.NewBuilder[string, string]().MaximumSize(2).
		Build(cache.LoaderFunc[string, string](func(string) (string, error) { return "", boom }))
	if _, err := c.Get("k"); !errors.Is(err, boom) {
		t.Errorf("Get error = %v, want %v", err, boom)
	}
	if got := c.Size(); got != 0 {
		t.Errorf("a failed load cached %d entries", got)
	}
}

// TestConcurrentAccess checks the cache is safe to share, as Guava's is.
func TestConcurrentAccess(t *testing.T) {
	c := greeter()
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if _, err := c.Get("China"); err != nil {
				t.Errorf("Get: %v", err)
			}
			c.Put("US", "US")
		}(i)
	}
	wg.Wait()
	if got := c.Size(); got > 2 {
		t.Errorf("size = %d, want at most 2", got)
	}
}
