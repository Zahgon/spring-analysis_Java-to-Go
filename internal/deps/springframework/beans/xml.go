package beans

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/seaswalker/spring-analysis/internal/deps/springframework/aop"
)

// The namespaces the two configuration files use.
const (
	nsBeans   = "http://www.springframework.org/schema/beans"
	nsAop     = "http://www.springframework.org/schema/aop"
	nsContext = "http://www.springframework.org/schema/context"
	nsMvc     = "http://www.springframework.org/schema/mvc"
)

// XMLElement is a parsed element of a Spring configuration file.
type XMLElement struct {
	XMLName  xml.Name
	Attrs    []xml.Attr   `xml:",any,attr"`
	Children []XMLElement `xml:",any"`
}

// Attr returns the value of the named attribute, and whether it was present.
func (e XMLElement) Attr(name string) (string, bool) {
	for _, a := range e.Attrs {
		if a.Name.Local == name {
			return a.Value, true
		}
	}
	return "", false
}

// AttrOr returns the named attribute or a fallback.
func (e XMLElement) AttrOr(name, fallback string) string {
	if v, ok := e.Attr(name); ok {
		return v
	}
	return fallback
}

// NamespaceHandler extends the reader with support for a non-beans namespace,
// the role org.springframework.beans.factory.xml.NamespaceHandler plays.
// Returning an error for an unrecognised element is deliberate: silently
// ignoring configuration is how a migration ends up quietly doing less than
// the original.
type NamespaceHandler interface {
	// Namespace is the schema URI this handler claims.
	Namespace() string
	// Parse applies one top-level element of that namespace to the context.
	Parse(el XMLElement, ctx *ApplicationContext) error
}

// XMLReader turns a Spring beans XML file into definitions, the role
// XmlBeanDefinitionReader plays.
type XMLReader struct {
	ctx      *ApplicationContext
	handlers map[string]NamespaceHandler
	// anonymousCounts tracks the "#0", "#1" suffixes Spring gives beans that
	// declare no id.
	anonymousCounts map[string]int
}

// NewXMLReader returns a reader that loads into ctx.
func NewXMLReader(ctx *ApplicationContext) *XMLReader {
	r := &XMLReader{ctx: ctx, handlers: map[string]NamespaceHandler{}, anonymousCounts: map[string]int{}}
	r.RegisterHandler(aopNamespaceHandler{})
	return r
}

// RegisterHandler adds support for a namespace.
func (r *XMLReader) RegisterHandler(h NamespaceHandler) { r.handlers[h.Namespace()] = h }

// LoadBeanDefinitionsFromFile reads path.
func (r *XMLReader) LoadBeanDefinitionsFromFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("Loading XML bean definitions from %s: %w", path, err)
	}
	defer f.Close()
	return r.LoadBeanDefinitions(f)
}

// LoadBeanDefinitions reads a <beans> document.
func (r *XMLReader) LoadBeanDefinitions(src io.Reader) error {
	var root XMLElement
	if err := xml.NewDecoder(src).Decode(&root); err != nil {
		return fmt.Errorf("parsing bean definitions: %w", err)
	}
	if root.XMLName.Local != "beans" {
		return fmt.Errorf("expected a <beans> document, got <%s>", root.XMLName.Local)
	}
	for _, child := range root.Children {
		if err := r.parseElement(child); err != nil {
			return err
		}
	}
	return nil
}

func (r *XMLReader) parseElement(el XMLElement) error {
	if el.XMLName.Space == nsBeans || el.XMLName.Space == "" {
		if el.XMLName.Local == "bean" {
			return r.parseBean(el)
		}
		return fmt.Errorf("unsupported beans element <%s>", el.XMLName.Local)
	}
	h, ok := r.handlers[el.XMLName.Space]
	if !ok {
		return fmt.Errorf("no NamespaceHandler for namespace %q (element <%s>)", el.XMLName.Space, el.XMLName.Local)
	}
	return h.Parse(el, r.ctx)
}

// parseBean turns a <bean> element into a definition.
func (r *XMLReader) parseBean(el XMLElement) error {
	className, ok := el.Attr("class")
	if !ok {
		return fmt.Errorf("<bean> without a class attribute is not supported")
	}
	descriptor, ok := LookupClass(className)
	if !ok {
		return fmt.Errorf("Cannot find class [%s]", className)
	}

	name, hasID := el.Attr("id")
	if !hasID {
		if n, hasName := el.Attr("name"); hasName {
			name, hasID = n, true
		}
	}
	if !hasID {
		// Spring names an id-less bean "<class>#<n>".
		name = fmt.Sprintf("%s#%d", className, r.anonymousCounts[className])
		r.anonymousCounts[className]++
	}

	properties := map[string]propertyValue{}
	for _, child := range el.Children {
		if child.XMLName.Local != "property" {
			return fmt.Errorf("unsupported <bean> child <%s>", child.XMLName.Local)
		}
		pname, ok := child.Attr("name")
		if !ok {
			return fmt.Errorf("<property> without a name attribute")
		}
		if v, ok := child.Attr("value"); ok {
			properties[pname] = propertyValue{literal: v}
			continue
		}
		if ref, ok := child.Attr("ref"); ok {
			properties[pname] = propertyValue{ref: ref, isRef: true}
			continue
		}
		return fmt.Errorf("<property name=%q> needs a value or a ref", pname)
	}

	lazy, err := parseBool(el.AttrOr("lazy-init", "false"))
	if err != nil {
		return fmt.Errorf("bean %q: lazy-init: %w", name, err)
	}

	r.ctx.RegisterBeanDefinition(&BeanDefinition{
		Name:      name,
		ClassName: className,
		Type:      descriptor.Type,
		Scope:     el.AttrOr("scope", ScopeSingleton),
		Lazy:      lazy,
		Create: func(f BeanFactory) (any, error) {
			bean := descriptor.New()
			return bean, applyPropertyValues(bean, properties, f)
		},
	})
	return nil
}

type propertyValue struct {
	literal string
	ref     string
	isRef   bool
}

// applyPropertyValues writes <property> values through the bean's setters,
// the way Spring's BeanWrapper does.
func applyPropertyValues(bean any, properties map[string]propertyValue, f BeanFactory) error {
	for name, pv := range properties {
		setter := reflect.ValueOf(bean).MethodByName("Set" + capitalise(name))
		if !setter.IsValid() || setter.Type().NumIn() != 1 {
			return fmt.Errorf("Invalid property '%s' of bean class [%T]: no setter", name, bean)
		}
		var value reflect.Value
		if pv.isRef {
			ref, err := f.GetBean(pv.ref)
			if err != nil {
				return err
			}
			value = reflect.ValueOf(ref)
		} else {
			converted, err := convertLiteral(pv.literal, setter.Type().In(0))
			if err != nil {
				return fmt.Errorf("Failed to convert property value of property '%s': %w", name, err)
			}
			value = converted
		}
		if !value.Type().AssignableTo(setter.Type().In(0)) {
			return fmt.Errorf("Cannot assign %s to property '%s' of type %s", value.Type(), name, setter.Type().In(0))
		}
		setter.Call([]reflect.Value{value})
	}
	return nil
}

// convertLiteral turns an XML attribute string into the setter's parameter
// type, covering the conversions Spring's default editors do for these files.
func convertLiteral(literal string, target reflect.Type) (reflect.Value, error) {
	switch target.Kind() {
	case reflect.String:
		return reflect.ValueOf(literal).Convert(target), nil
	case reflect.Bool:
		b, err := parseBool(literal)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(b).Convert(target), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(literal, 10, 64)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(n).Convert(target), nil
	default:
		return reflect.Value{}, fmt.Errorf("no converter for type %s", target)
	}
}

func parseBool(s string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "true", "on", "yes", "1":
		return true, nil
	case "false", "off", "no", "0", "":
		return false, nil
	default:
		return false, fmt.Errorf("invalid boolean value %q", s)
	}
}

func capitalise(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// aopNamespaceHandler implements <aop:config>.
type aopNamespaceHandler struct{}

func (aopNamespaceHandler) Namespace() string { return nsAop }

// Parse reads <aop:config expose-proxy="..."> and its <aop:advisor> children,
// registering each advisor and attaching it to every bean its pointcut
// matches. Spring does the matching with an auto-proxy creator at
// post-processing time; the effect is the same and is applied here, where the
// definitions are already known.
func (aopNamespaceHandler) Parse(el XMLElement, ctx *ApplicationContext) error {
	if el.XMLName.Local != "config" {
		return fmt.Errorf("unsupported aop element <aop:%s>", el.XMLName.Local)
	}
	exposeProxy, err := parseBool(el.AttrOr("expose-proxy", "false"))
	if err != nil {
		return fmt.Errorf("<aop:config expose-proxy>: %w", err)
	}
	ctx.SetExposeProxy(exposeProxy)

	for _, child := range el.Children {
		if child.XMLName.Local != "advisor" {
			return fmt.Errorf("unsupported aop element <aop:%s>", child.XMLName.Local)
		}
		adviceRef, ok := child.Attr("advice-ref")
		if !ok {
			return fmt.Errorf("<aop:advisor> without advice-ref")
		}
		expression, ok := child.Attr("pointcut")
		if !ok {
			return fmt.Errorf("<aop:advisor> without pointcut")
		}
		pointcut, err := aop.ParseExecution(expression)
		if err != nil {
			return err
		}

		advice, err := ctx.GetBean(adviceRef)
		if err != nil {
			return err
		}
		interceptor, ok := advice.(aop.MethodInterceptor)
		if !ok {
			return fmt.Errorf("advice-ref %q is not a MethodInterceptor", adviceRef)
		}

		advisorName := child.AttrOr("id", adviceRef+"Advisor")
		ctx.RegisterAdvisor(advisorName, aop.Advisor{Pointcut: pointcut, Advice: interceptor})

		for _, name := range ctx.GetBeanDefinitionNames() {
			def, _ := ctx.GetBeanDefinition(name)
			if def.ClassName == "" || name == adviceRef {
				continue
			}
			if pointcut.MatchesClass(def.ClassName) {
				def.AdvisorNames = append(def.AdvisorNames, advisorName)
			}
		}
	}
	return nil
}
