package aop

import (
	"fmt"
	"strings"
)

// Pointcut decides whether advice applies to a method.
type Pointcut interface {
	Matches(m Method) bool
	// MatchesClass reports whether the pointcut could select any method of
	// the named class. An auto-proxy creator asks this first, to decide
	// whether the bean needs a proxy at all.
	MatchesClass(className string) bool
	// Expression returns the pointcut as it was written.
	Expression() string
}

// ExecutionPointcut matches an AspectJ execution() designator of the shape the
// application uses:
//
//	execution(<return> <class>.<method>(..))
//
// where <return> and <method> may be "*". That is the whole grammar the two
// pointcuts in this repository need — config.xml's
// "execution(* aop.SimpleAopBean.*(..))" and AspectDemo's
// "execution(void base.aop.AopDemo.send(..))". Anything richer is rejected at
// parse time rather than silently matching nothing.
type ExecutionPointcut struct {
	expression    string
	returnType    string
	declaringSpec string
	methodSpec    string
}

// ParseExecution parses an execution() pointcut expression.
func ParseExecution(expression string) (*ExecutionPointcut, error) {
	expr := strings.TrimSpace(expression)
	if !strings.HasPrefix(expr, "execution(") || !strings.HasSuffix(expr, ")") {
		return nil, fmt.Errorf("aop: unsupported pointcut %q: only execution(...) is supported", expression)
	}
	body := strings.TrimSpace(expr[len("execution(") : len(expr)-1])

	parts := strings.Fields(body)
	if len(parts) != 2 {
		return nil, fmt.Errorf("aop: unsupported pointcut %q: expected '<return> <class>.<method>(..)'", expression)
	}
	returnType, signature := parts[0], parts[1]

	open := strings.Index(signature, "(")
	if open < 0 || !strings.HasSuffix(signature, ")") {
		return nil, fmt.Errorf("aop: unsupported pointcut %q: missing parameter list", expression)
	}
	params := signature[open+1 : len(signature)-1]
	if params != ".." {
		return nil, fmt.Errorf("aop: unsupported pointcut %q: only the '(..)' parameter pattern is supported", expression)
	}

	qualified := signature[:open]
	dot := strings.LastIndex(qualified, ".")
	if dot < 0 {
		return nil, fmt.Errorf("aop: unsupported pointcut %q: expected a qualified method name", expression)
	}
	return &ExecutionPointcut{
		expression:    expression,
		returnType:    returnType,
		declaringSpec: qualified[:dot],
		methodSpec:    qualified[dot+1:],
	}, nil
}

// MustParseExecution is ParseExecution for expressions fixed at compile time.
func MustParseExecution(expression string) *ExecutionPointcut {
	p, err := ParseExecution(expression)
	if err != nil {
		panic(err)
	}
	return p
}

// Expression returns the pointcut as written.
func (p *ExecutionPointcut) Expression() string { return p.expression }

// Matches reports whether the pointcut selects m.
func (p *ExecutionPointcut) Matches(m Method) bool {
	return p.MatchesClass(m.DeclaringClass) && matchSpec(p.methodSpec, m.Name)
}

// MatchesClass reports whether the pointcut's declaring-type pattern selects
// the named class.
func (p *ExecutionPointcut) MatchesClass(className string) bool {
	return matchSpec(p.declaringSpec, className)
}

// matchSpec compares an AspectJ name pattern against a name. "*" matches
// anything; a pattern may also use "*" as a prefix or suffix wildcard.
func matchSpec(spec, name string) bool {
	switch {
	case spec == "*":
		return true
	case strings.HasPrefix(spec, "*") && strings.HasSuffix(spec, "*") && len(spec) > 1:
		return strings.Contains(name, strings.Trim(spec, "*"))
	case strings.HasPrefix(spec, "*"):
		return strings.HasSuffix(name, spec[1:])
	case strings.HasSuffix(spec, "*"):
		return strings.HasPrefix(name, spec[:len(spec)-1])
	default:
		return spec == name
	}
}
