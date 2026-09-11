package aop

import (
	"fmt"

	springaop "github.com/seaswalker/spring-analysis/internal/deps/springframework/aop"
)

// SimpleMethodInterceptor prints the name of the method it advises and then
// lets the invocation continue.
type SimpleMethodInterceptor struct{}

// NewSimpleMethodInterceptor returns the interceptor.
func NewSimpleMethodInterceptor() *SimpleMethodInterceptor { return &SimpleMethodInterceptor{} }

// Invoke prints before proceeding, so its line always precedes the target's.
func (i *SimpleMethodInterceptor) Invoke(invocation springaop.MethodInvocation) ([]any, error) {
	fmt.Println("SimpleMethodInterceptor被调用: " + invocation.GetMethod().Name)
	return invocation.Proceed()
}

var _ springaop.MethodInterceptor = (*SimpleMethodInterceptor)(nil)
