// Package webapp assembles the web application: it reads the Spring MVC
// configuration, registers the scanned controllers, and returns the
// dispatcher.
//
// It stands in for the servlet container plus web.xml. The container itself is
// out of scope — a Go program serves HTTP directly — but everything web.xml
// configured has an equivalent here: the context path, the welcome file, the
// UTF-8 encoding the CharacterEncodingFilter forced, and the location of the
// Spring configuration.
package webapp

import (
	"fmt"

	"github.com/seaswalker/spring-analysis/internal/controller"
	"github.com/seaswalker/spring-analysis/internal/deps/springframework/beans"
	"github.com/seaswalker/spring-analysis/internal/deps/springframework/web"
)

// The deployment settings web.xml and the Maven Tomcat plugin fixed.
const (
	// ContextPath is the path the application is deployed under.
	ContextPath = "/spring"
	// DefaultPort is the port the tomcat7 plugin bound.
	DefaultPort = 8080
	// ConfigLocation is web.xml's contextConfigLocation, relative to the
	// classpath root.
	ConfigLocation = "spring-servlet.xml"
	// WelcomeFile is web.xml's single <welcome-file>.
	WelcomeFile = "index.html"
)

// controllersByPackage maps a component-scan base package to the controllers
// it contains.
//
// A JVM scan reads the classpath and finds @Controller classes by itself. Go
// has no runtime class discovery, so the scan is resolved against what the
// binary was linked with — which is the same set, written down.
var controllersByPackage = map[string][]func() web.Controller{
	"controller": {func() web.Controller { return controller.NewSimpleController() }},
}

// New reads the MVC configuration and returns a configured dispatcher.
// viewRoot is where the view templates live.
func New(viewRoot string) (*web.DispatcherServlet, error) {
	config := &web.MvcConfiguration{}

	ctx := beans.NewApplicationContext()
	reader := beans.NewXMLReader(ctx)
	reader.RegisterHandler(web.ContextNamespaceHandler{Config: config})
	reader.RegisterHandler(web.MvcNamespaceHandler{Config: config})
	if err := reader.LoadBeanDefinitionsFromFile(beans.ResolveClassPath(ConfigLocation)); err != nil {
		return nil, err
	}
	if err := ctx.Refresh(); err != nil {
		return nil, err
	}

	resolver, err := beans.GetBeanAs[*web.UrlBasedViewResolver](ctx)
	if err != nil {
		return nil, fmt.Errorf("no ViewResolver configured in %s: %w", ConfigLocation, err)
	}
	resolver.SetRoot(viewRoot)

	dispatcher := web.NewDispatcherServlet(ContextPath)
	dispatcher.SetViewResolver(resolver)
	dispatcher.SetWelcomeFile(WelcomeFile)
	if config.DefaultServletHandler {
		dispatcher.SetStaticRoot(viewRoot)
	}

	for _, pkg := range config.ComponentScanPackages {
		factories, ok := controllersByPackage[pkg]
		if !ok {
			return nil, fmt.Errorf("component scan of package %q found no controllers", pkg)
		}
		for _, factory := range factories {
			if err := dispatcher.RegisterController(factory()); err != nil {
				return nil, err
			}
		}
	}
	return dispatcher, nil
}
