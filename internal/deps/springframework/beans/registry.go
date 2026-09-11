package beans

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"

	"github.com/seaswalker/spring-analysis/internal/deps/javart"
	"github.com/seaswalker/spring-analysis/internal/deps/springframework/aop"
)

// ClassDescriptor is what the JVM read out of a class file and Go has to be
// told: the name the class had in the original, the Go type that replaces it,
// how to construct one, and — for a class the container may need to advise —
// how to wrap it in a typed façade over an AOP proxy.
type ClassDescriptor struct {
	// Name is the fully-qualified Java name, e.g. "aop.SimpleAopBean". It is
	// what config.xml refers to, and what pointcuts match against.
	Name string
	// Type is the Go type instances have.
	Type reflect.Type
	// New constructs an instance.
	New func() any
	// Facade wraps an AOP proxy in an object of the same interface the target
	// satisfies: its method bodies hand their own name to proxy.Invoke, which
	// runs the advisor chain and then the target. It stands in for the
	// subclass CGLIB would have generated. A class with no Facade cannot be
	// advised.
	Facade func(proxy *aop.Proxy) any
}

var (
	registryMu sync.RWMutex
	registry   = map[string]ClassDescriptor{}
	byType     = map[reflect.Type]string{}
)

// Register records a class so that configuration referring to it by its
// original name can be resolved. Packages register their own types from an
// init function, which keeps the mapping next to the type it describes.
//
// Registering the same name twice is a programming error and panics, exactly
// as a duplicate class on the classpath would be a deployment error.
func Register(d ClassDescriptor) {
	registryMu.Lock()
	defer registryMu.Unlock()
	if d.Name == "" {
		panic("beans: ClassDescriptor.Name must not be empty")
	}
	if _, dup := registry[d.Name]; dup {
		panic(fmt.Sprintf("beans: class %q registered twice", d.Name))
	}
	registry[d.Name] = d
	if d.Type != nil {
		byType[d.Type] = d.Name
	}
}

// LookupClass returns the descriptor registered for name.
func LookupClass(name string) (ClassDescriptor, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	d, ok := registry[name]
	return d, ok
}

// RegisteredClasses returns every registered class name, sorted.
func RegisteredClasses() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()
	names := make([]string, 0, len(registry))
	for n := range registry {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// ClassNameOf returns the original Java name registered for a value's type, or
// the Go type name when the value's type was never registered.
func ClassNameOf(v any) string {
	t := reflect.TypeOf(v)
	if t == nil {
		return javart.Null
	}
	registryMu.RLock()
	defer registryMu.RUnlock()
	if name, ok := byType[t]; ok {
		return name
	}
	return t.String()
}

// SimpleClassName renders a value's runtime type the way Java's
// getClass().getSimpleName() does: the class name without its package.
//
// A proxied object reports the generated type's name, which is why the AOP
// demonstration prints something other than the target's own class.
func SimpleClassName(v any) string {
	if p, ok := v.(aop.Proxied); ok {
		return p.ProxyClassName()
	}
	name := ClassNameOf(v)
	if i := strings.LastIndex(name, "."); i >= 0 {
		return name[i+1:]
	}
	return name
}
