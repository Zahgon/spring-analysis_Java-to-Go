// Package annotation holds the @Aspect demonstration.
package annotation

import (
	"fmt"

	springaop "github.com/seaswalker/spring-analysis/internal/deps/springframework/aop"
)

// AspectDemo is the @Aspect the notes describe: a named pointcut, and a
// @Before advice bound to it.
//
// AspectJ's annotations are strings evaluated by a weaver; the pointcut
// expression carries over verbatim and is parsed by the same pointcut parser
// the XML configuration uses.
type AspectDemo struct{}

// NewAspectDemo returns the aspect.
func NewAspectDemo() *AspectDemo { return &AspectDemo{} }

// PointcutExpression is the expression the original's
// @Pointcut("execution(void base.aop.AopDemo.send(..))") declares, under the
// name the annotated method gave it.
const PointcutExpression = "execution(void base.aop.AopDemo.send(..))"

// BeforeSend is the named pointcut. It has no body in the original either —
// an @Pointcut method is a declaration, not code.
func (a *AspectDemo) BeforeSend() springaop.Pointcut {
	return springaop.MustParseExecution(PointcutExpression)
}

// Before is the @Before("beforeSend()") advice.
func (a *AspectDemo) Before() { fmt.Println("send之前") }

// Advisor binds the advice to its pointcut, which is what the weaver builds
// out of the two annotations.
func (a *AspectDemo) Advisor() springaop.Advisor {
	return springaop.Advisor{Pointcut: a.BeforeSend(), Advice: beforeAdvice{aspect: a}}
}

// beforeAdvice adapts a @Before method to the interceptor chain: run the
// advice, then proceed to the target.
type beforeAdvice struct{ aspect *AspectDemo }

func (b beforeAdvice) Invoke(invocation springaop.MethodInvocation) ([]any, error) {
	b.aspect.Before()
	return invocation.Proceed()
}
