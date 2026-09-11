package javart

import "strings"

// GoMethodName maps a method name as the original declared it onto the Go
// method that replaces it: Java's camelCase convention makes a method
// exported in Go by capitalising its first letter, and nothing else changes.
//
// The distinction matters wherever a name is both dispatched on and printed.
// An interceptor reports the name the original declared — "testB" — while the
// call it forwards has to reach the Go method "TestB".
func GoMethodName(name string) string {
	if name == "" {
		return name
	}
	return strings.ToUpper(name[:1]) + name[1:]
}

// JavaMethodName is the inverse: the name the original declared for a Go
// method.
func JavaMethodName(name string) string {
	if name == "" {
		return name
	}
	return strings.ToLower(name[:1]) + name[1:]
}
