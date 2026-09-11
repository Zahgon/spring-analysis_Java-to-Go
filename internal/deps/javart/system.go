package javart

import (
	"os"
	"reflect"
	"runtime/debug"
	"strings"
)

// packagePathSuffix is this package's own path within the module. Trimming it
// from the package's import path yields the module path.
//
// The module path is not reachable any other way that works in every build
// mode: debug.ReadBuildInfo reports an empty Main.Path and no dependencies
// from a test binary under some toolchains, so a program that only asked it
// would report less about itself when run under `go test` than when run
// normally. A type's PkgPath is fixed at compile time and is always there.
const packagePathSuffix = "/internal/deps/javart"

// marker exists only so that its package path can be read.
type marker struct{}

// ModulePath returns the import path of the module the program was built from.
func ModulePath() string {
	return strings.TrimSuffix(reflect.TypeOf(marker{}).PkgPath(), packagePathSuffix)
}

// CodePath returns the list of code units the running program was built from,
// the role java.class.path plays on the JVM.
//
// A Go binary is statically linked, so the equivalent information is the
// executable, the module it was built from, and every dependency the linker
// recorded, joined by the platform's path separator the way a classpath is.
// The value is environment-specific in both languages and is printed, never
// compared.
func CodePath() string {
	parts := []string{}
	if exe, err := os.Executable(); err == nil {
		parts = append(parts, exe)
	}
	parts = append(parts, ModulePath())
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, dep := range info.Deps {
			parts = append(parts, dep.Path+"@"+dep.Version)
		}
	}
	return strings.Join(parts, string(os.PathListSeparator))
}
