package beans

// ObjectFactory produces a fresh instance of a bean, the callback a Scope
// implementation uses when it decides it needs one.
type ObjectFactory interface {
	GetObject() (any, error)
}

// ObjectFactoryFunc adapts a function to ObjectFactory.
type ObjectFactoryFunc func() (any, error)

// GetObject calls f.
func (f ObjectFactoryFunc) GetObject() (any, error) { return f() }

// Scope mirrors org.springframework.beans.factory.config.Scope.
//
// Returning nil from Remove, ResolveContextualObject or GetConversationID is
// legitimate and means "not supported by this scope" — Spring's own
// implementations do it.
type Scope interface {
	Get(name string, objectFactory ObjectFactory) any
	Remove(name string) any
	RegisterDestructionCallback(name string, callback func())
	ResolveContextualObject(key string) any
	GetConversationID() string
}

// The scope names Spring defines for a plain (non-web) container.
const (
	ScopeSingleton = "singleton"
	ScopePrototype = "prototype"
)
