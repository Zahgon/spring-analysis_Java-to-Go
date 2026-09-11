package beans

import (
	"fmt"
	"reflect"

	"github.com/seaswalker/spring-analysis/internal/deps/springframework/aop"
)

// ApplicationContext is the container: a definition registry, a singleton
// cache, the extension points, and the refresh lifecycle.
type ApplicationContext struct {
	definitions     map[string]*BeanDefinition
	names           []string // registration order — what getBeanDefinitionNames() returns
	singletons      map[string]any
	scopes          map[string]Scope
	postProcessors  []BeanPostProcessor
	factoryPostProc []BeanFactoryPostProcessor
	advisors        map[string]aop.Advisor
	exposeProxy     bool
	refreshed       bool
	inCreation      map[string]bool
}

// NewApplicationContext returns an empty, unrefreshed context.
func NewApplicationContext() *ApplicationContext {
	return &ApplicationContext{
		definitions: map[string]*BeanDefinition{},
		singletons:  map[string]any{},
		scopes:      map[string]Scope{},
		advisors:    map[string]aop.Advisor{},
		inCreation:  map[string]bool{},
	}
}

// RegisterBeanDefinition adds or replaces a definition. A new name is appended
// to the registration order; re-registering a name keeps its original
// position, as Spring's DefaultListableBeanFactory does.
func (c *ApplicationContext) RegisterBeanDefinition(def *BeanDefinition) {
	if _, exists := c.definitions[def.Name]; !exists {
		c.names = append(c.names, def.Name)
	}
	c.definitions[def.Name] = def
}

// GetBeanDefinition returns the definition registered under name.
func (c *ApplicationContext) GetBeanDefinition(name string) (*BeanDefinition, bool) {
	d, ok := c.definitions[name]
	return d, ok
}

// GetBeanDefinitionNames returns every definition name in registration order.
func (c *ApplicationContext) GetBeanDefinitionNames() []string {
	out := make([]string, len(c.names))
	copy(out, c.names)
	return out
}

// ContainsBean reports whether a definition exists under name.
func (c *ApplicationContext) ContainsBean(name string) bool {
	_, ok := c.definitions[name]
	return ok
}

// RegisterScope makes a custom Scope available to definitions naming it.
func (c *ApplicationContext) RegisterScope(name string, scope Scope) {
	c.scopes[name] = scope
}

// AddBeanPostProcessor registers a BeanPostProcessor.
func (c *ApplicationContext) AddBeanPostProcessor(p BeanPostProcessor) {
	c.postProcessors = append(c.postProcessors, p)
}

// AddBeanFactoryPostProcessor registers a BeanFactoryPostProcessor, which runs
// during refresh before any singleton is created.
func (c *ApplicationContext) AddBeanFactoryPostProcessor(p BeanFactoryPostProcessor) {
	c.factoryPostProc = append(c.factoryPostProc, p)
}

// RegisterAdvisor records an advisor under a name so definitions can reference
// it, matching <aop:advisor advice-ref="..."/>.
func (c *ApplicationContext) RegisterAdvisor(name string, advisor aop.Advisor) {
	c.advisors[name] = advisor
}

// SetExposeProxy sets the <aop:config expose-proxy> flag for proxies this
// context creates.
func (c *ApplicationContext) SetExposeProxy(v bool) { c.exposeProxy = v }

// Refresh runs the BeanFactoryPostProcessors and then eagerly instantiates
// every non-lazy singleton, in registration order — the ordering the
// annotation-configuration demonstration depends on.
func (c *ApplicationContext) Refresh() error {
	for _, p := range c.factoryPostProc {
		if err := p.PostProcessBeanFactory(c); err != nil {
			return err
		}
	}
	for _, name := range c.GetBeanDefinitionNames() {
		def := c.definitions[name]
		if def.Lazy || (def.Scope != "" && def.Scope != ScopeSingleton) {
			continue
		}
		if _, err := c.GetBean(name); err != nil {
			return err
		}
	}
	c.refreshed = true
	return nil
}

// GetBean returns the bean registered under name, creating it if its scope
// says to.
func (c *ApplicationContext) GetBean(name string) (any, error) {
	def, ok := c.definitions[name]
	if !ok {
		return nil, &NoSuchBeanDefinitionError{BeanName: name}
	}
	switch def.Scope {
	case "", ScopeSingleton:
		if bean, cached := c.singletons[name]; cached {
			return bean, nil
		}
		bean, err := c.createBean(def)
		if err != nil {
			return nil, err
		}
		c.singletons[name] = bean
		return bean, nil
	case ScopePrototype:
		return c.createBean(def)
	default:
		scope, ok := c.scopes[def.Scope]
		if !ok {
			return nil, &BeanCreationError{BeanName: name,
				Err: &NoSuchBeanDefinitionError{BeanName: "scope '" + def.Scope + "'"}}
		}
		var createErr error
		bean := scope.Get(name, ObjectFactoryFunc(func() (any, error) {
			b, err := c.createBean(def)
			createErr = err
			return b, err
		}))
		return bean, createErr
	}
}

// GetBeanOfType returns the single bean assignable to t, reproducing
// getBean(Class)'s two failure modes.
func (c *ApplicationContext) GetBeanOfType(t reflect.Type) (any, error) {
	var matches []string
	for _, name := range c.names {
		if typeMatches(c.definitions[name].Type, t) {
			matches = append(matches, name)
		}
	}
	switch len(matches) {
	case 0:
		return nil, &NoSuchBeanDefinitionError{TypeName: JavaTypeName(t)}
	case 1:
		return c.GetBean(matches[0])
	default:
		return nil, &NoUniqueBeanDefinitionError{TypeName: JavaTypeName(t), Names: matches}
	}
}

// typeMatches reports whether a definition of type have satisfies a request
// for want, allowing for pointer receivers and interface targets.
func typeMatches(have, want reflect.Type) bool {
	if have == nil || want == nil {
		return false
	}
	if have == want || have.AssignableTo(want) {
		return true
	}
	if want.Kind() == reflect.Interface && have.Implements(want) {
		return true
	}
	if have.Kind() == reflect.Ptr && have.Elem() == want {
		return true
	}
	if want.Kind() == reflect.Ptr && want.Elem() == have {
		return true
	}
	return false
}

// createBean instantiates a definition, wires the lifecycle callbacks around
// it, and proxies it if any advisor applies.
func (c *ApplicationContext) createBean(def *BeanDefinition) (any, error) {
	if c.inCreation[def.Name] {
		return nil, &BeanCreationError{BeanName: def.Name,
			Err: errCircularReference{name: def.Name}}
	}
	c.inCreation[def.Name] = true
	defer delete(c.inCreation, def.Name)

	bean, err := def.Create(c)
	if err != nil {
		return nil, &BeanCreationError{BeanName: def.Name, Err: err}
	}

	if aware, ok := bean.(BeanNameAware); ok {
		aware.SetBeanName(def.Name)
	}
	if aware, ok := bean.(BeanFactoryAware); ok {
		aware.SetBeanFactory(c)
	}

	for _, p := range c.postProcessors {
		bean, err = p.PostProcessBeforeInitialization(bean, def.Name)
		if err != nil {
			return nil, &BeanCreationError{BeanName: def.Name, Err: err}
		}
	}
	if init, ok := bean.(InitializingBean); ok {
		if err := init.AfterPropertiesSet(); err != nil {
			return nil, &BeanCreationError{BeanName: def.Name, Err: err}
		}
	}
	for _, p := range c.postProcessors {
		bean, err = p.PostProcessAfterInitialization(bean, def.Name)
		if err != nil {
			return nil, &BeanCreationError{BeanName: def.Name, Err: err}
		}
	}

	return c.maybeProxy(def, bean)
}

// maybeProxy wraps the bean in an AOP proxy when advisors match it, returning
// the typed façade callers hold.
func (c *ApplicationContext) maybeProxy(def *BeanDefinition, bean any) (any, error) {
	if bean == nil || len(def.AdvisorNames) == 0 {
		return bean, nil
	}
	advisors := make([]aop.Advisor, 0, len(def.AdvisorNames))
	for _, name := range def.AdvisorNames {
		advisor, ok := c.advisors[name]
		if !ok {
			return nil, &BeanCreationError{BeanName: def.Name,
				Err: &NoSuchBeanDefinitionError{BeanName: name}}
		}
		advisors = append(advisors, advisor)
	}

	descriptor, ok := LookupClass(def.ClassName)
	if !ok || descriptor.Facade == nil {
		return nil, &BeanCreationError{BeanName: def.Name, Err: errNotProxyable{class: def.ClassName}}
	}
	proxy := aop.NewProxy(bean, def.ClassName, advisors, c.exposeProxy)
	facade := descriptor.Facade(proxy)
	proxy.SetFacade(facade)
	return facade, nil
}

type errCircularReference struct{ name string }

func (e errCircularReference) Error() string {
	return "Requested bean '" + e.name + "' is currently in creation: is there an unresolvable circular reference?"
}

// notProxyableFormat reports a bean that is advised but exposes no proxy.
const notProxyableFormat = "class '%s' is advised but registers no proxy fa\u00e7ade"

type errNotProxyable struct{ class string }

func (e errNotProxyable) Error() string {
	return fmt.Sprintf(notProxyableFormat, e.class)
}
