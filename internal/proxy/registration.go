package proxy

import (
	"reflect"

	"github.com/seaswalker/spring-analysis/internal/deps/springframework/beans"
)

func init() {
	beans.Register(beans.ClassDescriptor{
		Name: "test.proxy.UserServiceImpl",
		Type: reflect.TypeOf(&UserServiceImpl{}),
		New:  func() any { return NewUserServiceImpl() },
	})
}
