package main

import (
	"testing"

	"github.com/seaswalker/spring-analysis/internal/testsupport"
)

// TestRun covers the command's body.
func TestRun(t *testing.T) {
	want := "Method printName is proxyed.\nName is XXX\nAge: 18\n"
	if got := testsupport.CaptureStdout(t, run); got != want {
		t.Errorf("run printed %q, want %q", got, want)
	}
}
