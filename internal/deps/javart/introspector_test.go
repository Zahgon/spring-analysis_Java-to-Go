package javart_test

import (
	"strings"
	"testing"

	"github.com/seaswalker/spring-analysis/internal/base"
	"github.com/seaswalker/spring-analysis/internal/deps/javart"
)

// TestGetBeanInfoDiscoversInheritedProperties covers what the intro
// demonstration prints: properties from the type and from what it embeds,
// sorted by name, with the read-only one carrying no setter.
func TestGetBeanInfoDiscoversInheritedProperties(t *testing.T) {
	descriptors := javart.GetBeanInfo(&base.Student{})

	var names []string
	byName := map[string]javart.PropertyDescriptor{}
	for _, d := range descriptors {
		names = append(names, d.Name)
		byName[d.Name] = d
	}

	want := []string{"age", "class", "id", "name"}
	if strings.Join(names, ",") != strings.Join(want, ",") {
		t.Fatalf("properties = %v, want %v (sorted, inherited included)", names, want)
	}
	if byName["class"].WriteMethod != "" {
		t.Errorf("the read-only property has a write method: %q", byName["class"].WriteMethod)
	}
	for _, n := range []string{"age", "id", "name"} {
		if byName[n].ReadMethod == "" || byName[n].WriteMethod == "" {
			t.Errorf("property %q is missing an accessor: %+v", n, byName[n])
		}
	}
	if !strings.Contains(byName["id"].ReadMethod, "GetId") {
		t.Errorf("inherited read method = %q", byName["id"].ReadMethod)
	}
	if !strings.Contains(byName["age"].ReadMethod, "int") {
		t.Errorf("age read method does not report its result type: %q", byName["age"].ReadMethod)
	}
}

// TestGetBeanInfoOfNil returns nothing rather than failing.
func TestGetBeanInfoOfNil(t *testing.T) {
	if got := javart.GetBeanInfo(nil); got != nil {
		t.Errorf("GetBeanInfo(nil) = %v, want nil", got)
	}
}
