package base

import (
	"reflect"

	"github.com/seaswalker/spring-analysis/internal/deps/springframework/beans"
)

// SimpleBeanFactoryPostProcessor renames the wired student once the
// definitions are loaded and before any singleton is created.
type SimpleBeanFactoryPostProcessor struct{}

// NewSimpleBeanFactoryPostProcessor returns the processor.
func NewSimpleBeanFactoryPostProcessor() *SimpleBeanFactoryPostProcessor {
	return &SimpleBeanFactoryPostProcessor{}
}

// PostProcessBeanFactory fetches the SimpleBean and sets its student's name.
func (p *SimpleBeanFactoryPostProcessor) PostProcessBeanFactory(factory beans.ConfigurableListableBeanFactory) error {
	bean, err := factory.GetBeanOfType(reflect.TypeOf(&SimpleBean{}))
	if err != nil {
		return err
	}
	bean.(*SimpleBean).GetStudent().SetName("^_^")
	return nil
}

// SimpleBeanPostProcessor returns nil from both callbacks.
//
// That is what the original does, and it is not an oversight to correct: a
// BeanPostProcessor returning null replaces the bean with null, and the class
// exists to show what happens when one does. Returning the bean instead would
// quietly change the demonstration into a different one.
type SimpleBeanPostProcessor struct{}

// NewSimpleBeanPostProcessor returns the processor.
func NewSimpleBeanPostProcessor() *SimpleBeanPostProcessor { return &SimpleBeanPostProcessor{} }

// PostProcessBeforeInitialization returns nil.
func (p *SimpleBeanPostProcessor) PostProcessBeforeInitialization(bean any, beanName string) (any, error) {
	return nil, nil
}

// PostProcessAfterInitialization returns nil.
func (p *SimpleBeanPostProcessor) PostProcessAfterInitialization(bean any, beanName string) (any, error) {
	return nil, nil
}

var (
	_ beans.BeanFactoryPostProcessor = (*SimpleBeanFactoryPostProcessor)(nil)
	_ beans.BeanPostProcessor        = (*SimpleBeanPostProcessor)(nil)
)
