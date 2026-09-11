// Package aop reproduces the Spring AOP behaviour the application depends on:
// the aopalliance MethodInterceptor/MethodInvocation chain, an
// execution(...) pointcut, proxy creation, and the expose-proxy contract that
// makes AopContext.currentProxy() work from inside a target method.
package aop

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/seaswalker/spring-analysis/internal/deps/javart"
)

// Method identifies the method being invoked, as org.aopalliance's
// MethodInvocation.getMethod() does.
type Method struct {
	// Name is the method's name as the original declared it — unqualified,
	// and in the source language's camelCase, because interceptors print it
	// and pointcuts match on it.
	Name string
	// DeclaringClass is the fully-qualified name the class had in the
	// original, e.g. "aop.SimpleAopBean". Pointcuts match against it.
	DeclaringClass string
}

// MethodInvocation mirrors org.aopalliance.intercept.MethodInvocation.
type MethodInvocation interface {
	GetMethod() Method
	GetArguments() []any
	GetThis() any
	Proceed() ([]any, error)
}

// MethodInterceptor mirrors org.aopalliance.intercept.MethodInterceptor.
type MethodInterceptor interface {
	Invoke(invocation MethodInvocation) ([]any, error)
}

// Advisor pairs a pointcut with the advice it applies.
type Advisor struct {
	Pointcut Pointcut
	Advice   MethodInterceptor
}

// Proxy is the advised object: it owns the target, the advisor chain, and the
// expose-proxy flag, and runs an invocation through them.
//
// In the original this object is a CGLIB-generated subclass of the target.
// Go cannot generate a type at run time, so the generated part — the typed
// façade that callers hold — is written out beside each advised class and
// delegates here. Everything observable happens in this file.
type Proxy struct {
	target       any
	targetClass  string
	advisors     []Advisor
	exposeProxy  bool
	facade       any
	proxyClasses []string
}

// currentProxy carries the proxy for the goroutine running an advised call,
// reproducing the ThreadLocal that AopContext uses.
var currentProxy = javart.NewThreadLocal[any]()

// NewProxy builds a proxy over target. targetClass is the name the original
// class had, so pointcuts written against Java names keep working.
func NewProxy(target any, targetClass string, advisors []Advisor, exposeProxy bool) *Proxy {
	return &Proxy{target: target, targetClass: targetClass, advisors: advisors, exposeProxy: exposeProxy}
}

// SetFacade records the typed object callers hold. It is what
// AopContext.currentProxy() hands back, so a target that re-enters through the
// proxy gets a value it can use at its own interface type.
func (p *Proxy) SetFacade(facade any) { p.facade = facade }

// Target returns the advised object.
func (p *Proxy) Target() any { return p.target }

// TargetClass returns the original class name of the advised object.
func (p *Proxy) TargetClass() string { return p.targetClass }

// ClassName names the proxy's runtime type the way the original's
// CGLIB-generated subclass named itself: the target's simple name, a marker
// identifying the generator, and a discriminator. The discriminator is derived
// from the advisor set rather than from a JVM class hash, so it is stable
// across runs; the original's was not.
func (p *Proxy) ClassName() string {
	simple := p.targetClass
	if i := strings.LastIndex(simple, "."); i >= 0 {
		simple = simple[i+1:]
	}
	return fmt.Sprintf("%s$$EnhancerBySpringGoAOP$$%08x", simple, p.discriminator())
}

// discriminator hashes the advised method surface, so two proxies over
// different advice are named differently.
func (p *Proxy) discriminator() uint32 {
	const (
		offset32 = 2166136261
		prime32  = 16777619
	)
	h := uint32(offset32)
	write := func(s string) {
		for i := 0; i < len(s); i++ {
			h ^= uint32(s[i])
			h *= prime32
		}
	}
	write(p.targetClass)
	for _, a := range p.advisors {
		write(a.Pointcut.Expression())
		write(reflect.TypeOf(a.Advice).String())
	}
	return h
}

// Invoke runs the named method through every matching advisor and then the
// target itself. name is the method as the original declared it.
func (p *Proxy) Invoke(name string, args ...any) ([]any, error) {
	method := Method{Name: name, DeclaringClass: p.targetClass}

	chain := make([]MethodInterceptor, 0, len(p.advisors))
	for _, a := range p.advisors {
		if a.Pointcut.Matches(method) {
			chain = append(chain, a.Advice)
		}
	}

	if p.exposeProxy {
		exposed := p.facade
		if exposed == nil {
			exposed = p
		}
		old, had := currentProxy.Set(exposed)
		defer func() {
			if had {
				currentProxy.Set(old)
			} else {
				currentProxy.Remove()
			}
		}()
	}

	inv := &reflectiveInvocation{proxy: p, method: method, args: args, chain: chain}
	return inv.Proceed()
}

// reflectiveInvocation is Spring's ReflectiveMethodInvocation: it walks the
// interceptor chain, and calls the target once the chain is exhausted.
type reflectiveInvocation struct {
	proxy  *Proxy
	method Method
	args   []any
	chain  []MethodInterceptor
	index  int
}

func (i *reflectiveInvocation) GetMethod() Method   { return i.method }
func (i *reflectiveInvocation) GetArguments() []any { return i.args }
func (i *reflectiveInvocation) GetThis() any        { return i.proxy.target }

// Proceed advances to the next interceptor, or invokes the target.
func (i *reflectiveInvocation) Proceed() ([]any, error) {
	if i.index < len(i.chain) {
		interceptor := i.chain[i.index]
		i.index++
		return interceptor.Invoke(i)
	}
	return invokeTarget(i.proxy.target, javart.GoMethodName(i.method.Name), i.args)
}

func invokeTarget(target any, name string, args []any) ([]any, error) {
	rv := reflect.ValueOf(target)
	fn := rv.MethodByName(name)
	if !fn.IsValid() {
		return nil, fmt.Errorf("aop: target %s has no method %s", rv.Type(), name)
	}
	in := make([]reflect.Value, len(args))
	for k, a := range args {
		if a == nil {
			in[k] = reflect.Zero(fn.Type().In(k))
			continue
		}
		in[k] = reflect.ValueOf(a)
	}
	out := fn.Call(in)
	results := make([]any, len(out))
	for k, o := range out {
		results[k] = o.Interface()
	}
	return results, nil
}

// Proxied is implemented by the typed façade in front of a Proxy. It is how
// code that would have called getClass() on a CGLIB-generated object gets the
// same two answers: the generated type's name, and the object behind it.
type Proxied interface {
	ProxyClassName() string
	ProxyTarget() any
}

// Unwrap returns the target behind a façade, or the value itself when it is
// not proxied — the role AopProxyUtils.getSingletonTarget plays.
func Unwrap(v any) any {
	if p, ok := v.(Proxied); ok {
		return p.ProxyTarget()
	}
	return v
}
