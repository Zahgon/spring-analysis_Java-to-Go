package javaconfig_test

import (
	"strings"
	"testing"

	"github.com/seaswalker/spring-analysis/internal/base"
	"github.com/seaswalker/spring-analysis/internal/deps/javart"
	"github.com/seaswalker/spring-analysis/internal/deps/springframework/beans"
	springctx "github.com/seaswalker/spring-analysis/internal/deps/springframework/context"
	"github.com/seaswalker/spring-analysis/internal/javaconfig"
	"github.com/seaswalker/spring-analysis/internal/testsupport"
)

// definitionNames is the registry the demonstration prints, verbatim from the
// original.
const definitionNames = "[org.springframework.context.annotation.internalConfigurationAnnotationProcessor, " +
	"org.springframework.context.annotation.internalAutowiredAnnotationProcessor, " +
	"org.springframework.context.annotation.internalRequiredAnnotationProcessor, " +
	"org.springframework.context.annotation.internalCommonAnnotationProcessor, " +
	"org.springframework.context.event.internalEventListenerProcessor, " +
	"org.springframework.context.event.internalEventListenerFactory, " +
	"simpleBeanConfig, java_config.StudentConfig, student, simpleBean]"

// TestBeanDefinitionNamesAndOrder covers the container's naming and ordering:
// the six internal processors first, the directly-registered configuration
// under its decapitalised simple name, the imported one under its
// fully-qualified name, then the bean methods with the imported class's first.
func TestBeanDefinitionNamesAndOrder(t *testing.T) {
	var ctx *beans.ApplicationContext
	var err error
	testsupport.CaptureStdout(t, func() {
		ctx, err = springctx.NewAnnotationConfigApplicationContext(javaconfig.NewSimpleBeanConfig())
	})
	if err != nil {
		t.Fatalf("building the context: %v", err)
	}
	if got := javart.ArraysToString(ctx.GetBeanDefinitionNames()); got != definitionNames {
		t.Errorf("definition names:\ngot:  %s\nwant: %s", got, definitionNames)
	}
}

// TestImportAwareFiresBeforeAnyUserLookup covers the ordering the
// demonstration's first two lines show.
func TestImportAwareFiresBeforeAnyUserLookup(t *testing.T) {
	var ctx *beans.ApplicationContext
	var err error
	refresh := testsupport.Lines(testsupport.CaptureStdout(t, func() {
		ctx, err = springctx.NewAnnotationConfigApplicationContext(javaconfig.NewSimpleBeanConfig())
	}))
	if err != nil {
		t.Fatalf("building the context: %v", err)
	}
	if len(refresh) != 1 || refresh[0] != "importaware" {
		t.Fatalf("refresh printed %v, want [importaware]", refresh)
	}

	after := testsupport.CaptureStdout(t, func() {
		bean, err := beans.GetBeanAs[*base.SimpleBean](ctx)
		if err != nil {
			t.Errorf("getting simpleBean: %v", err)
			return
		}
		println(bean.GetStudent().GetName())
	})
	if strings.Contains(after, "importaware") {
		t.Error("ImportAware fired again on a later lookup")
	}
}

// TestSimpleBeanCarriesTheConfiguredStudent covers what the demonstration's
// second line prints.
func TestSimpleBeanCarriesTheConfiguredStudent(t *testing.T) {
	var ctx *beans.ApplicationContext
	var err error
	testsupport.CaptureStdout(t, func() {
		ctx, err = springctx.NewAnnotationConfigApplicationContext(javaconfig.NewSimpleBeanConfig())
	})
	if err != nil {
		t.Fatalf("building the context: %v", err)
	}
	bean, err := beans.GetBeanAs[*base.SimpleBean](ctx)
	if err != nil {
		t.Fatalf("getting simpleBean: %v", err)
	}
	student := bean.GetStudent()
	if student.GetName() != "skywalker" || student.GetAge() != 22 {
		t.Errorf("student = %v, want name skywalker aged 22", student)
	}
}

// TestStudentIsPrototypeScoped covers the @Scope("prototype") declaration.
func TestStudentIsPrototypeScoped(t *testing.T) {
	var ctx *beans.ApplicationContext
	testsupport.CaptureStdout(t, func() {
		ctx, _ = springctx.NewAnnotationConfigApplicationContext(javaconfig.NewSimpleBeanConfig())
	})
	first, err := ctx.GetBean("student")
	if err != nil {
		t.Fatalf("GetBean(student): %v", err)
	}
	second, _ := ctx.GetBean("student")
	if first == second {
		t.Error("student is a singleton, want a prototype")
	}
}

// TestConfigurationClassNames keeps the names the registry prints.
func TestConfigurationClassNames(t *testing.T) {
	if got := javaconfig.NewSimpleBeanConfig().ConfigurationClassName(); got != "java_config.SimpleBeanConfig" {
		t.Errorf("SimpleBeanConfig class name = %q", got)
	}
	if got := javaconfig.NewStudentConfig().ConfigurationClassName(); got != "java_config.StudentConfig" {
		t.Errorf("StudentConfig class name = %q", got)
	}
	if imports := javaconfig.NewSimpleBeanConfig().Imports(); len(imports) != 1 {
		t.Errorf("Imports = %v, want one entry", imports)
	}
}

// TestStudentBeanMethod covers the factory method the configuration declares.
func TestStudentBeanMethod(t *testing.T) {
	student := javaconfig.NewStudentConfig().Student()
	if student.GetName() != "skywalker" || student.GetAge() != 22 {
		t.Errorf("Student() = %v", student)
	}
	methods := javaconfig.NewStudentConfig().BeanMethods()
	if len(methods) != 1 || methods[0].Name != "student" || methods[0].Scope != beans.ScopePrototype {
		t.Errorf("BeanMethods = %+v", methods)
	}
}
