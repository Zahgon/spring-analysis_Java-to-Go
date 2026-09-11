// Package web reproduces the Spring MVC surface the application uses:
// DispatcherServlet routing, the @RequestParam / @RequestBody argument
// contracts and the status codes they produce, Model and BindingResult, and
// UrlBasedViewResolver's prefix/suffix view resolution.
package web

// Model mirrors org.springframework.ui.Model: named attributes handed to the
// view.
type Model struct {
	attributes map[string]any
}

// NewModel returns an empty model.
func NewModel() *Model { return &Model{attributes: map[string]any{}} }

// AddAttribute stores value under name and returns the model, so calls chain
// the way Spring's do.
func (m *Model) AddAttribute(name string, value any) *Model {
	m.attributes[name] = value
	return m
}

// Get returns the attribute stored under name.
func (m *Model) Get(name string) (any, bool) {
	v, ok := m.attributes[name]
	return v, ok
}

// Attributes returns a copy of the model's attributes, which is what a view
// renders against.
func (m *Model) Attributes() map[string]any {
	out := make(map[string]any, len(m.attributes))
	for k, v := range m.attributes {
		out[k] = v
	}
	return out
}
