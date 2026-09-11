package task

import (
	"reflect"

	"github.com/seaswalker/spring-analysis/internal/deps/springframework/beans"
)

func init() {
	beans.Register(beans.ClassDescriptor{
		Name: "task.Task",
		Type: reflect.TypeOf(&Task{}),
		New:  func() any { return NewTask() },
	})
}
