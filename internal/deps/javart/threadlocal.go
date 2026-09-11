package javart

import (
	"bytes"
	"runtime"
	"strconv"
	"sync"
)

// ThreadLocal reproduces java.lang.ThreadLocal: a value visible only to the
// goroutine that set it, readable from anywhere in that goroutine's call stack
// without being threaded through the call signatures.
//
// Go deliberately has no goroutine-local storage — context.Context is the
// idiomatic answer, and it is the right answer wherever a value can be passed.
// Spring's AopContext is the case where it cannot: the whole point of
// expose-proxy is that a target method reaches its own proxy without the
// caller handing it one. Keying on the goroutine's identity is the only
// construct that reproduces that, so it lives here, in the shim layer,
// used by exactly one caller.
type ThreadLocal[T any] struct {
	mu     sync.RWMutex
	values map[uint64]T
}

// NewThreadLocal returns an empty ThreadLocal.
func NewThreadLocal[T any]() *ThreadLocal[T] {
	return &ThreadLocal[T]{values: make(map[uint64]T)}
}

// Get returns the calling goroutine's value and whether one was set.
func (t *ThreadLocal[T]) Get() (T, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	v, ok := t.values[goroutineID()]
	return v, ok
}

// Set stores value for the calling goroutine and returns what it replaced,
// mirroring the get-then-set pairing AopContext uses to restore the previous
// proxy when a nested invocation returns.
func (t *ThreadLocal[T]) Set(value T) (old T, had bool) {
	id := goroutineID()
	t.mu.Lock()
	defer t.mu.Unlock()
	old, had = t.values[id]
	t.values[id] = value
	return old, had
}

// Remove clears the calling goroutine's value, as ThreadLocal.remove does.
// Leaving it set would leak: goroutine IDs are reused.
func (t *ThreadLocal[T]) Remove() {
	id := goroutineID()
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.values, id)
}

var goroutinePrefix = []byte("goroutine ")

// goroutineID returns the calling goroutine's runtime identity, parsed from
// the first line of its own stack trace ("goroutine 17 [running]:").
func goroutineID() uint64 {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	line := buf[:n]
	line = bytes.TrimPrefix(line, goroutinePrefix)
	if i := bytes.IndexByte(line, ' '); i >= 0 {
		line = line[:i]
	}
	id, err := strconv.ParseUint(string(line), 10, 64)
	if err != nil {
		return 0
	}
	return id
}
