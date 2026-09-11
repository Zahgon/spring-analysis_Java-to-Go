package javaconfig

import (
	"reflect"

	"github.com/seaswalker/spring-analysis/internal/base"
	"github.com/seaswalker/spring-analysis/internal/deps/springframework/beans"
	springctx "github.com/seaswalker/spring-analysis/internal/deps/springframework/context"
)

// SimpleBeanConfigClassName is the name the original class had.
const SimpleBeanConfigClassName = "java_config.SimpleBeanConfig"

// SimpleBeanConfig is the @Configuration that imports StudentConfig and
// declares the SimpleBean bean.
type SimpleBeanConfig struct {
	// studentConfig is the @Autowired field. It is filled by Autowire when
	// the container instantiates this class.
	studentConfig *StudentConfig
}

// NewSimpleBeanConfig returns the configuration.
func NewSimpleBeanConfig() *SimpleBeanConfig { return &SimpleBeanConfig{} }

// ConfigurationClassName returns the original class name.
func (c *SimpleBeanConfig) ConfigurationClassName() string { return SimpleBeanConfigClassName }

// Imports declares @Import(StudentConfig.class).
func (c *SimpleBeanConfig) Imports() []springctx.Configuration {
	return []springctx.Configuration{NewStudentConfig()}
}

// Autowire resolves the @Autowired StudentConfig field out of the container,
// which is what AutowiredAnnotationBeanPostProcessor does.
func (c *SimpleBeanConfig) Autowire(factory beans.BeanFactory) error {
	config, err := beans.GetBeanAs[*StudentConfig](factory)
	if err != nil {
		return err
	}
	c.studentConfig = config
	return nil
}

// SimpleBean is the @Bean factory method. It calls the imported
// configuration's bean method directly, so a prototype Student is created for
// it — which is also what forces the student definition to be registered.
func (c *SimpleBeanConfig) SimpleBean() *base.SimpleBean {
	return base.NewSimpleBeanOf(c.studentConfig.Student())
}

// BeanMethods declares SimpleBean as a singleton bean named for the method.
func (c *SimpleBeanConfig) BeanMethods() []springctx.BeanMethod {
	return []springctx.BeanMethod{{
		Name:    "simpleBean",
		Type:    reflect.TypeOf(&base.SimpleBean{}),
		Factory: func() (any, error) { return c.SimpleBean(), nil },
	}}
}

var (
	_ springctx.Configuration = (*SimpleBeanConfig)(nil)
	_ springctx.Importer      = (*SimpleBeanConfig)(nil)
	_ springctx.Autowired     = (*SimpleBeanConfig)(nil)
)
