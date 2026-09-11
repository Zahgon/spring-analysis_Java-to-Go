package transaction

import (
	"reflect"

	"github.com/seaswalker/spring-analysis/internal/deps/springframework/beans"
)

// The two beans carry @Component in the original, so a component scan would
// find them. Nothing scans base.transaction — config.xml declares no
// <context:component-scan> — but the classes are registered so that the
// container can resolve them by their original names, and so that asking for
// one that was never declared fails the way it does in the original.
func init() {
	beans.Register(beans.ClassDescriptor{
		Name: "base.transaction.NestedBean",
		Type: reflect.TypeOf(&NestedBean{}),
		New:  func() any { return NewNestedBean() },
	})
	beans.Register(beans.ClassDescriptor{
		Name: "base.transaction.TransactionBean",
		Type: reflect.TypeOf(&TransactionBean{}),
		New:  func() any { return NewTransactionBean() },
	})
}
