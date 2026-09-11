package controller

import (
	"reflect"

	"github.com/seaswalker/spring-analysis/internal/deps/springframework/beans"
)

func init() {
	beans.Register(beans.ClassDescriptor{
		Name: "controller.SimpleController",
		Type: reflect.TypeOf(&SimpleController{}),
		New:  func() any { return NewSimpleController() },
	})
}
