package javart

import (
	"fmt"
	"reflect"
	"strings"
)

// Null is how Java renders an absent value once it reaches a string: the four
// characters "null", not the empty string and not Go's "<nil>".
//
// It is named because it is load-bearing and easy to get wrong. It reaches the
// output of four separate demonstrations — SimpleModel.toString, the
// bean-definition listing, the property-descriptor listing, and the web
// application's greeting — and every one of them was compared against the Java
// original byte for byte.
const Null = "null"

// ArraysToString renders a slice the way java.util.Arrays.toString does:
// the elements, joined by ", ", inside square brackets. A nil slice renders
// as "null", matching Arrays.toString(null).
//
// Go's own "%v" joins with a single space, so this is not cosmetic: the
// bean-definition listing and the split demonstration both print through it.
func ArraysToString(v any) string {
	if v == nil {
		return Null
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		panic(fmt.Sprintf("javart: ArraysToString expects a slice or array, got %T", v))
	}
	if rv.Kind() == reflect.Slice && rv.IsNil() {
		return Null
	}
	parts := make([]string, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		parts[i] = ToString(rv.Index(i).Interface())
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

// ToString renders a single value the way Java's String.valueOf would: nil
// becomes the four characters "null", everything else uses its natural
// rendering.
func ToString(v any) string {
	if v == nil {
		return Null
	}
	if rv := reflect.ValueOf(v); rv.Kind() == reflect.Ptr && rv.IsNil() {
		return Null
	}
	switch t := v.(type) {
	case string:
		return t
	case fmt.Stringer:
		return t.String()
	case *int:
		return fmt.Sprintf("%d", *t)
	}
	return fmt.Sprintf("%v", v)
}

// IdentityString renders a slice reference the way Java renders an array
// reference passed to println: a type descriptor and an identity, never the
// contents. The identity differs per run in Java and per run here; only the
// shape is contractual.
func IdentityString(v any) string {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Slice {
		panic(fmt.Sprintf("javart: IdentityString expects a slice, got %T", v))
	}
	return fmt.Sprintf("%s@%x", rv.Type().String(), rv.Pointer())
}
