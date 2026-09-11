package proxy_test

import (
	"strings"
	"testing"

	"github.com/seaswalker/spring-analysis/internal/deps/javart"
	"github.com/seaswalker/spring-analysis/internal/proxy"
	"github.com/seaswalker/spring-analysis/internal/testsupport"
)

// TestProxiedCallIsReportedOnce covers the demonstration: the call that enters
// through the proxy is reported, and the one the target makes on itself is
// not.
func TestProxiedCallIsReportedOnce(t *testing.T) {
	target := proxy.NewUserServiceImpl()
	proxied := proxy.NewUserServiceProxy(target, proxy.NewHandler(target))

	lines := testsupport.Lines(testsupport.CaptureStdout(t, proxied.PrintName))
	want := []string{"Method printName is proxyed.", "Name is XXX", "Age: 18"}
	if strings.Join(lines, "\n") != strings.Join(want, "\n") {
		t.Errorf("got:\n%s\nwant:\n%s", strings.Join(lines, "\n"), strings.Join(want, "\n"))
	}
	if strings.Count(strings.Join(lines, "\n"), "is proxyed") != 1 {
		t.Error("the self-call was reported; it must bypass the proxy")
	}
}

// TestEveryInterfaceMethodGoesThroughTheHandler checks the proxy covers the
// whole interface, not just the method the demonstration calls.
func TestEveryInterfaceMethodGoesThroughTheHandler(t *testing.T) {
	target := proxy.NewUserServiceImpl()
	proxied := proxy.NewUserServiceProxy(target, proxy.NewHandler(target))

	lines := testsupport.Lines(testsupport.CaptureStdout(t, proxied.PrintAge))
	want := []string{"Method printAge is proxyed.", "Age: 18"}
	if strings.Join(lines, "\n") != strings.Join(want, "\n") {
		t.Errorf("got:\n%s\nwant:\n%s", strings.Join(lines, "\n"), strings.Join(want, "\n"))
	}
}

// TestUnproxiedTargetPrintsNoHandlerLine is the control case.
func TestUnproxiedTargetPrintsNoHandlerLine(t *testing.T) {
	lines := testsupport.Lines(testsupport.CaptureStdout(t, proxy.NewUserServiceImpl().PrintName))
	want := []string{"Name is XXX", "Age: 18"}
	if strings.Join(lines, "\n") != strings.Join(want, "\n") {
		t.Errorf("got:\n%s\nwant:\n%s", strings.Join(lines, "\n"), strings.Join(want, "\n"))
	}
}

// recordingHandler notes what it was handed without forwarding.
type recordingHandler struct {
	method string
	proxy  any
}

func (h *recordingHandler) Invoke(p any, m javart.Method, args []any) ([]any, error) {
	h.method, h.proxy = m.GetName(), p
	return nil, nil
}

// TestHandlerReceivesTheProxyAndTheOriginalMethodName covers what an
// InvocationHandler is given.
func TestHandlerReceivesTheProxyAndTheOriginalMethodName(t *testing.T) {
	h := &recordingHandler{}
	proxied := proxy.NewUserServiceProxy(proxy.NewUserServiceImpl(), h)
	proxied.PrintName()

	if h.method != "printName" {
		t.Errorf("handler saw method %q, want %q", h.method, "printName")
	}
	if h.proxy != any(proxied) {
		t.Error("the handler was not handed the proxy")
	}
}

// TestDispatcherResolvesTheTarget covers the reflection half of
// java.lang.reflect.Proxy.
func TestDispatcherResolvesTheTarget(t *testing.T) {
	target := proxy.NewUserServiceImpl()
	var seen javart.Method
	d := javart.NewInvocationDispatcher(target, handlerFunc(func(p any, m javart.Method, args []any) ([]any, error) {
		seen = m
		return m.Invoke(nil, args...)
	}))
	if d.Target() != any(target) {
		t.Error("Target does not report the bound object")
	}
	out := testsupport.CaptureStdout(t, func() {
		if _, err := d.Dispatch(nil, "printAge"); err != nil {
			t.Errorf("Dispatch: %v", err)
		}
	})
	if out != "Age: 18\n" {
		t.Errorf("Dispatch printed %q", out)
	}
	if seen.GetName() != "printAge" {
		t.Errorf("method name = %q", seen.GetName())
	}
	if _, err := d.Dispatch(nil, "absent"); err == nil {
		t.Error("dispatching an absent method succeeded")
	}
}

type handlerFunc func(any, javart.Method, []any) ([]any, error)

func (f handlerFunc) Invoke(p any, m javart.Method, args []any) ([]any, error) { return f(p, m, args) }
