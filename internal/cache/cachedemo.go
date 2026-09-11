// Package cache holds the Guava cache demonstration.
package cache

import (
	"fmt"

	"github.com/seaswalker/spring-analysis/internal/deps/guava/cache"
)

// cacheValuePrefix is what the demo's CacheLoader prepends to each key.
const cacheValuePrefix = "Hello: "

// MaximumSize is the bound the demonstration builds the cache with.
const MaximumSize = 2

// NewDemoCache builds the demonstration's cache: bounded to MaximumSize, with
// a loader that greets whatever key it is asked for.
func NewDemoCache() *cache.LoadingCache[string, string] {
	return cache.NewBuilder[string, string]().
		MaximumSize(MaximumSize).
		Build(cache.LoaderFunc[string, string](func(key string) (string, error) {
			return cacheValuePrefix + key, nil
		}))
}

// CacheLoader runs the demonstration: a miss that the loader fills, a put that
// overrides what the loader would have produced, and a third key that pushes
// the cache past its bound.
func CacheLoader() error {
	c := NewDemoCache()

	china, err := c.Get("China")
	if err != nil {
		return err
	}
	fmt.Println(china)

	c.Put("US", "US")
	us, err := c.Get("US")
	if err != nil {
		return err
	}
	fmt.Println(us)

	// The third distinct key does not enlarge the cache: it is bounded at
	// two, so adding this one evicts an entry.
	c.Put("UK", "UK")
	return nil
}
