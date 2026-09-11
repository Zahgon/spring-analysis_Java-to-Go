// Package javart reproduces the behaviour the original Java application took
// from the JDK standard library. Go's standard library covers most of it, but
// a handful of behaviours are observable in the program's output and have no
// Go equivalent with the same semantics:
//
//   - java.util.Arrays.toString rendering ("[a, b, c]", not Go's "[a b c]")
//   - .properties parsing, including \uXXXX escapes and its whitespace rules
//   - java.util.ResourceBundle locale fallback
//   - java.beans.Introspector property discovery over a getter/setter surface
//   - java.lang.reflect.Proxy dispatch through an InvocationHandler
//   - java.lang.ThreadLocal, which Spring's AopContext is built on
//   - java.util.Date.toString formatting
//
// Nothing here is a general-purpose reimplementation of the JDK; each file
// reproduces exactly the behaviour the demonstrations depend on.
package javart
