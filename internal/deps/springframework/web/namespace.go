package web

import (
	"fmt"
	"reflect"

	"github.com/seaswalker/spring-analysis/internal/deps/springframework/beans"
)

// ViewRoot is where view templates live. The original resolved them against
// the web application root; the Go port ships them under web/.
var ViewRoot = "web"

func init() {
	beans.Register(beans.ClassDescriptor{
		Name: "org.springframework.web.servlet.view.UrlBasedViewResolver",
		Type: reflect.TypeOf(&UrlBasedViewResolver{}),
		New:  func() any { return NewUrlBasedViewResolver(ViewRoot) },
	})
}

// MvcConfiguration is what <context:component-scan>, <mvc:annotation-driven/>
// and <mvc:default-servlet-handler/> record. The dispatcher reads it when it
// is wired up.
type MvcConfiguration struct {
	// ComponentScanPackages are the packages <context:component-scan> names.
	ComponentScanPackages []string
	// AnnotationDriven records <mvc:annotation-driven/>.
	AnnotationDriven bool
	// DefaultServletHandler records <mvc:default-servlet-handler/>.
	DefaultServletHandler bool
}

// ContextNamespaceHandler implements the <context:*> elements the
// configuration uses.
type ContextNamespaceHandler struct{ Config *MvcConfiguration }

// Namespace returns the spring-context schema URI.
func (ContextNamespaceHandler) Namespace() string {
	return "http://www.springframework.org/schema/context"
}

// Parse handles <context:component-scan base-package="..."/>.
func (h ContextNamespaceHandler) Parse(el beans.XMLElement, _ *beans.ApplicationContext) error {
	if el.XMLName.Local != "component-scan" {
		return fmt.Errorf("unsupported context element <context:%s>", el.XMLName.Local)
	}
	pkg, ok := el.Attr("base-package")
	if !ok {
		return fmt.Errorf("<context:component-scan> without base-package")
	}
	h.Config.ComponentScanPackages = append(h.Config.ComponentScanPackages, pkg)
	return nil
}

// MvcNamespaceHandler implements the <mvc:*> elements the configuration uses.
type MvcNamespaceHandler struct{ Config *MvcConfiguration }

// Namespace returns the spring-mvc schema URI.
func (MvcNamespaceHandler) Namespace() string {
	return "http://www.springframework.org/schema/mvc"
}

// Parse handles <mvc:annotation-driven/> and <mvc:default-servlet-handler/>.
func (h MvcNamespaceHandler) Parse(el beans.XMLElement, _ *beans.ApplicationContext) error {
	switch el.XMLName.Local {
	case "annotation-driven":
		h.Config.AnnotationDriven = true
	case "default-servlet-handler":
		h.Config.DefaultServletHandler = true
	default:
		return fmt.Errorf("unsupported mvc element <mvc:%s>", el.XMLName.Local)
	}
	return nil
}
