package javart_test

import (
	"sync"
	"testing"

	"github.com/seaswalker/spring-analysis/internal/deps/javart"
)

// TestThreadLocalIsPerGoroutine covers what AopContext relies on: a value set
// on one goroutine is invisible to another.
func TestThreadLocalIsPerGoroutine(t *testing.T) {
	tl := javart.NewThreadLocal[string]()
	tl.Set("outer")

	var inner string
	var had bool
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		inner, had = tl.Get()
	}()
	wg.Wait()

	if had {
		t.Errorf("another goroutine saw %q", inner)
	}
	if got, ok := tl.Get(); !ok || got != "outer" {
		t.Errorf("Get on the setting goroutine = (%q, %v), want (\"outer\", true)", got, ok)
	}
	tl.Remove()
	if _, ok := tl.Get(); ok {
		t.Error("Remove left the value in place")
	}
}

// TestThreadLocalSetReturnsPrevious covers the get-then-restore pairing a
// nested advised invocation needs.
func TestThreadLocalSetReturnsPrevious(t *testing.T) {
	tl := javart.NewThreadLocal[int]()
	defer tl.Remove()

	if _, had := tl.Set(1); had {
		t.Error("the first Set reported a previous value")
	}
	old, had := tl.Set(2)
	if !had || old != 1 {
		t.Errorf("Set returned (%d, %v), want (1, true)", old, had)
	}
	tl.Set(old)
	if got, _ := tl.Get(); got != 1 {
		t.Errorf("after restoring, Get = %d, want 1", got)
	}
}
