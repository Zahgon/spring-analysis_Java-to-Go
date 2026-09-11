// Package aop holds the AOP demonstration beans the notes refer to. Nothing
// in either XML file registers them; they exist to show the shapes advice can
// be attached to.
package aop

import "fmt"

// AopDemoInter is the interface AopDemo implements, so that the demonstration
// covers interface-based advising as well as class-based.
type AopDemoInter interface {
	Inter()
}

// AopDemo is the advised target.
type AopDemo struct{}

// NewAopDemo returns the target.
func NewAopDemo() *AopDemo { return &AopDemo{} }

// Send prints the send message.
func (d *AopDemo) Send() { fmt.Println("send from aopdemo") }

// Receive prints the receive message.
func (d *AopDemo) Receive() { fmt.Println("receive from aopdemo") }

// Inter prints the interface message.
func (d *AopDemo) Inter() { fmt.Println("inter") }

var _ AopDemoInter = (*AopDemo)(nil)
