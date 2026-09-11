package beans

import (
	"reflect"

	"github.com/seaswalker/spring-analysis/internal/deps/javart"
)

// BeanDefinition describes one bean the container can produce.
type BeanDefinition struct {
	// Name is the definition's registered name. Order of registration is the
	// order GetBeanDefinitionNames reports.
	Name string
	// ClassName is the original Java class name, when the definition came
	// from one.
	ClassName string
	// Type is the Go type instances have, used by lookup-by-type.
	Type reflect.Type
	// Scope is "singleton", "prototype", or the name of a registered custom
	// scope. Empty means singleton.
	Scope string
	// Create builds a raw instance, before post-processing.
	Create func(factory BeanFactory) (any, error)
	// Advisors, when non-empty, means the bean is proxied.
	AdvisorNames []string
	// Lazy suppresses eager instantiation during refresh.
	Lazy bool
}

// BeanFactory is the lookup surface a bean or a post-processor sees.
type BeanFactory interface {
	// GetBean returns the bean registered under name.
	GetBean(name string) (any, error)
	// GetBeanOfType returns the single bean whose type is assignable to t.
	GetBeanOfType(t reflect.Type) (any, error)
	// ContainsBean reports whether a definition exists under name.
	ContainsBean(name string) bool
	// GetBeanDefinitionNames returns every definition name, in registration
	// order.
	GetBeanDefinitionNames() []string
}

// ConfigurableListableBeanFactory is the surface a BeanFactoryPostProcessor
// sees: lookup plus definition access and scope registration.
type ConfigurableListableBeanFactory interface {
	BeanFactory
	GetBeanDefinition(name string) (*BeanDefinition, bool)
	RegisterBeanDefinition(def *BeanDefinition)
	RegisterScope(name string, scope Scope)
	AddBeanPostProcessor(p BeanPostProcessor)
}

// GetBeanAs is GetBeanOfType with the assertion the caller would otherwise
// write, which is how getBean(Class) reads at a Java call site.
func GetBeanAs[T any](f BeanFactory) (T, error) {
	var zero T
	target := reflect.TypeOf(zero)
	if target == nil {
		// T is an interface; recover its type through a pointer.
		target = reflect.TypeOf((*T)(nil)).Elem()
	}
	bean, err := f.GetBeanOfType(target)
	if err != nil {
		return zero, err
	}
	typed, ok := bean.(T)
	if !ok {
		return zero, &NoSuchBeanDefinitionError{TypeName: JavaTypeName(target)}
	}
	return typed, nil
}

// JavaTypeName renders a Go type using the original's Java name when the type
// is registered, so error messages keep the text the original printed.
func JavaTypeName(t reflect.Type) string {
	if t == nil {
		return javart.Null
	}
	registryMu.RLock()
	defer registryMu.RUnlock()
	if name, ok := byType[t]; ok {
		return name
	}
	if t.Kind() == reflect.Ptr {
		if name, ok := byType[t.Elem()]; ok {
			return name
		}
	}
	if name, ok := byType[reflect.PtrTo(t)]; ok {
		return name
	}
	return t.String()
}
