package base

import (
	"fmt"

	"github.com/seaswalker/spring-analysis/internal/annotation"
)

// SimpleBean is the bean the container demonstrations wire a Student into.
type SimpleBean struct {
	student *Student
}

// NewSimpleBean is the no-argument constructor.
func NewSimpleBean() *SimpleBean { return &SimpleBean{} }

// NewSimpleBeanOf is the (Student) constructor.
func NewSimpleBeanOf(student *Student) *SimpleBean {
	return &SimpleBean{student: student}
}

// GetStudent returns the wired student.
func (b *SimpleBean) GetStudent() *Student { return b.student }

// SetStudent wires a student.
func (b *SimpleBean) SetStudent(student *Student) { b.student = student }

// Send prints the bean's message.
func (b *SimpleBean) Send() {
	fmt.Println("I am send method from SimpleBean!")
}

// Init is the @Init-marked initialisation callback.
func (b *SimpleBean) Init() {
	fmt.Println("Init!")
}

// InitMethods declares Init as the @Init-marked method, which is what the
// annotation carried in the original.
func (b *SimpleBean) InitMethods() []string { return []string{"Init"} }

// SimpleBean satisfies the marker at compile time.
var _ annotation.Initializer = (*SimpleBean)(nil)
