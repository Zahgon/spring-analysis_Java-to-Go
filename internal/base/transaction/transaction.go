// Package transaction holds the @Transactional demonstration beans.
//
// Neither XML file configures a transaction manager, so nothing here ever
// starts a transaction; what the beans demonstrate is the declaration itself
// and the call from an outer propagation to a nested one.
package transaction

import (
	"fmt"

	"github.com/seaswalker/spring-analysis/internal/deps/springframework/tx"
)

// NestedBean is the @Component whose method declares NESTED propagation.
type NestedBean struct{}

// NewNestedBean returns the bean.
func NewNestedBean() *NestedBean { return &NestedBean{} }

// Nest prints the nested-transaction message.
func (b *NestedBean) Nest() { fmt.Println("嵌套事务") }

// TransactionalMethods declares Nest as @Transactional(propagation = NESTED).
func (b *NestedBean) TransactionalMethods() map[string]tx.Attributes {
	return map[string]tx.Attributes{"Nest": {Propagation: tx.Nested}}
}

// TransactionBean is the @Component whose method declares REQUIRED
// propagation and calls into the nested one.
type TransactionBean struct {
	nestedBean *NestedBean
}

// NewTransactionBean returns the bean.
func NewTransactionBean() *TransactionBean { return &TransactionBean{} }

// GetNestedBean returns the wired nested bean.
func (b *TransactionBean) GetNestedBean() *NestedBean { return b.nestedBean }

// SetNestedBean wires the nested bean.
func (b *TransactionBean) SetNestedBean(nestedBean *NestedBean) { b.nestedBean = nestedBean }

// Process prints the transaction message and calls the nested bean.
func (b *TransactionBean) Process() {
	fmt.Println("事务执行")
	b.nestedBean.Nest()
}

// TransactionalMethods declares Process as
// @Transactional(propagation = REQUIRED).
func (b *TransactionBean) TransactionalMethods() map[string]tx.Attributes {
	return map[string]tx.Attributes{"Process": {Propagation: tx.Required}}
}

var (
	_ tx.Transactional = (*NestedBean)(nil)
	_ tx.Transactional = (*TransactionBean)(nil)
)
