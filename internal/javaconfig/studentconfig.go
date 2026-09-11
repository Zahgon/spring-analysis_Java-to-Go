// Package javaconfig holds the annotation-driven container configuration.
//
// The original package is named java_config; Go package names are lowercase
// and unpunctuated, so the directory carries the name and the original's
// fully-qualified class names are declared explicitly, because the container
// prints them.
package javaconfig

import (
	"fmt"
	"reflect"

	"github.com/seaswalker/spring-analysis/internal/base"
	"github.com/seaswalker/spring-analysis/internal/deps/springframework/beans"
	springctx "github.com/seaswalker/spring-analysis/internal/deps/springframework/context"
)

// StudentConfigClassName is the name the original class had. The container
// registers an imported configuration class under its fully-qualified name,
// and the demonstration prints the registry, so this string is contract.
const StudentConfigClassName = "java_config.StudentConfig"

// StudentConfig is the @Configuration declaring the Student bean.
type StudentConfig struct{}

// NewStudentConfig returns the configuration.
func NewStudentConfig() *StudentConfig { return &StudentConfig{} }

// ConfigurationClassName returns the original class name.
func (c *StudentConfig) ConfigurationClassName() string { return StudentConfigClassName }

// Student is the @Bean @Scope("prototype") factory method.
func (c *StudentConfig) Student() *base.Student {
	student := base.NewStudent()
	student.SetAge(22)
	student.SetName("skywalker")
	return student
}

// BeanMethods declares Student as a prototype-scoped bean named for the
// method, which is what Spring's default bean-name generator produces.
func (c *StudentConfig) BeanMethods() []springctx.BeanMethod {
	return []springctx.BeanMethod{{
		Name:    "student",
		Scope:   beans.ScopePrototype,
		Type:    reflect.TypeOf(&base.Student{}),
		Factory: func() (any, error) { return c.Student(), nil },
	}}
}

// SetImportMetadata is the ImportAware callback, invoked when the class is
// instantiated because it was imported.
func (c *StudentConfig) SetImportMetadata(metadata springctx.AnnotationMetadata) {
	fmt.Println("importaware")
}

var (
	_ springctx.Configuration = (*StudentConfig)(nil)
	_ springctx.ImportAware   = (*StudentConfig)(nil)
)
