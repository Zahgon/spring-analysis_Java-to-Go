package web_test

import (
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/seaswalker/spring-analysis/internal/deps/springframework/validation"
	"github.com/seaswalker/spring-analysis/internal/deps/springframework/web"
	"github.com/seaswalker/spring-analysis/internal/testsupport"
)

// TestDataBinderRegistersValidators covers the two calls the original's
// @InitBinder body makes — both commented out there, both reproduced here.
func TestDataBinderRegistersValidators(t *testing.T) {
	binder := &web.DataBinder{}
	if len(binder.Validators()) != 0 {
		t.Fatal("a fresh binder has validators")
	}

	binder.SetValidator(stubValidator{})
	if len(binder.Validators()) != 1 {
		t.Errorf("after SetValidator: %d validators, want 1", len(binder.Validators()))
	}
	binder.AddValidators(stubValidator{}, stubValidator{})
	if len(binder.Validators()) != 3 {
		t.Errorf("after AddValidators: %d validators, want 3", len(binder.Validators()))
	}
	binder.SetValidator(stubValidator{})
	if len(binder.Validators()) != 1 {
		t.Errorf("SetValidator did not replace: %d validators, want 1", len(binder.Validators()))
	}
}

type stubValidator struct{}

func (stubValidator) Supports(reflect.Type) bool       { return true }
func (stubValidator) Validate(any, *validation.Errors) {}

// TestDispatcherAccessors covers the context path and the binder the
// dispatcher exposes.
func TestDispatcherAccessors(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"/spring", "/spring"},
		{"/spring/", "/spring"},
		{"spring", "/spring"},
		{"", ""},
	} {
		if got := web.NewDispatcherServlet(tc.in).ContextPath(); got != tc.want {
			t.Errorf("context path for %q = %q, want %q", tc.in, got, tc.want)
		}
	}
	d := web.NewDispatcherServlet("/spring")
	if d.DataBinder() == nil {
		t.Error("the dispatcher exposes no DataBinder")
	}
}

// TestViewResolverReadsItsConfiguration covers the properties
// spring-servlet.xml sets on the resolver, including the viewClass it names.
func TestViewResolverReadsItsConfiguration(t *testing.T) {
	r := web.NewUrlBasedViewResolver(filepath.Join(testsupport.RepoRoot(t), "web"))
	r.SetPrefix("/WEB-INF/")
	r.SetSuffix(".jsp")
	r.SetViewClass("org.springframework.web.servlet.view.JstlView")

	if got := r.GetViewClass(); got != "org.springframework.web.servlet.view.JstlView" {
		t.Errorf("viewClass = %q", got)
	}

	view, err := r.ResolveViewName("echo")
	if err != nil {
		t.Fatalf("ResolveViewName: %v", err)
	}
	if view.ContentType() != "text/html;charset=UTF-8" {
		t.Errorf("Content-Type = %q", view.ContentType())
	}
	var out strings.Builder
	if err := view.Render(map[string]any{"echo": "hello world"}, &out); err != nil {
		t.Fatalf("Render: %v", err)
	}
	want := "\n<html>\n    <body>\n        <h1>hello world</h1>\n    </body>\n</html>\n"
	if out.String() != want {
		t.Errorf("rendered %q, want %q", out.String(), want)
	}

	// A second resolution comes from the cache.
	if again, err := r.ResolveViewName("echo"); err != nil || again != view {
		t.Errorf("the resolver did not cache the view: %v, %v", again, err)
	}
	if _, err := r.ResolveViewName("absent"); err == nil {
		t.Error("an absent view resolved")
	}
}

// TestRequestBodyMediaTypes covers the 415 that precedes any parsing, and the
// 400 an unreadable body produces.
func TestRequestBodyMediaTypes(t *testing.T) {
	var target map[string]any

	form := httptest.NewRequest("POST", "/x", strings.NewReader(`{}`))
	form.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err := (&web.Request{HTTP: form}).RequestBody(&target); err != web.ErrUnsupportedMediaType {
		t.Errorf("wrong media type produced %v, want ErrUnsupportedMediaType", err)
	}

	withCharset := httptest.NewRequest("POST", "/x", strings.NewReader(`{"a":1}`))
	withCharset.Header.Set("Content-Type", "application/json; charset=UTF-8")
	if err := (&web.Request{HTTP: withCharset}).RequestBody(&target); err != nil {
		t.Errorf("a charset parameter was rejected: %v", err)
	}

	empty := httptest.NewRequest("POST", "/x", strings.NewReader("  "))
	empty.Header.Set("Content-Type", "application/json")
	if err := (&web.Request{HTTP: empty}).RequestBody(&target); err != web.ErrNotReadable {
		t.Errorf("an empty body produced %v, want ErrNotReadable", err)
	}
}

// TestRequestParamFallsBackToTheForm covers a parameter arriving in a posted
// form rather than the query string, and the message a missing one carries.
func TestRequestParamFallsBackToTheForm(t *testing.T) {
	form := httptest.NewRequest("POST", "/x", strings.NewReader("name=bob"))
	form.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	got, err := (&web.Request{HTTP: form}).RequestParam("name")
	if err != nil || got != "bob" {
		t.Errorf("RequestParam = (%q, %v), want (\"bob\", nil)", got, err)
	}

	bare := httptest.NewRequest("GET", "/x", nil)
	_, err = (&web.Request{HTTP: bare}).RequestParam("name")
	if err == nil {
		t.Fatal("a missing required parameter succeeded")
	}
	if want := "Required String parameter 'name' is not present"; err.Error() != want {
		t.Errorf("message = %q, want %q", err, want)
	}
	if !strings.Contains(web.ErrMissingParameter.Error(), "Required request parameter") {
		t.Errorf("sentinel message = %q", web.ErrMissingParameter)
	}
}

// TestModel covers the attribute map a handler fills.
func TestModel(t *testing.T) {
	m := web.NewModel()
	if m.AddAttribute("echo", "hello") != m {
		t.Error("AddAttribute does not chain")
	}
	if got, ok := m.Get("echo"); !ok || got != "hello" {
		t.Errorf("Get = (%v, %v)", got, ok)
	}
	if _, ok := m.Get("absent"); ok {
		t.Error("an absent attribute resolved")
	}
	attrs := m.Attributes()
	attrs["echo"] = "mutated"
	if got, _ := m.Get("echo"); got != "hello" {
		t.Error("Attributes returned the model's own map")
	}
}

// TestAmbiguousMappingIsRejected covers two handlers claiming one path.
func TestAmbiguousMappingIsRejected(t *testing.T) {
	d := web.NewDispatcherServlet("/spring")
	if err := d.RegisterController(twoOnOnePath{}); err == nil {
		t.Error("an ambiguous mapping was accepted")
	}
}

type twoOnOnePath struct{}

func (twoOnOnePath) RequestMappings() []web.HandlerMapping {
	handler := func(*web.Request) (string, error) { return "echo", nil }
	return []web.HandlerMapping{
		{Mapping: web.RequestMapping{Path: "/dup"}, Handler: handler},
		{Mapping: web.RequestMapping{Path: "/dup"}, Handler: handler},
	}
}

// TestPostConstructFailureIsReported covers the lifecycle hook's error path.
func TestPostConstructFailureIsReported(t *testing.T) {
	d := web.NewDispatcherServlet("/spring")
	if err := d.RegisterController(failingPostConstruct{}); err == nil {
		t.Error("a failing @PostConstruct was ignored")
	}
}

type failingPostConstruct struct{}

func (failingPostConstruct) RequestMappings() []web.HandlerMapping { return nil }
func (failingPostConstruct) PostConstruct() error                  { return errUnavailable }

var errUnavailable = &unavailableError{}

type unavailableError struct{}

func (*unavailableError) Error() string { return "validator unavailable" }
