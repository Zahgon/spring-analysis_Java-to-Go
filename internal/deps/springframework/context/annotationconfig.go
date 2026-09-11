// Package context reproduces
// org.springframework.context.annotation.AnnotationConfigApplicationContext:
// registering configuration classes, following their imports, and turning
// their bean methods into definitions with the names and in the order Spring
// produces.
package context

import (
	"reflect"
	"strings"

	"github.com/seaswalker/spring-analysis/internal/deps/springframework/beans"
)

// The definitions Spring registers into every annotation-driven context
// before any user class, in this order. The application prints the list, so
// both the names and the order are contract.
var internalProcessorNames = []string{
	"org.springframework.context.annotation.internalConfigurationAnnotationProcessor",
	"org.springframework.context.annotation.internalAutowiredAnnotationProcessor",
	"org.springframework.context.annotation.internalRequiredAnnotationProcessor",
	"org.springframework.context.annotation.internalCommonAnnotationProcessor",
	"org.springframework.context.event.internalEventListenerProcessor",
	"org.springframework.context.event.internalEventListenerFactory",
}

// internalProcessor is a placeholder bean for one of Spring's own
// infrastructure definitions. The application only ever prints their names, so
// nothing is gained by reimplementing machinery it never observes — but the
// definitions themselves are observable and are therefore real.
type internalProcessor struct{ name string }

// AnnotationMetadata is what ImportAware receives: which class did the
// importing.
type AnnotationMetadata struct {
	// ClassName is the fully-qualified name of the importing configuration
	// class.
	ClassName string
}

// BeanMethod is one @Bean method of a configuration class.
type BeanMethod struct {
	// Name is the bean's name — the method's name, as Spring's default
	// BeanNameGenerator produces it.
	Name string
	// Scope is "singleton" (the default) or "prototype".
	Scope string
	// Type is the Go type the method returns.
	Type reflect.Type
	// Factory invokes the method.
	Factory func() (any, error)
}

// Configuration is implemented by a class the original marked
// @Configuration. Go has no annotations, so a configuration class states its
// own metadata: the name it had, and the bean methods it declares, in
// declaration order.
type Configuration interface {
	// ConfigurationClassName is the fully-qualified Java name of the class.
	ConfigurationClassName() string
	// BeanMethods returns the @Bean methods, in declaration order.
	BeanMethods() []BeanMethod
}

// Importer is implemented by a configuration class the original marked
// @Import.
type Importer interface {
	Imports() []Configuration
}

// ImportAware mirrors
// org.springframework.context.annotation.ImportAware: an imported
// configuration class is told who imported it.
type ImportAware interface {
	SetImportMetadata(metadata AnnotationMetadata)
}

// Autowired is implemented by a configuration class with @Autowired fields.
// It is handed the factory and picks out what it needs.
type Autowired interface {
	Autowire(factory beans.BeanFactory) error
}

// NewAnnotationConfigApplicationContext registers the given configuration
// classes and refreshes, mirroring AnnotationConfigApplicationContext's
// constructor.
//
// Registration order reproduces Spring's exactly:
//
//  1. the six internal annotation processors;
//  2. each configuration class passed in, named by its decapitalised simple
//     name;
//  3. each imported configuration class, named by its fully-qualified name;
//  4. the @Bean definitions, imported classes first — because an importing
//     class's own bean methods are read after the classes it pulled in.
func NewAnnotationConfigApplicationContext(configs ...Configuration) (*beans.ApplicationContext, error) {
	ctx := beans.NewApplicationContext()

	for _, name := range internalProcessorNames {
		processorName := name
		ctx.RegisterBeanDefinition(&beans.BeanDefinition{
			Name: processorName,
			Type: reflect.TypeOf(&internalProcessor{}),
			Create: func(beans.BeanFactory) (any, error) {
				return &internalProcessor{name: processorName}, nil
			},
		})
	}

	// Parse: register each configuration class, following imports depth
	// first. An imported class is registered before the class that imported
	// it finishes parsing, and lands earlier in the bean-method ordering.
	var ordered []configEntry
	seen := map[Configuration]bool{}
	for _, config := range configs {
		registerConfiguration(ctx, config, "", seen, &ordered)
	}

	// Load the @Bean definitions in parse order.
	for _, entry := range ordered {
		for _, method := range entry.config.BeanMethods() {
			factory := method.Factory
			ctx.RegisterBeanDefinition(&beans.BeanDefinition{
				Name:  method.Name,
				Type:  method.Type,
				Scope: method.Scope,
				Create: func(beans.BeanFactory) (any, error) {
					return factory()
				},
			})
		}
	}

	if err := ctx.Refresh(); err != nil {
		return nil, err
	}
	return ctx, nil
}

type configEntry struct {
	name   string
	config Configuration
}

// registerConfiguration registers config and everything it imports.
// importedBy is empty for a class the caller passed in directly, and is the
// importing class's name otherwise — which decides both the definition name
// and whether ImportAware fires.
func registerConfiguration(ctx *beans.ApplicationContext, config Configuration, importedBy string, seen map[Configuration]bool, ordered *[]configEntry) {
	if seen[config] {
		return
	}
	seen[config] = true

	className := config.ConfigurationClassName()
	// A class registered directly is named by its decapitalised simple name;
	// one pulled in by @Import keeps its fully-qualified name.
	name := className
	if importedBy == "" {
		name = decapitalise(simpleName(className))
	}

	instance := config
	ctx.RegisterBeanDefinition(&beans.BeanDefinition{
		Name:      name,
		ClassName: className,
		Type:      reflect.TypeOf(config),
		Create: func(factory beans.BeanFactory) (any, error) {
			if aware, ok := instance.(ImportAware); ok && importedBy != "" {
				aware.SetImportMetadata(AnnotationMetadata{ClassName: importedBy})
			}
			if autowired, ok := instance.(Autowired); ok {
				if err := autowired.Autowire(factory); err != nil {
					return nil, err
				}
			}
			return instance, nil
		},
	})

	if importer, ok := config.(Importer); ok {
		for _, imported := range importer.Imports() {
			registerConfiguration(ctx, imported, className, seen, ordered)
		}
	}

	*ordered = append(*ordered, configEntry{name: name, config: config})
}

func simpleName(className string) string {
	if i := strings.LastIndex(className, "."); i >= 0 {
		return className[i+1:]
	}
	return className
}

func decapitalise(s string) string {
	if s == "" {
		return s
	}
	if len(s) > 1 && s[0] >= 'A' && s[0] <= 'Z' && s[1] >= 'A' && s[1] <= 'Z' {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}
