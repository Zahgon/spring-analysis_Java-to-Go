package controller_test

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/seaswalker/spring-analysis/internal/controller"
	"github.com/seaswalker/spring-analysis/internal/deps/springframework/web"
	"github.com/seaswalker/spring-analysis/internal/testsupport"
)

func newController(t *testing.T) *controller.SimpleController {
	t.Helper()
	c := controller.NewSimpleController()
	if err := c.PostConstruct(); err != nil {
		t.Fatalf("PostConstruct: %v", err)
	}
	return c
}

func jsonRequest(body string) *web.Request {
	r := httptest.NewRequest("POST", "/echoAgain", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	return &web.Request{HTTP: r, Model: web.NewModel()}
}

// TestEchoAgainPrintsViolationsAndStillSucceeds covers the asymmetry that
// makes the demonstration what it is: the violation is printed, the binding
// result stays clean, and the success message is what the view renders.
func TestEchoAgainPrintsViolationsAndStillSucceeds(t *testing.T) {
	c := newController(t)
	req := jsonRequest(`{"name":"bob","age":99}`)

	var view string
	var err error
	lines := testsupport.Lines(testsupport.CaptureStdout(t, func() {
		view, err = c.EchoAgain(req)
	}))
	if err != nil {
		t.Fatalf("EchoAgain: %v", err)
	}
	if view != controller.ViewName {
		t.Errorf("view = %q, want %q", view, controller.ViewName)
	}

	want := []string{"错误消息: 年龄最大不能超过90", "SimpleModel{name='bob', age=99, date='null'}"}
	if strings.Join(lines, "\n") != strings.Join(want, "\n") {
		t.Errorf("stdout:\ngot:  %v\nwant: %v", lines, want)
	}
	if req.BindingResult.HasErrors() {
		t.Error("the explicit validator populated the binding result; the original's does not")
	}
	if got, _ := req.Model.Get("echo"); got != "hello bob, your age is 99." {
		t.Errorf("model attribute = %v, want the success message", got)
	}
}

// TestEchoAgainWithinTheBoundPrintsOnlyTheModel covers the clean path.
func TestEchoAgainWithinTheBoundPrintsOnlyTheModel(t *testing.T) {
	c := newController(t)
	req := jsonRequest(`{"name":"bob","age":30}`)

	lines := testsupport.Lines(testsupport.CaptureStdout(t, func() {
		if _, err := c.EchoAgain(req); err != nil {
			t.Errorf("EchoAgain: %v", err)
		}
	}))
	want := []string{"SimpleModel{name='bob', age=30, date='null'}"}
	if strings.Join(lines, "\n") != strings.Join(want, "\n") {
		t.Errorf("stdout = %v, want %v", lines, want)
	}
	if got, _ := req.Model.Get("echo"); got != "hello bob, your age is 30." {
		t.Errorf("model attribute = %v", got)
	}
}

// TestEchoAgainWithAnAbsentAge renders the boxed null.
func TestEchoAgainWithAnAbsentAge(t *testing.T) {
	c := newController(t)
	req := jsonRequest(`{"name":"bob"}`)
	testsupport.CaptureStdout(t, func() {
		if _, err := c.EchoAgain(req); err != nil {
			t.Errorf("EchoAgain: %v", err)
		}
	})
	if got, _ := req.Model.Get("echo"); got != "hello bob, your age is null." {
		t.Errorf("model attribute = %v", got)
	}
}

// TestEchoAgainRejectsAnUnreadableBody covers the binding failure.
func TestEchoAgainRejectsAnUnreadableBody(t *testing.T) {
	c := newController(t)
	req := jsonRequest(`{"name":"bob","date":"2020-01-02 03:04:05"}`)
	testsupport.CaptureStdout(t, func() {
		if _, err := c.EchoAgain(req); err == nil {
			t.Error("an unreadable body bound successfully")
		}
	})
}

// TestEcho covers the greeting handler and its required parameter.
func TestEcho(t *testing.T) {
	c := newController(t)

	req := &web.Request{HTTP: httptest.NewRequest("GET", "/echo?name=world", nil), Model: web.NewModel()}
	view, err := c.Echo(req)
	if err != nil {
		t.Fatalf("Echo: %v", err)
	}
	if view != controller.ViewName {
		t.Errorf("view = %q", view)
	}
	if got, _ := req.Model.Get("echo"); got != "hello world" {
		t.Errorf("model attribute = %v", got)
	}

	missing := &web.Request{HTTP: httptest.NewRequest("GET", "/echo", nil), Model: web.NewModel()}
	if _, err := c.Echo(missing); err == nil {
		t.Error("Echo succeeded without the required parameter")
	}
}

// TestInitBinderRegistersNothing keeps the hook empty, as the original's
// commented-out body leaves it.
func TestInitBinderRegistersNothing(t *testing.T) {
	binder := &web.DataBinder{}
	newController(t).InitBinder(binder)
	if got := binder.Validators(); len(got) != 0 {
		t.Errorf("the @InitBinder hook registered %d validators, want none", len(got))
	}
}

// TestRequestMappings keeps the declared routes and their verbs.
func TestRequestMappings(t *testing.T) {
	mappings := newController(t).RequestMappings()
	if len(mappings) != 2 {
		t.Fatalf("declared %d mappings, want 2", len(mappings))
	}
	var paths []string
	for _, hm := range mappings {
		m := hm.Mapping
		paths = append(paths, m.Path)
		if m.Path == "/echoAgain" && strings.Join(m.Methods, ",") != "POST" {
			t.Errorf("/echoAgain methods = %v, want [POST]", m.Methods)
		}
		if m.Path == "/echo" && len(m.Methods) != 0 {
			t.Errorf("/echo methods = %v, want any verb", m.Methods)
		}
	}
	if strings.Join(paths, ",") != "/echo,/echoAgain" {
		t.Errorf("paths = %v, want them in declaration order", paths)
	}
}
