package aop

import (
	"fmt"

	springaop "github.com/seaswalker/spring-analysis/internal/deps/springframework/aop"
)

// AopDemoAdvice carries the before/after advice methods an <aop:aspect> would
// bind, and is itself a MethodInterceptor.
type AopDemoAdvice struct{}

// NewAopDemoAdvice returns the advice.
func NewAopDemoAdvice() *AopDemoAdvice { return &AopDemoAdvice{} }

// BeforeSend is the before-advice for Send.
func (a *AopDemoAdvice) BeforeSend() { fmt.Println("before send") }

// AfterSend is the after-advice for Send.
func (a *AopDemoAdvice) AfterSend() { fmt.Println("after send") }

// BeforeReceive is the before-advice for Receive.
func (a *AopDemoAdvice) BeforeReceive() { fmt.Println("before receive") }

// AfterReceive is the after-advice for Receive.
func (a *AopDemoAdvice) AfterReceive() { fmt.Println("after receive") }

// Invoke returns nil without proceeding.
//
// The original returns null here too, and it is deliberate: an interceptor
// that never calls proceed() suppresses the target. Calling proceed would make
// this a different demonstration.
func (a *AopDemoAdvice) Invoke(invocation springaop.MethodInvocation) ([]any, error) {
	return nil, nil
}

var _ springaop.MethodInterceptor = (*AopDemoAdvice)(nil)
