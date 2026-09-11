package task_test

import (
	"testing"

	"github.com/seaswalker/spring-analysis/internal/task"
	"github.com/seaswalker/spring-analysis/internal/testsupport"
)

// TestPrint covers the task's message.
func TestPrint(t *testing.T) {
	if got := testsupport.CaptureStdout(t, task.NewTask().Print); got != "print执行\n" {
		t.Errorf("Print printed %q", got)
	}
}

// TestAsyncDeclaration keeps the executor name the @Async annotation carried.
func TestAsyncDeclaration(t *testing.T) {
	methods := task.NewTask().AsyncMethods()
	if got, ok := methods["Print"]; !ok || got != task.AsyncExecutorName {
		t.Errorf("@Async executor for Print = %q (found %v), want %q", got, ok, task.AsyncExecutorName)
	}
	if len(methods) != 1 {
		t.Errorf("declared %d async methods, want 1", len(methods))
	}
}
