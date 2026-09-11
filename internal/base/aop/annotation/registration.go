package annotation

import (
	"reflect"

	"github.com/seaswalker/spring-analysis/internal/deps/springframework/beans"
)

func init() {
	beans.Register(beans.ClassDescriptor{
		Name: "base.aop.annotation.AspectDemo",
		Type: reflect.TypeOf(&AspectDemo{}),
		New:  func() any { return NewAspectDemo() },
	})
}
