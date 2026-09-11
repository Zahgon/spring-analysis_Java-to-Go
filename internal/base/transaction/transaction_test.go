package transaction_test

import (
	"strings"
	"testing"

	"github.com/seaswalker/spring-analysis/internal/base/transaction"
	"github.com/seaswalker/spring-analysis/internal/deps/springframework/tx"
	"github.com/seaswalker/spring-analysis/internal/testsupport"
)

// TestProcessCallsTheNestedBean covers what the transaction demonstration
// would print if config.xml declared the beans.
func TestProcessCallsTheNestedBean(t *testing.T) {
	bean := transaction.NewTransactionBean()
	nested := transaction.NewNestedBean()
	bean.SetNestedBean(nested)
	if bean.GetNestedBean() != nested {
		t.Error("SetNestedBean did not take")
	}

	lines := testsupport.Lines(testsupport.CaptureStdout(t, bean.Process))
	want := []string{"事务执行", "嵌套事务"}
	if strings.Join(lines, "\n") != strings.Join(want, "\n") {
		t.Errorf("got %v, want %v", lines, want)
	}
}

// TestDeclaredPropagation keeps the two @Transactional declarations.
func TestDeclaredPropagation(t *testing.T) {
	outer, ok := tx.AttributesFor(transaction.NewTransactionBean(), "Process")
	if !ok || outer.Propagation != tx.Required {
		t.Errorf("Process propagation = %v (found %v), want REQUIRED", outer.Propagation, ok)
	}
	inner, ok := tx.AttributesFor(transaction.NewNestedBean(), "Nest")
	if !ok || inner.Propagation != tx.Nested {
		t.Errorf("Nest propagation = %v (found %v), want NESTED", inner.Propagation, ok)
	}
	if _, ok := tx.AttributesFor(transaction.NewNestedBean(), "Absent"); ok {
		t.Error("an undeclared method reported attributes")
	}
	if _, ok := tx.AttributesFor("not a bean", "Nest"); ok {
		t.Error("a non-transactional value reported attributes")
	}
}

// TestPropagationNames keeps the enum names the constants print as.
func TestPropagationNames(t *testing.T) {
	want := map[tx.Propagation]string{
		tx.Required:     "REQUIRED",
		tx.Supports:     "SUPPORTS",
		tx.Mandatory:    "MANDATORY",
		tx.RequiresNew:  "REQUIRES_NEW",
		tx.NotSupported: "NOT_SUPPORTED",
		tx.Never:        "NEVER",
		tx.Nested:       "NESTED",
	}
	for p, name := range want {
		if got := p.String(); got != name {
			t.Errorf("%d.String() = %q, want %q", int(p), got, name)
		}
	}
	if got := tx.Propagation(99).String(); !strings.Contains(got, "99") {
		t.Errorf("an unknown constant printed %q", got)
	}
}
