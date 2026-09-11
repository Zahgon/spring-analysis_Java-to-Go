// Package aop holds the AOP demonstration config.xml wires up: an advised
// bean, its subclass, and the interceptor applied to it.
package aop

import (
	"fmt"

	springaop "github.com/seaswalker/spring-analysis/internal/deps/springframework/aop"
)

// SimpleAopBeanAPI is the method surface the advised bean exposes.
//
// The original has no interface here: Spring falls back to CGLIB and proxies
// the class by subclassing it. Go decides interface satisfaction at compile
// time and cannot subclass, so the advisable surface is named. Every method
// the pointcut "execution(* aop.SimpleAopBean.*(..))" can select is on it.
type SimpleAopBeanAPI interface {
	Boo()
	TestB()
	TestC()
}

// SimpleAopBean is the advised target.
type SimpleAopBean struct{}

// NewSimpleAopBean returns the target.
func NewSimpleAopBean() *SimpleAopBean { return &SimpleAopBean{} }

// Boo prints and then calls TestB on the target itself.
//
// The call does not go through the proxy — it is an ordinary self-invocation —
// so TestB is not advised when it is reached this way.
func (b *SimpleAopBean) Boo() {
	fmt.Println("testA执行")
	b.TestB()
}

// TestB prints, then re-enters through the current proxy to reach TestC, so
// that TestC is advised even though the call starts inside the target. That
// re-entry is what <aop:config expose-proxy="true"> exists for.
//
// Recovering the proxy fails when there is no advised invocation in progress.
// The original raises an unchecked IllegalStateException there, ending the
// program; a panic is the same contract in Go.
func (b *SimpleAopBean) TestB() {
	fmt.Println("testB执行")
	proxy, err := springaop.CurrentProxy()
	if err != nil {
		panic(err)
	}
	proxy.(SimpleAopBeanAPI).TestC()
}

// TestC prints the innermost message.
func (b *SimpleAopBean) TestC() {
	fmt.Println("testC执行")
}

// SimpleChildAopBean overrides TestC.
//
// The original extends SimpleAopBean; Go embeds it, which promotes Boo and
// TestB unchanged and lets TestC shadow the embedded one. Dispatch stays
// virtual where it matters: TestB reaches TestC through the proxy, which
// invokes it on the actual target, so a child instance prints the child's
// message.
type SimpleChildAopBean struct {
	SimpleAopBean
}

// NewSimpleChildAopBean returns the subclass.
func NewSimpleChildAopBean() *SimpleChildAopBean { return &SimpleChildAopBean{} }

// TestC prints the child's message.
func (b *SimpleChildAopBean) TestC() {
	fmt.Println("child testC")
}

var (
	_ SimpleAopBeanAPI = (*SimpleAopBean)(nil)
	_ SimpleAopBeanAPI = (*SimpleChildAopBean)(nil)
)
