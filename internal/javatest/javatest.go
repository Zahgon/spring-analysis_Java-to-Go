// Package javatest holds the language-and-standard-library demonstrations.
//
// The original package is named test; Go reserves that word for the testing
// convention (a package called test with _test.go files beside it reads as
// something else entirely), so the directory is javatest, after the class.
package javatest

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/seaswalker/spring-analysis/internal/base"
	"github.com/seaswalker/spring-analysis/internal/deps/javart"
)

// List is the general-purpose list MyList narrows: its element accessor
// returns the unconstrained element type, the role java.util.ArrayList's
// Object-returning get plays.
//
// It carries nothing but that accessor. Go reports a flat method set — every
// method promoted from an embedded type is in it, with no way to ask which
// type declared which — so anything else here would show up in the listing
// below as though MyList had declared it.
type List struct {
	elements []any
}

// Get returns the element at index, or nil when the index is out of range.
func (l *List) Get(index int) any {
	if index < 0 || index >= len(l.elements) {
		return nil
	}
	return l.elements[index]
}

// MyList narrows the element accessor to a string.
//
// In the original this is an inner class extending ArrayList whose get(int) is
// declared to return String; javac emits a bridge method returning Object
// alongside it, so the class reports two declared get methods.
//
// Go has no bridge methods. MyList's own Get shadows the embedded one, and the
// method set reports exactly one Get — the narrowed one. The widened accessor
// is not gone, it is simply reached through the embedded value instead of
// through the method set. The demonstration prints what the language actually
// reports, which is the point of it.
type MyList struct {
	List
}

// NewMyList returns an empty MyList.
func NewMyList() *MyList { return &MyList{} }

// Get returns the empty string, narrowing the embedded accessor's result.
func (m *MyList) Get(index int) string { return "" }

// DeclaredMethods returns MyList's method set, rendered as the original
// rendered its declared methods: the name, and the type the method returns.
func DeclaredMethods() []string {
	t := reflect.TypeOf(&MyList{})
	out := make([]string, 0, t.NumMethod())
	for i := 0; i < t.NumMethod(); i++ {
		m := t.Method(i)
		result := "void"
		if m.Type.NumOut() > 0 {
			result = m.Type.Out(0).String()
		}
		out = append(out, fmt.Sprintf("name: %s, return: %s", m.Name, result))
	}
	return out
}

// SplitInput is the tab-separated record the split demonstration reads.
const SplitInput = "1\t2\taug\tfri\t14.7\t66\t2.7\t0\t0"

// MonthIndex and WeatherIndex are the columns the demonstration labels.
//
// WeatherIndex points at the temperature column, not at a weather word. That
// is what the original reads, and it is reproduced rather than corrected.
const (
	MonthIndex   = 2
	WeatherIndex = 4
)

// Classpath prints the code units the running program was built from, the
// role java.class.path plays on the JVM.
func Classpath() {
	fmt.Println(javart.CodePath())
}

// FindClass prints the URL of every resource matching "base/*".
//
// It prints nothing: the lookup is by literal path, not by pattern, so the
// asterisk matches no resource. Reproducing that empty result is the
// demonstration.
func FindClass(classPath string) {
	for _, url := range javart.NewClassLoader(classPath).GetResources("base/*") {
		fmt.Println(url)
	}
}

// Intro prints Student's property descriptors: each property's read method,
// then its write method, in property-name order.
func Intro() {
	for _, pd := range javart.GetBeanInfo(&base.Student{}) {
		fmt.Println(javart.ToString(nilIfEmpty(pd.ReadMethod)))
		fmt.Println(javart.ToString(nilIfEmpty(pd.WriteMethod)))
	}
}

// nilIfEmpty turns a missing accessor into the nil that prints as "null",
// which is what a property with no setter printed in the original.
func nilIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// Split splits the record on tabs and prints the parts, the slice's identity,
// and the two labelled columns.
func Split() {
	arr := strings.Split(SplitInput, "\t")
	fmt.Println(javart.ArraysToString(arr))
	fmt.Println(javart.IdentityString(arr))
	fmt.Println("月份: " + arr[MonthIndex])
	fmt.Println("天气: " + arr[WeatherIndex])
}

// PrintDeclaredMethods prints MyList's declared methods, which is what the
// class's main method does.
func PrintDeclaredMethods() {
	for _, line := range DeclaredMethods() {
		fmt.Println(line)
	}
}
