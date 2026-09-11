package beans

// BeanPostProcessor mirrors
// org.springframework.beans.factory.config.BeanPostProcessor.
//
// Spring's contract says a processor returning null replaces the bean with
// null, and it does not "helpfully" substitute the original. That matters
// here: the application ships a processor that returns null from both
// callbacks, and reproducing the null is the point of it.
type BeanPostProcessor interface {
	PostProcessBeforeInitialization(bean any, beanName string) (any, error)
	PostProcessAfterInitialization(bean any, beanName string) (any, error)
}

// BeanFactoryPostProcessor mirrors
// org.springframework.beans.factory.config.BeanFactoryPostProcessor: it runs
// against the definitions, before any singleton is instantiated.
type BeanFactoryPostProcessor interface {
	PostProcessBeanFactory(factory ConfigurableListableBeanFactory) error
}

// InitializingBean is called after a bean's properties are set.
type InitializingBean interface {
	AfterPropertiesSet() error
}

// DisposableBean is called when the context closes.
type DisposableBean interface {
	Destroy() error
}

// BeanNameAware receives the name its definition was registered under.
type BeanNameAware interface {
	SetBeanName(name string)
}

// BeanFactoryAware receives the factory that created it.
type BeanFactoryAware interface {
	SetBeanFactory(factory BeanFactory)
}
