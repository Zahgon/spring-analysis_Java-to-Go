package javart

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// PropertyDescriptor mirrors java.beans.PropertyDescriptor: a property name
// with the accessor methods that back it. A property with no setter has an
// empty WriteMethod, which prints as "null".
type PropertyDescriptor struct {
	Name        string
	ReadMethod  string
	WriteMethod string
}

// GetBeanInfo reproduces java.beans.Introspector.getBeanInfo for the
// getter/setter surface: it pairs GetX/SetX (and IsX for booleans) into
// properties, includes methods promoted from embedded types the way Java
// includes inherited ones, and returns them sorted by property name.
//
// Java always reports the read-only "class" property that Object.getClass
// contributes; the equivalent here is the type's own identity, reported under
// the same name so the demonstration's output keeps its shape.
func GetBeanInfo(v any) []PropertyDescriptor {
	t := reflect.TypeOf(v)
	if t == nil {
		return nil
	}
	base := t
	for base.Kind() == reflect.Ptr {
		base = base.Elem()
	}

	reads := map[string]string{}
	writes := map[string]string{}
	for _, mt := range []reflect.Type{t, reflect.PtrTo(base)} {
		for i := 0; i < mt.NumMethod(); i++ {
			m := mt.Method(i)
			switch {
			case strings.HasPrefix(m.Name, "Get") && m.Type.NumIn() == 1 && m.Type.NumOut() == 1:
				reads[decapitalise(m.Name[3:])] = methodSignature(base, m, m.Type.Out(0))
			case strings.HasPrefix(m.Name, "Is") && m.Type.NumIn() == 1 && m.Type.NumOut() == 1 && m.Type.Out(0).Kind() == reflect.Bool:
				reads[decapitalise(m.Name[2:])] = methodSignature(base, m, m.Type.Out(0))
			case strings.HasPrefix(m.Name, "Set") && m.Type.NumIn() == 2 && m.Type.NumOut() == 0:
				writes[decapitalise(m.Name[3:])] = methodSignature(base, m, nil)
			}
		}
	}
	// The "class" property: read-only, contributed by the root of the type
	// hierarchy rather than by the bean itself.
	reads["class"] = "func (any) Type() reflect.Type"

	names := make([]string, 0, len(reads)+len(writes))
	seen := map[string]bool{}
	for n := range reads {
		if !seen[n] {
			names, seen[n] = append(names, n), true
		}
	}
	for n := range writes {
		if !seen[n] {
			names, seen[n] = append(names, n), true
		}
	}
	sort.Strings(names)

	out := make([]PropertyDescriptor, 0, len(names))
	for _, n := range names {
		out = append(out, PropertyDescriptor{Name: n, ReadMethod: reads[n], WriteMethod: writes[n]})
	}
	return out
}

// methodSignature renders a method the way Java's Method.toString does for the
// original's output: receiver, name, parameters and result.
func methodSignature(base reflect.Type, m reflect.Method, result reflect.Type) string {
	params := make([]string, 0, m.Type.NumIn()-1)
	for i := 1; i < m.Type.NumIn(); i++ {
		params = append(params, m.Type.In(i).String())
	}
	sig := fmt.Sprintf("func (%s.%s) %s(%s)", base.PkgPath(), base.Name(), m.Name, strings.Join(params, ", "))
	if result != nil {
		sig += " " + result.String()
	}
	return sig
}

func decapitalise(s string) string {
	if s == "" {
		return s
	}
	// Java's Introspector.decapitalize leaves a run of leading capitals alone
	// ("URL" stays "URL"), and lowercases a single leading capital.
	if len(s) > 1 && s[0] >= 'A' && s[0] <= 'Z' && s[1] >= 'A' && s[1] <= 'Z' {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}
