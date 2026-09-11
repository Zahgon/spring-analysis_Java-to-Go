package main

import (
	"testing"

	"github.com/seaswalker/spring-analysis/internal/testsupport"
)

// TestRun covers the command's body.
func TestRun(t *testing.T) {
	if got := testsupport.CaptureStdout(t, run); got != "name: Get, return: string\n" {
		t.Errorf("run printed %q", got)
	}
}
