package main

import (
	"strings"
	"testing"

	"github.com/seaswalker/spring-analysis/internal/testsupport"
)

// TestRun covers the command's body and the three lines it prints.
func TestRun(t *testing.T) {
	testsupport.ChdirRepoRoot(t)
	var err error
	lines := testsupport.Lines(testsupport.CaptureStdout(t, func() { err = run() }))
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if len(lines) != 3 {
		t.Fatalf("printed %d lines, want 3:\n%s", len(lines), strings.Join(lines, "\n"))
	}
	if lines[0] != "importaware" || lines[1] != "skywalker" {
		t.Errorf("first two lines = %q, %q", lines[0], lines[1])
	}
	if !strings.HasSuffix(lines[2], "simpleBeanConfig, java_config.StudentConfig, student, simpleBean]") {
		t.Errorf("definition list = %q", lines[2])
	}
}
