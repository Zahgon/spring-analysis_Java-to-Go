package aop

import (
	"reflect"

	"github.com/seaswalker/spring-analysis/internal/deps/springframework/beans"
)

// The demonstration classes are registered under the names they had, so that
// configuration naming any of them resolves. None of the shipped XML does.
func init() {
	beans.Register(beans.ClassDescriptor{
		Name: "base.aop.AopDemo",
		Type: reflect.TypeOf(&AopDemo{}),
		New:  func() any { return NewAopDemo() },
	})
	beans.Register(beans.ClassDescriptor{
		Name: "base.aop.AopDemoAdvice",
		Type: reflect.TypeOf(&AopDemoAdvice{}),
		New:  func() any { return NewAopDemoAdvice() },
	})
}
