package aop_test

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/seaswalker/spring-analysis/internal/deps/springframework/aop"
)

// target is a stand-in advised object.
type target struct{ calls []string }

func (t *target) Hello() { t.calls = append(t.calls, "hello") }
func (t *target) Sum(a, b int) int {
	t.calls = append(t.calls, "sum")
	return a + b
}

// recorder is an interceptor that notes the method name and proceeds.
type recorder struct{ seen *[]string }

func (r recorder) Invoke(inv aop.MethodInvocation) ([]any, error) {
	*r.seen = append(*r.seen, inv.GetMethod().Name)
	return inv.Proceed()
}

// suppressor is an interceptor that never proceeds.
type suppressor struct{}

func (suppressor) Invoke(aop.MethodInvocation) ([]any, error) { return nil, nil }

func advisor(expr string, advice aop.MethodInterceptor) aop.Advisor {
	return aop.Advisor{Pointcut: aop.MustParseExecution(expr), Advice: advice}
}

// TestProxyRunsTheChainThenTheTarget covers the ordering an interceptor
// depends on, and the results coming back out.
func TestProxyRunsTheChainThenTheTarget(t *testing.T) {
	tgt := &target{}
	var seen []string
	p := aop.NewProxy(tgt, "p.Target", []aop.Advisor{advisor("execution(* p.Target.*(..))", recorder{&seen})}, false)

	out, err := p.Invoke("sum", 2, 3)
	if err != nil {
		t.Fatalf("Invoke: %v", err)
	}
	if len(out) != 1 || out[0] != 5 {
		t.Errorf("Invoke returned %v, want [5]", out)
	}
	if strings.Join(seen, ",") != "sum" {
		t.Errorf("interceptor saw %v, want [sum]", seen)
	}
	if strings.Join(tgt.calls, ",") != "sum" {
		t.Errorf("target saw %v, want [sum]", tgt.calls)
	}
}

// TestNonMatchingAdvisorIsNotApplied checks the pointcut actually filters.
func TestNonMatchingAdvisorIsNotApplied(t *testing.T) {
	tgt := &target{}
	var seen []string
	p := aop.NewProxy(tgt, "p.Target", []aop.Advisor{advisor("execution(* p.Target.hello(..))", recorder{&seen})}, false)

	if _, err := p.Invoke("sum", 1, 1); err != nil {
		t.Fatal(err)
	}
	if len(seen) != 0 {
		t.Errorf("the interceptor ran for an unmatched method: %v", seen)
	}
	if _, err := p.Invoke("hello"); err != nil {
		t.Fatal(err)
	}
	if strings.Join(seen, ",") != "hello" {
		t.Errorf("interceptor saw %v, want [hello]", seen)
	}
}

// TestInterceptorCanSuppressTheTarget covers an interceptor that never calls
// Proceed — which is what AopDemoAdvice does.
func TestInterceptorCanSuppressTheTarget(t *testing.T) {
	tgt := &target{}
	p := aop.NewProxy(tgt, "p.Target", []aop.Advisor{advisor("execution(* p.Target.*(..))", suppressor{})}, false)
	if _, err := p.Invoke("hello"); err != nil {
		t.Fatal(err)
	}
	if len(tgt.calls) != 0 {
		t.Errorf("the target ran despite the interceptor not proceeding: %v", tgt.calls)
	}
}

// TestExposeProxyPublishesToTheGoroutine covers what makes self-invocation
// re-entry work, and that the exposure is scoped to the invocation.
func TestExposeProxyPublishesToTheGoroutine(t *testing.T) {
	tgt := &target{}
	var inside any
	var insideErr error
	watcher := interceptorFunc(func(inv aop.MethodInvocation) ([]any, error) {
		inside, insideErr = aop.CurrentProxy()
		return inv.Proceed()
	})
	p := aop.NewProxy(tgt, "p.Target", []aop.Advisor{advisor("execution(* p.Target.*(..))", watcher)}, true)
	facade := "the facade"
	p.SetFacade(facade)

	if _, err := p.Invoke("hello"); err != nil {
		t.Fatal(err)
	}
	if insideErr != nil {
		t.Fatalf("CurrentProxy inside the invocation: %v", insideErr)
	}
	if inside != facade {
		t.Errorf("CurrentProxy = %v, want the façade", inside)
	}
	if _, err := aop.CurrentProxy(); err == nil {
		t.Error("the proxy stayed exposed after the invocation returned")
	}
}

// TestExposeProxyRestoresAnOuterInvocation covers the nested case: an inner
// advised call must not leave the outer proxy clobbered.
func TestExposeProxyRestoresAnOuterInvocation(t *testing.T) {
	inner := aop.NewProxy(&target{}, "p.Inner", nil, true)
	inner.SetFacade("inner")

	var afterNested any
	outerAdvice := interceptorFunc(func(inv aop.MethodInvocation) ([]any, error) {
		if _, err := inner.Invoke("hello"); err != nil {
			return nil, err
		}
		afterNested, _ = aop.CurrentProxy()
		return inv.Proceed()
	})
	outer := aop.NewProxy(&target{}, "p.Target", []aop.Advisor{advisor("execution(* p.Target.*(..))", outerAdvice)}, true)
	outer.SetFacade("outer")

	if _, err := outer.Invoke("hello"); err != nil {
		t.Fatal(err)
	}
	if afterNested != "outer" {
		t.Errorf("after the nested invocation, CurrentProxy = %v, want \"outer\"", afterNested)
	}
	if _, err := aop.CurrentProxy(); err == nil {
		t.Error("the proxy stayed exposed after the outer invocation returned")
	}
}

// TestWithoutExposeProxyNothingIsPublished is the other half of the flag.
func TestWithoutExposeProxyNothingIsPublished(t *testing.T) {
	var insideErr error
	watcher := interceptorFunc(func(inv aop.MethodInvocation) ([]any, error) {
		_, insideErr = aop.CurrentProxy()
		return inv.Proceed()
	})
	p := aop.NewProxy(&target{}, "p.Target", []aop.Advisor{advisor("execution(* p.Target.*(..))", watcher)}, false)
	if _, err := p.Invoke("hello"); err != nil {
		t.Fatal(err)
	}
	if insideErr == nil {
		t.Error("a proxy created without expose-proxy published itself")
	}
}

// TestExposeProxyIsPerGoroutine checks the exposure does not leak across
// concurrent invocations.
func TestExposeProxyIsPerGoroutine(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			name := fmt.Sprintf("facade-%d", i)
			var inside any
			watcher := interceptorFunc(func(inv aop.MethodInvocation) ([]any, error) {
				inside, _ = aop.CurrentProxy()
				return inv.Proceed()
			})
			p := aop.NewProxy(&target{}, "p.Target", []aop.Advisor{advisor("execution(* p.Target.*(..))", watcher)}, true)
			p.SetFacade(name)
			if _, err := p.Invoke("hello"); err != nil {
				t.Error(err)
				return
			}
			if inside != name {
				t.Errorf("goroutine %d saw %v, want %q", i, inside, name)
			}
		}(i)
	}
	wg.Wait()
}

// TestInvocationExposesItsContext covers the accessors an interceptor reads.
func TestInvocationExposesItsContext(t *testing.T) {
	tgt := &target{}
	var method aop.Method
	var args []any
	var this any
	watcher := interceptorFunc(func(inv aop.MethodInvocation) ([]any, error) {
		method, args, this = inv.GetMethod(), inv.GetArguments(), inv.GetThis()
		return inv.Proceed()
	})
	p := aop.NewProxy(tgt, "p.Target", []aop.Advisor{advisor("execution(* p.Target.*(..))", watcher)}, false)
	if _, err := p.Invoke("sum", 4, 5); err != nil {
		t.Fatal(err)
	}
	if method.Name != "sum" || method.DeclaringClass != "p.Target" {
		t.Errorf("method = %+v", method)
	}
	if len(args) != 2 || args[0] != 4 || args[1] != 5 {
		t.Errorf("arguments = %v, want [4 5]", args)
	}
	if this != any(tgt) {
		t.Errorf("this = %v, want the target", this)
	}
}

// TestInvokeUnknownMethod reports rather than panicking.
func TestInvokeUnknownMethod(t *testing.T) {
	p := aop.NewProxy(&target{}, "p.Target", nil, false)
	if _, err := p.Invoke("absent"); err == nil {
		t.Error("invoking an absent method succeeded")
	}
}

// TestProxyClassNameShape covers the generated name the AOP demonstration
// prints, and that different advice yields a different name.
func TestProxyClassNameShape(t *testing.T) {
	plain := aop.NewProxy(&target{}, "p.Target", nil, false)
	advised := aop.NewProxy(&target{}, "p.Target", []aop.Advisor{advisor("execution(* p.Target.*(..))", suppressor{})}, false)

	if !strings.HasPrefix(plain.ClassName(), "Target$$EnhancerBySpringGoAOP$$") {
		t.Errorf("class name = %q", plain.ClassName())
	}
	if plain.ClassName() == advised.ClassName() {
		t.Error("proxies with different advice share a class name")
	}
	if plain.TargetClass() != "p.Target" {
		t.Errorf("TargetClass = %q", plain.TargetClass())
	}
	if _, ok := plain.Target().(*target); !ok {
		t.Errorf("Target = %T", plain.Target())
	}
}

// TestUnwrap returns the target behind a façade and passes plain values
// through.
func TestUnwrap(t *testing.T) {
	if got := aop.Unwrap("plain"); got != "plain" {
		t.Errorf("Unwrap of an unproxied value = %v", got)
	}
}

// TestErrNoProxyExposedMessage keeps the message the original raised.
func TestErrNoProxyExposedMessage(t *testing.T) {
	if !strings.Contains(aop.ErrNoProxyExposed.Error(), "Cannot find current proxy") {
		t.Errorf("message = %q", aop.ErrNoProxyExposed)
	}
}

// interceptorFunc adapts a function to MethodInterceptor.
type interceptorFunc func(aop.MethodInvocation) ([]any, error)

func (f interceptorFunc) Invoke(inv aop.MethodInvocation) ([]any, error) { return f(inv) }
