// Package task holds the Spring Task demonstration.
//
// The original's method carries @Async("executor"), naming an executor bean.
// Neither XML file declares a <task:*> element, so no executor named
// "executor" is ever configured and the annotation has no effect: the method
// runs on the calling thread. That is the recorded behaviour, and the
// declaration is preserved rather than being turned into a goroutine, which
// would be a different program.
package task

import "fmt"

// AsyncExecutorName is the executor @Async names.
const AsyncExecutorName = "executor"

// Task is the bean carrying the asynchronous method.
type Task struct{}

// NewTask returns the bean.
func NewTask() *Task { return &Task{} }

// Print prints the task message.
func (t *Task) Print() { fmt.Println("print执行") }

// AsyncMethods declares which methods carry @Async and which executor each
// names, since Go has no annotations to read back.
func (t *Task) AsyncMethods() map[string]string {
	return map[string]string{"Print": AsyncExecutorName}
}
