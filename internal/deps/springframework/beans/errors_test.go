package beans_test

import (
	"errors"
	"testing"

	"github.com/seaswalker/spring-analysis/internal/deps/springframework/aop"
	"github.com/seaswalker/spring-analysis/internal/deps/springframework/beans"
)

// TestErrorMessages keeps the wording the original raised. The
// NoSuchBeanDefinition message is printed by a demonstration, so it is
// contract rather than diagnostics.
func TestErrorMessages(t *testing.T) {
	cases := []struct {
		err  error
		want string
	}{
		{
			&beans.NoSuchBeanDefinitionError{TypeName: "base.transaction.TransactionBean"},
			"No qualifying bean of type 'base.transaction.TransactionBean' available",
		},
		{
			&beans.NoSuchBeanDefinitionError{BeanName: "simpleBean"},
			"No bean named 'simpleBean' available",
		},
		{
			&beans.NoUniqueBeanDefinitionError{TypeName: "test.Widget", Names: []string{"a", "b"}},
			"No qualifying bean of type 'test.Widget' available: expected single matching bean but found 2: [a b]",
		},
		{
			&beans.BeanCreationError{BeanName: "w", Err: errors.New("boom")},
			"Error creating bean with name 'w': boom",
		},
	}
	for _, tc := range cases {
		if got := tc.err.Error(); got != tc.want {
			t.Errorf("message = %q, want %q", got, tc.want)
		}
	}
}

// TestBeanCreationErrorUnwraps lets callers reach the cause.
func TestBeanCreationErrorUnwraps(t *testing.T) {
	cause := errors.New("boom")
	err := &beans.BeanCreationError{BeanName: "w", Err: cause}
	if !errors.Is(err, cause) {
		t.Error("BeanCreationError does not unwrap to its cause")
	}
}

// TestCircularReferenceIsReported covers the guard against a definition that
// depends on itself.
func TestCircularReferenceIsReported(t *testing.T) {
	ctx := beans.NewApplicationContext()
	ctx.RegisterBeanDefinition(&beans.BeanDefinition{
		Name:   "loop",
		Create: func(f beans.BeanFactory) (any, error) { return f.GetBean("loop") },
	})
	_, err := ctx.GetBean("loop")
	if err == nil {
		t.Fatal("a self-referential definition resolved")
	}
	if got := err.Error(); got == "" || !contains(got, "currently in creation") {
		t.Errorf("message = %q, want it to report the cycle", got)
	}
}

// TestUnknownAdvisorIsReported covers a definition naming an advisor the
// context never registered.
func TestUnknownAdvisorIsReported(t *testing.T) {
	ctx := beans.NewApplicationContext()
	ctx.RegisterBeanDefinition(&beans.BeanDefinition{
		Name:         "w",
		ClassName:    "test.Widget",
		Create:       func(beans.BeanFactory) (any, error) { return &widget{}, nil },
		AdvisorNames: []string{"missing"},
	})
	if _, err := ctx.GetBean("w"); err == nil {
		t.Fatal("a bean naming an unregistered advisor resolved")
	}
}

// TestAdvisedClassWithoutFacadeIsReported covers the other creation failure:
// a class that a pointcut selects but that registers no proxy façade cannot be
// advised, and saying so beats handing back an unadvised bean.
func TestAdvisedClassWithoutFacadeIsReported(t *testing.T) {
	ctx := beans.NewApplicationContext()
	ctx.RegisterAdvisor("noop", aop.Advisor{
		Pointcut: aop.MustParseExecution("execution(* test.Widget.*(..))"),
		Advice:   passThrough{},
	})
	ctx.RegisterBeanDefinition(&beans.BeanDefinition{
		Name:         "w",
		ClassName:    "test.Widget",
		Create:       func(beans.BeanFactory) (any, error) { return &widget{}, nil },
		AdvisorNames: []string{"noop"},
	})
	_, err := ctx.GetBean("w")
	if err == nil {
		t.Fatal("a class with no façade was advised anyway")
	}
	if !contains(err.Error(), "registers no proxy façade") {
		t.Errorf("message = %q, want it to name the missing façade", err)
	}
}

type passThrough struct{}

func (passThrough) Invoke(inv aop.MethodInvocation) ([]any, error) { return inv.Proceed() }

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
