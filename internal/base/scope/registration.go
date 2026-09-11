package scope

import (
	"reflect"

	"github.com/seaswalker/spring-analysis/internal/deps/springframework/beans"
)

func init() {
	beans.Register(beans.ClassDescriptor{
		Name: "base.scope.OneScope",
		Type: reflect.TypeOf(&OneScope{}),
		New:  func() any { return NewOneScope() },
	})
}
