package beans_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/seaswalker/spring-analysis/internal/deps/springframework/beans"
)

type widget struct {
	name  string
	count int
	on    bool
	other *widget
}

func (w *widget) SetName(v string)   { w.name = v }
func (w *widget) SetCount(v int)     { w.count = v }
func (w *widget) SetOn(v bool)       { w.on = v }
func (w *widget) SetOther(v *widget) { w.other = v }
func (w *widget) GetName() string    { return w.name }

func init() {
	beans.Register(beans.ClassDescriptor{
		Name: "test.Widget",
		Type: reflect.TypeOf(&widget{}),
		New:  func() any { return &widget{} },
	})
}

func load(t *testing.T, doc string) *beans.ApplicationContext {
	t.Helper()
	ctx := beans.NewApplicationContext()
	if err := beans.NewXMLReader(ctx).LoadBeanDefinitions(strings.NewReader(doc)); err != nil {
		t.Fatalf("loading definitions: %v", err)
	}
	return ctx
}

const beansOpen = `<beans xmlns="http://www.springframework.org/schema/beans" xmlns:aop="http://www.springframework.org/schema/aop">`

// TestDefinitionNamesFollowRegistrationOrder covers what
// getBeanDefinitionNames reports, including the "#0" suffix an id-less bean
// gets.
func TestDefinitionNamesFollowRegistrationOrder(t *testing.T) {
	ctx := load(t, beansOpen+`
	  <bean id="first" class="test.Widget"/>
	  <bean class="test.Widget"/>
	  <bean class="test.Widget"/>
	</beans>`)

	want := []string{"first", "test.Widget#0", "test.Widget#1"}
	got := ctx.GetBeanDefinitionNames()
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("names = %v, want %v", got, want)
	}
	if !ctx.ContainsBean("first") || ctx.ContainsBean("absent") {
		t.Error("ContainsBean disagrees with the registry")
	}
}

// TestSingletonAndPrototypeScopes covers the two built-in scopes.
func TestSingletonAndPrototypeScopes(t *testing.T) {
	ctx := load(t, beansOpen+`
	  <bean id="single" class="test.Widget"/>
	  <bean id="many" class="test.Widget" scope="prototype"/>
	</beans>`)

	a, _ := ctx.GetBean("single")
	b, _ := ctx.GetBean("single")
	if a != b {
		t.Error("a singleton produced two instances")
	}
	c, _ := ctx.GetBean("many")
	d, _ := ctx.GetBean("many")
	if c == d {
		t.Error("a prototype produced one instance")
	}
}

// TestPropertyValues covers <property> binding through setters, including
// type conversion and refs.
func TestPropertyValues(t *testing.T) {
	ctx := load(t, beansOpen+`
	  <bean id="dep" class="test.Widget"><property name="name" value="dependency"/></bean>
	  <bean id="main" class="test.Widget">
	    <property name="name" value="configured"/>
	    <property name="count" value="42"/>
	    <property name="on" value="true"/>
	    <property name="other" ref="dep"/>
	  </bean>
	</beans>`)

	bean, err := ctx.GetBean("main")
	if err != nil {
		t.Fatalf("GetBean: %v", err)
	}
	w := bean.(*widget)
	if w.name != "configured" || w.count != 42 || !w.on {
		t.Errorf("properties = %+v", w)
	}
	if w.other == nil || w.other.name != "dependency" {
		t.Errorf("ref was not resolved: %+v", w.other)
	}
}

// TestLookupByTypeFailures covers the two messages getBean(Class) produces.
func TestLookupByTypeFailures(t *testing.T) {
	ctx := load(t, beansOpen+`<bean id="a" class="test.Widget"/><bean id="b" class="test.Widget"/></beans>`)

	if _, err := ctx.GetBeanOfType(reflect.TypeOf(&widget{})); err == nil {
		t.Error("an ambiguous lookup succeeded")
	} else if _, ok := err.(*beans.NoUniqueBeanDefinitionError); !ok {
		t.Errorf("error is %T, want *beans.NoUniqueBeanDefinitionError", err)
	}

	empty := beans.NewApplicationContext()
	_, err := empty.GetBeanOfType(reflect.TypeOf(&widget{}))
	var missing *beans.NoSuchBeanDefinitionError
	if !asNoSuchBean(err, &missing) {
		t.Fatalf("error is %T, want *beans.NoSuchBeanDefinitionError", err)
	}
	if want := "No qualifying bean of type 'test.Widget' available"; missing.Error() != want {
		t.Errorf("message = %q, want %q", missing.Error(), want)
	}

	if _, err := empty.GetBean("absent"); err == nil {
		t.Error("a lookup by an unknown name succeeded")
	} else if want := "No bean named 'absent' available"; err.Error() != want {
		t.Errorf("message = %q, want %q", err.Error(), want)
	}
}

func asNoSuchBean(err error, out **beans.NoSuchBeanDefinitionError) bool {
	e, ok := err.(*beans.NoSuchBeanDefinitionError)
	if ok {
		*out = e
	}
	return ok
}

// countingScope records how often the container asked it for an object.
type countingScope struct{ gets int }

func (s *countingScope) Get(name string, f beans.ObjectFactory) any {
	s.gets++
	v, _ := f.GetObject()
	return v
}
func (s *countingScope) Remove(string) any                          { return nil }
func (s *countingScope) RegisterDestructionCallback(string, func()) {}
func (s *countingScope) ResolveContextualObject(string) any         { return nil }
func (s *countingScope) GetConversationID() string                  { return "" }

// TestCustomScope covers a registered Scope taking over instantiation.
func TestCustomScope(t *testing.T) {
	ctx := load(t, beansOpen+`<bean id="scoped" class="test.Widget" scope="one"/></beans>`)
	scope := &countingScope{}
	ctx.RegisterScope("one", scope)

	for i := 0; i < 3; i++ {
		if _, err := ctx.GetBean("scoped"); err != nil {
			t.Fatalf("GetBean: %v", err)
		}
	}
	if scope.gets != 3 {
		t.Errorf("the scope was asked %d times, want 3", scope.gets)
	}

	unregistered := load(t, beansOpen+`<bean id="scoped" class="test.Widget" scope="nope"/></beans>`)
	if _, err := unregistered.GetBean("scoped"); err == nil {
		t.Error("an unregistered scope resolved")
	}
}

// nullProcessor is the shape SimpleBeanPostProcessor has: both callbacks
// return nil.
type nullProcessor struct{ before, after int }

func (p *nullProcessor) PostProcessBeforeInitialization(any, string) (any, error) {
	p.before++
	return nil, nil
}
func (p *nullProcessor) PostProcessAfterInitialization(any, string) (any, error) {
	p.after++
	return nil, nil
}

// TestBeanPostProcessorReturningNilReplacesTheBean covers Spring's contract,
// which the application's own processor depends on.
func TestBeanPostProcessorReturningNilReplacesTheBean(t *testing.T) {
	ctx := load(t, beansOpen+`<bean id="w" class="test.Widget"/></beans>`)
	p := &nullProcessor{}
	ctx.AddBeanPostProcessor(p)

	bean, err := ctx.GetBean("w")
	if err != nil {
		t.Fatalf("GetBean: %v", err)
	}
	if bean != nil {
		t.Errorf("bean = %v, want nil: a processor returning null replaces the bean", bean)
	}
	if p.before != 1 || p.after != 1 {
		t.Errorf("callbacks ran %d/%d times, want 1/1", p.before, p.after)
	}
}

// TestBeanFactoryPostProcessorRunsBeforeSingletons covers ordering during
// refresh.
type renamer struct{ ran bool }

func (r *renamer) PostProcessBeanFactory(f beans.ConfigurableListableBeanFactory) error {
	r.ran = true
	bean, err := f.GetBean("w")
	if err != nil {
		return err
	}
	bean.(*widget).SetName("^_^")
	return nil
}

func TestBeanFactoryPostProcessorRunsDuringRefresh(t *testing.T) {
	ctx := load(t, beansOpen+`<bean id="w" class="test.Widget"/></beans>`)
	r := &renamer{}
	ctx.AddBeanFactoryPostProcessor(r)
	if err := ctx.Refresh(); err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	if !r.ran {
		t.Fatal("the factory post-processor did not run")
	}
	bean, _ := ctx.GetBean("w")
	if got := bean.(*widget).GetName(); got != "^_^" {
		t.Errorf("name = %q, want %q", got, "^_^")
	}
}

// TestXMLErrors covers the configuration mistakes the reader must reject
// rather than ignore.
func TestXMLErrors(t *testing.T) {
	cases := map[string]string{
		"unknown class":       beansOpen + `<bean class="test.Absent"/></beans>`,
		"no class":            beansOpen + `<bean id="x"/></beans>`,
		"unknown element":     beansOpen + `<alias name="a" alias="b"/></beans>`,
		"unknown namespace":   `<beans xmlns="http://www.springframework.org/schema/beans" xmlns:x="urn:x"><x:thing/></beans>`,
		"bad property":        beansOpen + `<bean class="test.Widget"><property name="name"/></bean></beans>`,
		"bad lazy-init":       beansOpen + `<bean class="test.Widget" lazy-init="maybe"/></beans>`,
		"not a beans doc":     `<other/>`,
		"malformed":           `<beans>`,
		"unsupported aop":     beansOpen + `<aop:aspectj-autoproxy/></beans>`,
		"advisor no ref":      beansOpen + `<aop:config><aop:advisor pointcut="execution(* a.B.*(..))"/></aop:config></beans>`,
		"advisor no pointcut": beansOpen + `<aop:config><aop:advisor advice-ref="x"/></aop:config></beans>`,
	}
	for name, doc := range cases {
		t.Run(name, func(t *testing.T) {
			ctx := beans.NewApplicationContext()
			err := beans.NewXMLReader(ctx).LoadBeanDefinitions(strings.NewReader(doc))
			if err == nil {
				t.Errorf("loading %q succeeded, want an error", doc)
			}
		})
	}
}

// TestPropertyErrorsSurfaceAtBeanCreation covers where a bad <property>
// fails. Spring binds property values when it instantiates the bean, not when
// it reads the document, so these two are creation failures rather than parse
// failures.
func TestPropertyErrorsSurfaceAtBeanCreation(t *testing.T) {
	cases := map[string]string{
		"absent setter": beansOpen + `<bean id="w" class="test.Widget"><property name="nope" value="v"/></bean></beans>`,
		"bad number":    beansOpen + `<bean id="w" class="test.Widget"><property name="count" value="x"/></bean></beans>`,
	}
	for name, doc := range cases {
		t.Run(name, func(t *testing.T) {
			ctx := beans.NewApplicationContext()
			if err := beans.NewXMLReader(ctx).LoadBeanDefinitions(strings.NewReader(doc)); err != nil {
				t.Fatalf("parsing must succeed; the failure belongs to creation: %v", err)
			}
			if _, err := ctx.GetBean("w"); err == nil {
				t.Error("creating the bean succeeded, want an error")
			}
			if err := ctx.Refresh(); err == nil {
				t.Error("Refresh succeeded, want the creation failure to propagate")
			}
		})
	}
}

// TestClassRegistry covers the lookup the XML reader resolves class names
// through.
func TestClassRegistry(t *testing.T) {
	if _, ok := beans.LookupClass("test.Widget"); !ok {
		t.Error("the registered class is not resolvable")
	}
	if _, ok := beans.LookupClass("test.Absent"); ok {
		t.Error("an unregistered class resolved")
	}
	names := beans.RegisteredClasses()
	if len(names) == 0 || !sorted(names) {
		t.Errorf("RegisteredClasses = %v, want a sorted non-empty list", names)
	}
	if got := beans.ClassNameOf(&widget{}); got != "test.Widget" {
		t.Errorf("ClassNameOf = %q, want %q", got, "test.Widget")
	}
	if got := beans.ClassNameOf(nil); got != "null" {
		t.Errorf("ClassNameOf(nil) = %q, want %q", got, "null")
	}
	if got := beans.SimpleClassName(&widget{}); got != "Widget" {
		t.Errorf("SimpleClassName = %q, want %q", got, "Widget")
	}
	if got := beans.JavaTypeName(reflect.TypeOf(&widget{})); got != "test.Widget" {
		t.Errorf("JavaTypeName = %q", got)
	}
	if got := beans.JavaTypeName(nil); got != "null" {
		t.Errorf("JavaTypeName(nil) = %q", got)
	}
}

func sorted(s []string) bool {
	for i := 1; i < len(s); i++ {
		if s[i-1] > s[i] {
			return false
		}
	}
	return true
}

// TestDuplicateRegistrationPanics keeps a duplicate class name from being
// silently accepted.
func TestDuplicateRegistrationPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("registering a duplicate class name did not panic")
		}
	}()
	beans.Register(beans.ClassDescriptor{Name: "test.Widget", Type: reflect.TypeOf(&widget{})})
}
