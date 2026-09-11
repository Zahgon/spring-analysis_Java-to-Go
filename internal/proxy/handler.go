package proxy

import (
	"fmt"

	"github.com/seaswalker/spring-analysis/internal/deps/javart"
)

// Handler reports every call that reaches the proxy and then forwards it to
// the target.
type Handler struct {
	target any
}

// NewHandler binds a handler to a target.
func NewHandler(target any) *Handler { return &Handler{target: target} }

// Invoke prints the method's name and delegates to the target.
func (h *Handler) Invoke(proxy any, method javart.Method, args []any) ([]any, error) {
	name := method.GetName()
	fmt.Println("Method " + name + " is proxyed.")
	return method.Invoke(h.target, args...)
}

var _ javart.InvocationHandler = (*Handler)(nil)

// userServiceProxy is the type java.lang.reflect.Proxy would have generated.
//
// The JVM synthesises it from the interface at run time; Go decides interface
// satisfaction at compile time, so it is written out. Its behaviour is the
// generated one: each interface method funnels through the
// InvocationHandler, and nothing else does.
type userServiceProxy struct {
	dispatcher *javart.InvocationDispatcher
}

// PrintName routes PrintName through the handler.
func (p *userServiceProxy) PrintName() { p.dispatch("printName") }

// PrintAge routes PrintAge through the handler.
func (p *userServiceProxy) PrintAge() { p.dispatch("printAge") }

// dispatch forwards one call, naming the method as the original declared it,
// because that is the name the handler prints. A dispatch failure means the
// proxy and the target disagree about the interface, which is a programming
// error.
func (p *userServiceProxy) dispatch(method string) {
	if _, err := p.dispatcher.Dispatch(p, method); err != nil {
		panic(err)
	}
}

// NewUserServiceProxy is Proxy.newProxyInstance for the UserService
// interface: it returns an object that satisfies UserService and routes every
// call through handler.
func NewUserServiceProxy(target UserService, handler javart.InvocationHandler) UserService {
	return &userServiceProxy{dispatcher: javart.NewInvocationDispatcher(target, handler)}
}

var _ UserService = (*userServiceProxy)(nil)
