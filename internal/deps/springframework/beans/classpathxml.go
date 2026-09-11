package beans

import "path/filepath"

// ClassPath is where configuration and other classpath resources live. The
// original read them out of target/classes, which Maven filled from
// src/main/resources; the Go build has no such copy step, so the resources
// directory is the classpath root.
var ClassPath = "resources"

// ResolveClassPath turns a classpath-relative resource name into a filesystem
// path.
func ResolveClassPath(name string) string {
	return filepath.Join(ClassPath, filepath.FromSlash(name))
}

// NewClassPathXMLApplicationContext loads the named classpath resources and
// refreshes, mirroring
// org.springframework.context.support.ClassPathXmlApplicationContext.
func NewClassPathXMLApplicationContext(locations ...string) (*ApplicationContext, error) {
	ctx := NewApplicationContext()
	reader := NewXMLReader(ctx)
	for _, location := range locations {
		if err := reader.LoadBeanDefinitionsFromFile(ResolveClassPath(location)); err != nil {
			return nil, err
		}
	}
	if err := ctx.Refresh(); err != nil {
		return nil, err
	}
	return ctx, nil
}
