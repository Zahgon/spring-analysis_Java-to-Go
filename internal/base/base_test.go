package base_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/seaswalker/spring-analysis/internal/annotation"
	"github.com/seaswalker/spring-analysis/internal/base"
	"github.com/seaswalker/spring-analysis/internal/base/scope"
	"github.com/seaswalker/spring-analysis/internal/deps/springframework/beans"
	"github.com/seaswalker/spring-analysis/internal/testsupport"
)

// TestStudentToString keeps the exact rendering the original produced.
func TestStudentToString(t *testing.T) {
	s := base.NewStudentOf("skywalker", 22)
	if got, want := s.String(), "Student [name=skywalker, age=22]"; got != want {
		t.Errorf("String = %q, want %q", got, want)
	}
	empty := base.NewStudent()
	if got, want := empty.String(), "Student [name=, age=0]"; got != want {
		t.Errorf("String of an unset student = %q, want %q", got, want)
	}
}

// TestStudentProperties covers the accessors, including the inherited id.
func TestStudentProperties(t *testing.T) {
	s := base.NewStudent()
	s.SetName("a")
	s.SetAge(7)
	s.SetId("id-1")
	if s.GetName() != "a" || s.GetAge() != 7 || s.GetId() != "id-1" {
		t.Errorf("accessors disagree with the setters: %+v", s)
	}
}

// TestSimpleBean covers the bean's message, its @Init-marked method, and the
// marker itself.
func TestSimpleBean(t *testing.T) {
	student := base.NewStudentOf("s", 1)
	bean := base.NewSimpleBeanOf(student)
	if bean.GetStudent() != student {
		t.Error("the constructor did not wire the student")
	}

	if out := testsupport.CaptureStdout(t, bean.Send); out != "I am send method from SimpleBean!\n" {
		t.Errorf("Send printed %q", out)
	}
	if out := testsupport.CaptureStdout(t, bean.Init); out != "Init!\n" {
		t.Errorf("Init printed %q", out)
	}

	if got := annotation.MethodsOf(bean); len(got) != 1 || got[0] != "Init" {
		t.Errorf("@Init methods = %v, want [Init]", got)
	}
	if got := annotation.MethodsOf(student); got != nil {
		t.Errorf("a type with no marker reported %v", got)
	}
	if annotation.Name != "@Init" {
		t.Errorf("marker name = %q", annotation.Name)
	}

	empty := base.NewSimpleBean()
	empty.SetStudent(student)
	if empty.GetStudent() != student {
		t.Error("SetStudent did not take")
	}
}

// TestSimpleBeanFactoryPostProcessor covers the rename it performs.
func TestSimpleBeanFactoryPostProcessor(t *testing.T) {
	ctx := beans.NewApplicationContext()
	ctx.RegisterBeanDefinition(&beans.BeanDefinition{
		Name: "simpleBean",
		Type: reflect.TypeOf(&base.SimpleBean{}),
		Create: func(beans.BeanFactory) (any, error) {
			return base.NewSimpleBeanOf(base.NewStudentOf("before", 1)), nil
		},
	})
	ctx.AddBeanFactoryPostProcessor(base.NewSimpleBeanFactoryPostProcessor())
	if err := ctx.Refresh(); err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	bean, _ := ctx.GetBean("simpleBean")
	if got := bean.(*base.SimpleBean).GetStudent().GetName(); got != "^_^" {
		t.Errorf("name = %q, want %q", got, "^_^")
	}

	empty := beans.NewApplicationContext()
	empty.AddBeanFactoryPostProcessor(base.NewSimpleBeanFactoryPostProcessor())
	if err := empty.Refresh(); err == nil {
		t.Error("the post-processor succeeded with no SimpleBean to find")
	}
}

// TestSimpleBeanPostProcessorReturnsNil keeps the deliberately broken
// processor broken.
func TestSimpleBeanPostProcessorReturnsNil(t *testing.T) {
	p := base.NewSimpleBeanPostProcessor()
	bean := base.NewSimpleBean()
	for _, call := range []func(any, string) (any, error){
		p.PostProcessBeforeInitialization,
		p.PostProcessAfterInitialization,
	} {
		got, err := call(bean, "simpleBean")
		if err != nil {
			t.Fatalf("callback returned an error: %v", err)
		}
		if got != nil {
			t.Errorf("callback returned %v, want nil", got)
		}
	}
}

// TestOneScope covers the counter's off-by-one between the name and the age,
// and the callbacks that report nothing.
func TestOneScope(t *testing.T) {
	s := scope.NewOneScope()
	factory := beans.ObjectFactoryFunc(func() (any, error) {
		t.Error("OneScope must not consult the ObjectFactory")
		return nil, nil
	})

	var students []*base.Student
	out := testsupport.CaptureStdout(t, func() {
		for i := 0; i < 2; i++ {
			students = append(students, s.Get("student", factory).(*base.Student))
		}
	})
	if got := strings.Count(out, "get被调用"); got != 2 {
		t.Errorf("printed %d times, want 2:\n%s", got, out)
	}
	want := []string{"Student [name=skywalker-0, age=1]", "Student [name=skywalker-1, age=2]"}
	for i, w := range want {
		if got := students[i].String(); got != w {
			t.Errorf("call %d produced %q, want %q", i, got, w)
		}
	}

	if s.Remove("student") != nil || s.ResolveContextualObject("k") != nil || s.GetConversationID() != "" {
		t.Error("an unsupported callback reported a value")
	}
	s.RegisterDestructionCallback("student", func() { t.Error("the destruction callback ran") })
}
