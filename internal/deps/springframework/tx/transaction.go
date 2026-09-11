// Package tx reproduces the transaction metadata the application declares:
// org.springframework.transaction.annotation.Propagation and the
// @Transactional attributes that name it.
//
// No transaction manager is configured anywhere in the original — neither XML
// file contains <tx:annotation-driven/> or a DataSource — so the annotations
// are declarative only, and this package reproduces what is observable about
// them: which methods carry them and with what propagation.
package tx

import "fmt"

// Propagation mirrors
// org.springframework.transaction.annotation.Propagation.
type Propagation int

// The propagation constants, in the order the enum declares them, so that
// String and the zero value both line up with Java's.
const (
	Required Propagation = iota
	Supports
	Mandatory
	RequiresNew
	NotSupported
	Never
	Nested
)

// propagationNames render each constant as the enum name Java prints.
var propagationNames = map[Propagation]string{
	Required:     "REQUIRED",
	Supports:     "SUPPORTS",
	Mandatory:    "MANDATORY",
	RequiresNew:  "REQUIRES_NEW",
	NotSupported: "NOT_SUPPORTED",
	Never:        "NEVER",
	Nested:       "NESTED",
}

// String renders the constant as the enum name Java prints.
func (p Propagation) String() string {
	if name, ok := propagationNames[p]; ok {
		return name
	}
	return fmt.Sprintf("Propagation(%d)", int(p))
}

// Attributes are the @Transactional attributes of one method. Only the
// propagation is ever set in this repository; the remaining attributes keep
// their Spring defaults.
type Attributes struct {
	Propagation Propagation
	ReadOnly    bool
}

// Transactional is implemented by a bean whose methods the original annotated
// with @Transactional. Go has no annotations, so the bean declares the
// attributes itself, keyed by method name.
type Transactional interface {
	TransactionalMethods() map[string]Attributes
}

// AttributesFor returns the @Transactional attributes declared for a method,
// the role TransactionAttributeSource plays.
func AttributesFor(bean any, method string) (Attributes, bool) {
	t, ok := bean.(Transactional)
	if !ok {
		return Attributes{}, false
	}
	attrs, ok := t.TransactionalMethods()[method]
	return attrs, ok
}
