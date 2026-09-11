package base

import "fmt"

// BaseStudent holds the identity property Student inherits.
//
// The original is an abstract class; Go expresses "shared state and behaviour
// a concrete type picks up" by embedding, and abstractness by not registering
// BaseStudent as a bean in its own right.
type BaseStudent struct {
	id string
}

// GetId returns the student's id.
func (b *BaseStudent) GetId() string { return b.id }

// SetId sets the student's id.
func (b *BaseStudent) SetId(id string) { b.id = id }

// Student is the model bean the container demonstrations pass around.
type Student struct {
	BaseStudent
	name string
	age  int
}

// NewStudent is the no-argument constructor.
func NewStudent() *Student { return &Student{} }

// NewStudentOf is the (String, int) constructor. The original assigns age
// before name; the order is not observable, but the signature is.
func NewStudentOf(name string, age int) *Student {
	return &Student{name: name, age: age}
}

// GetName returns the student's name.
func (s *Student) GetName() string { return s.name }

// SetName sets the student's name.
func (s *Student) SetName(name string) { s.name = name }

// GetAge returns the student's age.
func (s *Student) GetAge() int { return s.age }

// SetAge sets the student's age.
func (s *Student) SetAge(age int) { s.age = age }

// String reproduces Student.toString() exactly.
func (s *Student) String() string {
	return fmt.Sprintf("Student [name=%s, age=%d]", s.name, s.age)
}
