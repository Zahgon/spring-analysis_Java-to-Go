// Command javatest is the reflection demonstration's entry point, ported from
// test.JavaTest's main method: it prints MyList's declared methods.
package main

import "github.com/seaswalker/spring-analysis/internal/javatest"

func main() { run() }

func run() {
	javatest.PrintDeclaredMethods()
}
