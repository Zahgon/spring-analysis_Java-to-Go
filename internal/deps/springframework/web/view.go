package web

import (
	"fmt"
	"html/template"
	"io"
	"os"
	"path"
	"path/filepath"
	"sync"
)

// View renders a model to the response.
type View interface {
	// ContentType is the value the response's Content-Type carries.
	ContentType() string
	// Render writes the view.
	Render(model map[string]any, w io.Writer) error
}

// ViewResolver turns a logical view name into a View, as
// org.springframework.web.servlet.ViewResolver does.
type ViewResolver interface {
	ResolveViewName(name string) (View, error)
}

// UrlBasedViewResolver mirrors
// org.springframework.web.servlet.view.UrlBasedViewResolver: it maps a
// logical view name to prefix + name + suffix.
//
// The original's viewClass is JstlView, which renders a JSP. The Go port
// renders the same markup through html/template — the templates are the JSP
// bodies with the page directive removed, since its only contribution to the
// output was a leading newline, which the template keeps.
type UrlBasedViewResolver struct {
	root      string
	prefix    string
	suffix    string
	viewClass string

	mu    sync.Mutex
	cache map[string]View
}

// NewUrlBasedViewResolver returns a resolver reading templates under root.
func NewUrlBasedViewResolver(root string) *UrlBasedViewResolver {
	return &UrlBasedViewResolver{root: root, cache: map[string]View{}}
}

// SetPrefix sets the prefix prepended to a view name.
func (r *UrlBasedViewResolver) SetPrefix(prefix string) { r.prefix = prefix }

// SetSuffix sets the suffix appended to a view name.
func (r *UrlBasedViewResolver) SetSuffix(suffix string) { r.suffix = suffix }

// SetViewClass records the view implementation the configuration names.
func (r *UrlBasedViewResolver) SetViewClass(viewClass string) { r.viewClass = viewClass }

// GetViewClass returns the configured view implementation.
func (r *UrlBasedViewResolver) GetViewClass() string { return r.viewClass }

// SetRoot sets the directory view paths resolve against.
func (r *UrlBasedViewResolver) SetRoot(root string) { r.root = root }

// ResolveViewName loads prefix + name + suffix.
//
// The original's suffix is ".jsp"; the Go port's templates are ".html", so the
// configured suffix is honoured for the lookup path and the extension is
// swapped. The rendered bytes are what the contract is about, and they are
// unchanged.
func (r *UrlBasedViewResolver) ResolveViewName(name string) (View, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if v, ok := r.cache[name]; ok {
		return v, nil
	}
	rel := path.Clean("/" + r.prefix + name)
	file := filepath.Join(r.root, filepath.FromSlash(rel)+templateSuffix(r.suffix))
	raw, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("Could not resolve view with name '%s': %w", name, err)
	}
	tmpl, err := template.New(name).Parse(string(raw))
	if err != nil {
		return nil, fmt.Errorf("Could not parse view '%s': %w", name, err)
	}
	view := &templateView{tmpl: tmpl}
	r.cache[name] = view
	return view, nil
}

// templateSuffix maps the configured view suffix onto the file the Go port
// ships. A ".jsp" configuration resolves the ".html" template that replaced
// it; any other suffix is used as written.
const (
	jspSuffix       = ".jsp"
	htmlSuffix      = ".html"
	htmlContentType = "text/html;charset=UTF-8"
)

func templateSuffix(suffix string) string {
	if suffix == jspSuffix {
		return htmlSuffix
	}
	return suffix
}

type templateView struct{ tmpl *template.Template }

// ContentType reproduces the JSP page directive's
// contentType="text/html; charset=UTF-8", as the container emits it.
func (v *templateView) ContentType() string { return htmlContentType }

// Render expands the template against the model.
func (v *templateView) Render(model map[string]any, w io.Writer) error {
	return v.tmpl.Execute(w, model)
}
