package web

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/seaswalker/spring-analysis/internal/deps/springframework/validation"
)

// HandlerFunc is a controller method: it fills the model and returns a logical
// view name.
type HandlerFunc func(req *Request) (viewName string, err error)

// RequestMapping mirrors @RequestMapping: a path, and optionally the HTTP
// methods it accepts. An empty Methods accepts any verb, as an unqualified
// @RequestMapping does.
type RequestMapping struct {
	Path    string
	Methods []string
}

// HandlerMapping pairs a mapping with the method that handles it.
type HandlerMapping struct {
	Mapping RequestMapping
	Handler HandlerFunc
}

// Controller is implemented by a bean the original marked @Controller: it
// declares its own request mappings, since Go has no annotations to scan.
type Controller interface {
	// RequestMappings returns each mapping and the method that handles it,
	// in declaration order.
	RequestMappings() []HandlerMapping
}

// PostConstruct is called once, after the controller is registered — the role
// javax.annotation.@PostConstruct plays for the controller's validator setup.
type PostConstruct interface {
	PostConstruct() error
}

// InitBinder mirrors @InitBinder: a hook run before argument binding.
type InitBinder interface {
	InitBinder(binder *DataBinder)
}

// DataBinder mirrors org.springframework.validation.DataBinder as the
// @InitBinder hook sees it: somewhere to register validators.
type DataBinder struct {
	validators []validation.Validator
}

// SetValidator replaces the binder's validators with a single one.
func (b *DataBinder) SetValidator(v validation.Validator) {
	b.validators = []validation.Validator{v}
}

// AddValidators appends validators to the binder.
func (b *DataBinder) AddValidators(vs ...validation.Validator) {
	b.validators = append(b.validators, vs...)
}

// Validators returns the registered validators.
func (b *DataBinder) Validators() []validation.Validator { return b.validators }

// DispatcherServlet routes requests to handlers and renders the view they
// name, mirroring org.springframework.web.servlet.DispatcherServlet.
type DispatcherServlet struct {
	// contextPath is the servlet context the application is deployed under,
	// "/spring" in the original.
	contextPath string
	handlers    map[string]*handlerEntry
	resolver    ViewResolver
	// staticRoot serves resources the way <mvc:default-servlet-handler/>
	// hands unmatched paths to the container's default servlet.
	staticRoot string
	// welcomeFile is served for the context root, as web.xml's
	// <welcome-file-list> specifies.
	welcomeFile string
	binder      *DataBinder
}

type handlerEntry struct {
	handler HandlerFunc
	methods map[string]bool
}

// NewDispatcherServlet returns a dispatcher for the given context path.
func NewDispatcherServlet(contextPath string) *DispatcherServlet {
	return &DispatcherServlet{
		contextPath: normaliseContextPath(contextPath),
		handlers:    map[string]*handlerEntry{},
		binder:      &DataBinder{},
	}
}

// SetViewResolver installs the resolver that turns view names into views.
func (d *DispatcherServlet) SetViewResolver(r ViewResolver) { d.resolver = r }

// SetStaticRoot enables default-servlet handling of static resources from dir.
func (d *DispatcherServlet) SetStaticRoot(dir string) { d.staticRoot = dir }

// SetWelcomeFile names the file served for the context root.
func (d *DispatcherServlet) SetWelcomeFile(name string) { d.welcomeFile = name }

// ContextPath returns the context the dispatcher is mounted at.
func (d *DispatcherServlet) ContextPath() string { return d.contextPath }

// DataBinder returns the binder @InitBinder hooks configured.
func (d *DispatcherServlet) DataBinder() *DataBinder { return d.binder }

// RegisterController registers every mapping a controller declares, and runs
// its @PostConstruct and @InitBinder hooks.
func (d *DispatcherServlet) RegisterController(c Controller) error {
	if pc, ok := c.(PostConstruct); ok {
		if err := pc.PostConstruct(); err != nil {
			return err
		}
	}
	if ib, ok := c.(InitBinder); ok {
		ib.InitBinder(d.binder)
	}
	for _, hm := range c.RequestMappings() {
		if _, dup := d.handlers[hm.Mapping.Path]; dup {
			return fmt.Errorf("Ambiguous mapping. Cannot map handler: there is already a handler for %q", hm.Mapping.Path)
		}
		methods := map[string]bool{}
		for _, verb := range hm.Mapping.Methods {
			methods[strings.ToUpper(verb)] = true
		}
		d.handlers[hm.Mapping.Path] = &handlerEntry{handler: hm.Handler, methods: methods}
	}
	return nil
}

// ServeHTTP dispatches one request.
func (d *DispatcherServlet) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rel, ok := d.strip(r.URL.Path)
	if !ok {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	if rel == "/" || rel == "" {
		d.serveWelcome(w)
		return
	}

	entry, ok := d.handlers[rel]
	if !ok {
		if d.serveStatic(w, rel) {
			return
		}
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	if len(entry.methods) > 0 && !entry.methods[r.Method] {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	req := &Request{HTTP: r, Model: NewModel()}
	viewName, err := entry.handler(req)
	if err != nil {
		writeHandlerError(w, err)
		return
	}
	d.render(w, viewName, req.Model)
}

// strip removes the context path, reporting false when the request is outside
// the deployed context.
func (d *DispatcherServlet) strip(urlPath string) (string, bool) {
	if d.contextPath == "" {
		return urlPath, true
	}
	if urlPath == d.contextPath {
		return "/", true
	}
	if !strings.HasPrefix(urlPath, d.contextPath+"/") {
		return "", false
	}
	return urlPath[len(d.contextPath):], true
}

func (d *DispatcherServlet) serveWelcome(w http.ResponseWriter) {
	if d.welcomeFile == "" || d.staticRoot == "" {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	body, err := os.ReadFile(path.Join(d.staticRoot, d.welcomeFile))
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	writeBody(w, http.StatusOK, "text/html;charset=UTF-8", body)
}

// serveStatic answers from the static root, as the default servlet does.
// Paths under WEB-INF are not public, the way a servlet container hides them.
func (d *DispatcherServlet) serveStatic(w http.ResponseWriter, rel string) bool {
	if d.staticRoot == "" {
		return false
	}
	clean := path.Clean(rel)
	if strings.HasPrefix(strings.ToUpper(clean), "/WEB-INF") {
		return false
	}
	file := path.Join(d.staticRoot, clean)
	body, err := os.ReadFile(file)
	if err != nil {
		return false
	}
	writeBody(w, http.StatusOK, "text/html;charset=UTF-8", body)
	return true
}

// render resolves the view name and writes the rendered body.
func (d *DispatcherServlet) render(w http.ResponseWriter, viewName string, model *Model) {
	if d.resolver == nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	view, err := d.resolver.ResolveViewName(viewName)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	var buf bytes.Buffer
	if err := view.Render(model.Attributes(), &buf); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	writeBody(w, http.StatusOK, view.ContentType(), buf.Bytes())
}

// writeHandlerError maps a handler failure onto the status Spring MVC's
// default exception resolver produces for it.
func writeHandlerError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrUnsupportedMediaType):
		http.Error(w, "Unsupported Media Type", http.StatusUnsupportedMediaType)
	case errors.Is(err, ErrMissingParameter), errors.Is(err, ErrNotReadable):
		http.Error(w, "Bad Request", http.StatusBadRequest)
	default:
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// writeBody writes a response with an explicit length, so the charset the
// container declared is preserved and the body is not chunked.
func writeBody(w http.ResponseWriter, status int, contentType string, body []byte) {
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(body)))
	w.WriteHeader(status)
	_, _ = w.Write(body)
}

func normaliseContextPath(p string) string {
	p = strings.TrimSuffix(p, "/")
	if p != "" && !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return p
}
