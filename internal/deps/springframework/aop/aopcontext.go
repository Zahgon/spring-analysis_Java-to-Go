package aop

import "errors"

// ErrNoProxyExposed is what AopContext.currentProxy() raises when the current
// call is not running inside an advised invocation, or when the proxy was
// created without expose-proxy. Spring throws IllegalStateException with the
// message reproduced here.
var ErrNoProxyExposed = errors.New(
	"Cannot find current proxy: Set 'exposeProxy' property on Advised to 'true' to make it available, " +
		"and ensure that AopContext.currentProxy() is invoked in the same thread as the AOP invocation context.")

// CurrentProxy returns the proxy handling the invocation the calling goroutine
// is inside, mirroring org.springframework.aop.framework.AopContext.
//
// This is what makes an advised method's call to itself go back through the
// advice instead of straight to the target.
func CurrentProxy() (any, error) {
	p, ok := currentProxy.Get()
	if !ok {
		return nil, ErrNoProxyExposed
	}
	return p, nil
}
