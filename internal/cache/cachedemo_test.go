package cache_test

import (
	"strings"
	"testing"

	"github.com/seaswalker/spring-analysis/internal/cache"
	"github.com/seaswalker/spring-analysis/internal/testsupport"
)

// TestCacheLoader is the original's CacheDemo.cacheLoader: a miss the loader
// fills, then a put whose value wins over the loader.
func TestCacheLoader(t *testing.T) {
	var err error
	lines := testsupport.Lines(testsupport.CaptureStdout(t, func() { err = cache.CacheLoader() }))
	if err != nil {
		t.Fatalf("CacheLoader: %v", err)
	}
	want := []string{"Hello: China", "US"}
	if strings.Join(lines, "\n") != strings.Join(want, "\n") {
		t.Errorf("got:\n%s\nwant:\n%s", strings.Join(lines, "\n"), strings.Join(want, "\n"))
	}
}

// TestCacheBoundedAtMaximumSize covers what the demonstration's trailing
// comment records but never prints: the third distinct key does not enlarge
// the cache.
func TestCacheBoundedAtMaximumSize(t *testing.T) {
	c := cache.NewDemoCache()
	for _, key := range []string{"China", "US", "UK"} {
		if _, err := c.Get(key); err != nil {
			t.Fatalf("Get(%q): %v", key, err)
		}
	}
	if got := c.Size(); got != cache.MaximumSize {
		t.Errorf("size after three distinct keys = %d, want %d", got, cache.MaximumSize)
	}
	if _, present := c.GetIfPresent("China"); present {
		t.Error("the least recently used entry survived eviction")
	}
}

// TestCachePutOverridesLoader checks that an explicit put wins over what the
// loader would have produced — the behaviour the demonstration's second
// printed line depends on.
func TestCachePutOverridesLoader(t *testing.T) {
	c := cache.NewDemoCache()
	c.Put("US", "US")
	got, err := c.Get("US")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got != "US" {
		t.Errorf("Get after Put = %q, want %q (the loader must not run)", got, "US")
	}
}
