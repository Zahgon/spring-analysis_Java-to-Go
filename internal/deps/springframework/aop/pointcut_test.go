package aop_test

import (
	"strings"
	"testing"

	"github.com/seaswalker/spring-analysis/internal/deps/springframework/aop"
)

// TestExecutionPointcutMatching covers the two expressions the repository
// ships, and the wildcard rules they use.
func TestExecutionPointcutMatching(t *testing.T) {
	all := aop.MustParseExecution("execution(* aop.SimpleAopBean.*(..))")
	send := aop.MustParseExecution("execution(void base.aop.AopDemo.send(..))")

	cases := []struct {
		pointcut aop.Pointcut
		method   aop.Method
		want     bool
	}{
		{all, aop.Method{Name: "testB", DeclaringClass: "aop.SimpleAopBean"}, true},
		{all, aop.Method{Name: "boo", DeclaringClass: "aop.SimpleAopBean"}, true},
		{all, aop.Method{Name: "testB", DeclaringClass: "aop.Other"}, false},
		{send, aop.Method{Name: "send", DeclaringClass: "base.aop.AopDemo"}, true},
		{send, aop.Method{Name: "receive", DeclaringClass: "base.aop.AopDemo"}, false},
		{send, aop.Method{Name: "send", DeclaringClass: "base.aop.Other"}, false},
	}
	for _, tc := range cases {
		if got := tc.pointcut.Matches(tc.method); got != tc.want {
			t.Errorf("%s.Matches(%s.%s) = %v, want %v",
				tc.pointcut.Expression(), tc.method.DeclaringClass, tc.method.Name, got, tc.want)
		}
	}

	if !all.MatchesClass("aop.SimpleAopBean") {
		t.Error("MatchesClass rejected the declaring class")
	}
	if send.MatchesClass("base.aop.Other") {
		t.Error("MatchesClass accepted an unrelated class")
	}
}

// TestExecutionPointcutRejectsUnsupportedExpressions checks that an
// expression outside the supported grammar fails loudly instead of silently
// matching nothing.
func TestExecutionPointcutRejectsUnsupportedExpressions(t *testing.T) {
	for _, expr := range []string{
		"within(aop..*)",
		"execution(* aop.SimpleAopBean.*(String))",
		"execution(*)",
		"execution(* noPackage(..))",
		"execution(* aop.SimpleAopBean.*",
	} {
		if _, err := aop.ParseExecution(expr); err == nil {
			t.Errorf("ParseExecution(%q) succeeded, want an error", expr)
		}
	}
}

// TestNamePatternWildcards covers the prefix, suffix and contains forms.
func TestNamePatternWildcards(t *testing.T) {
	cases := []struct {
		expr   string
		method string
		want   bool
	}{
		{"execution(* p.C.test*(..))", "testB", true},
		{"execution(* p.C.test*(..))", "boo", false},
		{"execution(* p.C.*B(..))", "testB", true},
		{"execution(* p.C.*B(..))", "testC", false},
		{"execution(* p.C.*est*(..))", "testB", true},
	}
	for _, tc := range cases {
		p := aop.MustParseExecution(tc.expr)
		if got := p.Matches(aop.Method{Name: tc.method, DeclaringClass: "p.C"}); got != tc.want {
			t.Errorf("%s vs %q = %v, want %v", tc.expr, tc.method, got, tc.want)
		}
	}
}

// TestMustParseExecutionPanicsOnBadInput covers the compile-time-constant
// helper.
func TestMustParseExecutionPanicsOnBadInput(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustParseExecution did not panic")
		} else if !strings.Contains(strings.ToLower(err(r)), "unsupported") {
			t.Errorf("panic value = %v", r)
		}
	}()
	aop.MustParseExecution("within(*)")
}

func err(v any) string {
	if e, ok := v.(error); ok {
		return e.Error()
	}
	return ""
}
