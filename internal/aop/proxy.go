package aop

import (
	"reflect"

	springaop "github.com/seaswalker/spring-analysis/internal/deps/springframework/aop"
	"github.com/seaswalker/spring-analysis/internal/deps/springframework/beans"
)

// simpleAopBeanProxy is the object callers hold once the bean is advised.
//
// In the original this type does not exist in source: CGLIB generates a
// subclass of SimpleAopBean at run time and every inherited method is
// overridden to enter the interceptor chain. Go decides interface
// satisfaction at compile time, so the generated type is written out. It is
// mechanical — one method per advisable method, each handing its name to the
// invoker — and it holds no logic of its own.
type simpleAopBeanProxy struct {
	proxy *springaop.Proxy
}

// Boo enters the interceptor chain for Boo.
func (p *simpleAopBeanProxy) Boo() { p.dispatch("boo") }

// TestB enters the interceptor chain for TestB.
func (p *simpleAopBeanProxy) TestB() { p.dispatch("testB") }

// TestC enters the interceptor chain for TestC.
func (p *simpleAopBeanProxy) TestC() { p.dispatch("testC") }

// dispatch runs one advised call, naming the method as the original declared
// it — that name reaches the interceptor and the pointcut. A failure here is a
// container defect, not a condition the demonstration can handle, which is why
// it panics rather than widening every method signature with an error the
// original does not have.
func (p *simpleAopBeanProxy) dispatch(method string) {
	if _, err := p.proxy.Invoke(method); err != nil {
		panic(err)
	}
}

// ProxyClassName returns the name the generated proxy type carries.
func (p *simpleAopBeanProxy) ProxyClassName() string { return p.proxy.ClassName() }

// ProxyTarget returns the advised object.
func (p *simpleAopBeanProxy) ProxyTarget() any { return p.proxy.Target() }

var (
	_ SimpleAopBeanAPI  = (*simpleAopBeanProxy)(nil)
	_ springaop.Proxied = (*simpleAopBeanProxy)(nil)
)

func init() {
	beans.Register(beans.ClassDescriptor{
		Name: "aop.SimpleAopBean",
		Type: reflect.TypeOf(&SimpleAopBean{}),
		New:  func() any { return NewSimpleAopBean() },
		Facade: func(proxy *springaop.Proxy) any {
			return &simpleAopBeanProxy{proxy: proxy}
		},
	})
	beans.Register(beans.ClassDescriptor{
		Name: "aop.SimpleChildAopBean",
		Type: reflect.TypeOf(&SimpleChildAopBean{}),
		New:  func() any { return NewSimpleChildAopBean() },
		Facade: func(proxy *springaop.Proxy) any {
			return &simpleAopBeanProxy{proxy: proxy}
		},
	})
	beans.Register(beans.ClassDescriptor{
		Name: "aop.SimpleMethodInterceptor",
		Type: reflect.TypeOf(&SimpleMethodInterceptor{}),
		New:  func() any { return NewSimpleMethodInterceptor() },
	})
}
