// Package proxy holds the JDK dynamic-proxy demonstration.
package proxy

import "fmt"

// UserService is the interface the dynamic proxy is created over.
type UserService interface {
	PrintName()
	PrintAge()
}

// UserServiceImpl is the proxied target.
type UserServiceImpl struct{}

// NewUserServiceImpl returns the target.
func NewUserServiceImpl() *UserServiceImpl { return &UserServiceImpl{} }

// PrintName prints the name and then calls PrintAge on the target itself.
//
// The self-call is the point of the demonstration: it never reaches the
// proxy, so the handler does not report it.
func (s *UserServiceImpl) PrintName() {
	fmt.Println("Name is XXX")
	s.PrintAge()
}

// PrintAge prints the age.
func (s *UserServiceImpl) PrintAge() {
	fmt.Println("Age: 18")
}

var _ UserService = (*UserServiceImpl)(nil)
