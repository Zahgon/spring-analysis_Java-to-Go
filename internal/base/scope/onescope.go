// Package scope holds the application's custom
// org.springframework.beans.factory.config.Scope implementation.
package scope

import (
	"fmt"

	"github.com/seaswalker/spring-analysis/internal/base"
	"github.com/seaswalker/spring-analysis/internal/deps/springframework/beans"
)

// OneScope is a Scope that hands back a new object on every call.
type OneScope struct {
	index int
}

// NewOneScope returns a scope whose counter starts at zero.
func NewOneScope() *OneScope { return &OneScope{} }

// Get returns a fresh Student, ignoring the ObjectFactory the container
// offers.
//
// The counter is read for the name and incremented, then read again for the
// age — so the first call yields "skywalker-0" aged 1, the second
// "skywalker-1" aged 2. That off-by-one between the two fields is the
// original's behaviour, not a slip in the port.
func (s *OneScope) Get(name string, objectFactory beans.ObjectFactory) any {
	fmt.Println("get被调用")
	student := base.NewStudentOf(fmt.Sprintf("skywalker-%d", s.index), s.index+1)
	s.index++
	return student
}

// Remove is not supported by this scope.
func (s *OneScope) Remove(name string) any { return nil }

// RegisterDestructionCallback discards the callback: nothing this scope hands
// out is ever destroyed, so there is nothing to run one on. The original's body
// is empty for the same reason; the discard is written out so that the decision
// to drop the callback is visible rather than implied.
func (s *OneScope) RegisterDestructionCallback(name string, callback func()) {
	_ = callback
}

// ResolveContextualObject is not supported by this scope.
func (s *OneScope) ResolveContextualObject(key string) any { return nil }

// GetConversationID is not supported by this scope.
func (s *OneScope) GetConversationID() string { return "" }

var _ beans.Scope = (*OneScope)(nil)
