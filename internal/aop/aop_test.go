package aop_test

import (
	"strings"
	"testing"

	"github.com/seaswalker/spring-analysis/internal/aop"
	springaop "github.com/seaswalker/spring-analysis/internal/deps/springframework/aop"
	"github.com/seaswalker/spring-analysis/internal/deps/springframework/beans"
	"github.com/seaswalker/spring-analysis/internal/testsupport"
)

// context builds the container from the shipped config.xml, with the
// classpath pointed at the repository's resources directory.
func context(t *testing.T) *beans.ApplicationContext {
	t.Helper()
	saved := beans.ClassPath
	beans.ClassPath = testsupport.RepoRoot(t) + "/resources"
	t.Cleanup(func() { beans.ClassPath = saved })

	ctx, err := beans.NewClassPathXMLApplicationContext("config.xml")
	if err != nil {
		t.Fatalf("building the context from config.xml: %v", err)
	}
	return ctx
}

// TestInterceptionOrderAndSelfInvocation is the AOP demonstration's contract:
// the interceptor's line precedes the target's, and the target's re-entry
// through the exposed proxy makes the inner call advised too.
func TestInterceptionOrderAndSelfInvocation(t *testing.T) {
	bean, err := beans.GetBeanAs[aop.SimpleAopBeanAPI](context(t))
	if err != nil {
		t.Fatalf("getting the advised bean: %v", err)
	}
	lines := testsupport.Lines(testsupport.CaptureStdout(t, bean.TestB))
	want := []string{
		"SimpleMethodInterceptor被调用: testB",
		"testB执行",
		"SimpleMethodInterceptor被调用: testC",
		"testC执行",
	}
	if strings.Join(lines, "\n") != strings.Join(want, "\n") {
		t.Errorf("got:\n%s\nwant:\n%s", strings.Join(lines, "\n"), strings.Join(want, "\n"))
	}
}

// TestBooSelfInvocationIsNotAdvised checks the other half of the
// demonstration: Boo reaches TestB on the target directly, so TestB is not
// advised, while the proxy re-entry inside it still is.
func TestBooSelfInvocationIsNotAdvised(t *testing.T) {
	bean, err := beans.GetBeanAs[aop.SimpleAopBeanAPI](context(t))
	if err != nil {
		t.Fatalf("getting the advised bean: %v", err)
	}
	lines := testsupport.Lines(testsupport.CaptureStdout(t, bean.Boo))
	want := []string{
		"SimpleMethodInterceptor被调用: boo",
		"testA执行",
		"testB执行",
		"SimpleMethodInterceptor被调用: testC",
		"testC执行",
	}
	if strings.Join(lines, "\n") != strings.Join(want, "\n") {
		t.Errorf("got:\n%s\nwant:\n%s", strings.Join(lines, "\n"), strings.Join(want, "\n"))
	}
}

// TestProxyTypeIsNotTheTargetType checks what the demonstration's last line
// shows: the object callers hold is a generated proxy, not the target class.
func TestProxyTypeIsNotTheTargetType(t *testing.T) {
	bean, err := beans.GetBeanAs[aop.SimpleAopBeanAPI](context(t))
	if err != nil {
		t.Fatalf("getting the advised bean: %v", err)
	}
	name := beans.SimpleClassName(bean)
	if !strings.HasPrefix(name, "SimpleAopBean$$") {
		t.Errorf("proxy class name = %q, want it to be a generated subclass of SimpleAopBean", name)
	}
	if name == "SimpleAopBean" {
		t.Error("the bean is not proxied")
	}
	if _, ok := springaop.Unwrap(bean).(*aop.SimpleAopBean); !ok {
		t.Errorf("the target behind the proxy is %T, want *aop.SimpleAopBean", springaop.Unwrap(bean))
	}
}

// TestProxyClassNameIsStable checks that two containers name the same proxy
// identically. The original's CGLIB name carried a per-run hash; the port's
// does not, and that is the deliberate difference.
func TestProxyClassNameIsStable(t *testing.T) {
	first, err := beans.GetBeanAs[aop.SimpleAopBeanAPI](context(t))
	if err != nil {
		t.Fatalf("first container: %v", err)
	}
	second, err := beans.GetBeanAs[aop.SimpleAopBeanAPI](context(t))
	if err != nil {
		t.Fatalf("second container: %v", err)
	}
	if a, b := beans.SimpleClassName(first), beans.SimpleClassName(second); a != b {
		t.Errorf("proxy class names differ between runs: %q and %q", a, b)
	}
}

// TestChildOverrideDispatchesVirtually checks that the subclass's narrowed
// method wins when the proxy re-enters, which is the dispatch
// SimpleChildAopBean exists to demonstrate.
func TestChildOverrideDispatchesVirtually(t *testing.T) {
	target := aop.NewSimpleChildAopBean()
	proxy := springaop.NewProxy(target, "aop.SimpleAopBean", []springaop.Advisor{{
		Pointcut: springaop.MustParseExecution("execution(* aop.SimpleAopBean.*(..))"),
		Advice:   aop.NewSimpleMethodInterceptor(),
	}}, true)

	descriptor, ok := beans.LookupClass("aop.SimpleChildAopBean")
	if !ok || descriptor.Facade == nil {
		t.Fatal("aop.SimpleChildAopBean registers no proxy façade")
	}
	facade := descriptor.Facade(proxy)
	proxy.SetFacade(facade)

	lines := testsupport.Lines(testsupport.CaptureStdout(t, facade.(aop.SimpleAopBeanAPI).TestB))
	want := []string{
		"SimpleMethodInterceptor被调用: testB",
		"testB执行",
		"SimpleMethodInterceptor被调用: testC",
		"child testC",
	}
	if strings.Join(lines, "\n") != strings.Join(want, "\n") {
		t.Errorf("got:\n%s\nwant:\n%s", strings.Join(lines, "\n"), strings.Join(want, "\n"))
	}
}

// TestCurrentProxyRequiresAnAdvisedInvocation checks that reaching for the
// current proxy outside one fails, as it does in the original.
func TestCurrentProxyRequiresAnAdvisedInvocation(t *testing.T) {
	if _, err := springaop.CurrentProxy(); err == nil {
		t.Fatal("CurrentProxy succeeded outside an advised invocation")
	}
	defer func() {
		if recover() == nil {
			t.Error("calling TestB on a bare target did not fail")
		}
	}()
	aop.NewSimpleAopBean().TestB()
}
