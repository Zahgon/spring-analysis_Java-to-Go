package base

import (
	"reflect"

	"github.com/seaswalker/spring-analysis/internal/deps/springframework/beans"
)

// The container resolves classes named in XML through this registry, which is
// the information the JVM would have read out of the class files.
func init() {
	beans.Register(beans.ClassDescriptor{
		Name: "base.Student",
		Type: reflect.TypeOf(&Student{}),
		New:  func() any { return NewStudent() },
	})
	beans.Register(beans.ClassDescriptor{
		Name: "base.SimpleBean",
		Type: reflect.TypeOf(&SimpleBean{}),
		New:  func() any { return NewSimpleBean() },
	})
	beans.Register(beans.ClassDescriptor{
		Name: "base.SimpleBeanFactoryPostProcessor",
		Type: reflect.TypeOf(&SimpleBeanFactoryPostProcessor{}),
		New:  func() any { return NewSimpleBeanFactoryPostProcessor() },
	})
	beans.Register(beans.ClassDescriptor{
		Name: "base.SimpleBeanPostProcessor",
		Type: reflect.TypeOf(&SimpleBeanPostProcessor{}),
		New:  func() any { return NewSimpleBeanPostProcessor() },
	})
}
