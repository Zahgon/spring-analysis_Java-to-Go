package validator

import (
	"reflect"

	"github.com/seaswalker/spring-analysis/internal/deps/springframework/beans"
)

func init() {
	beans.Register(beans.ClassDescriptor{
		Name: "validator.SimpleModelValidator",
		Type: reflect.TypeOf(&SimpleModelValidator{}),
		New:  func() any { return NewSimpleModelValidator() },
	})
}
