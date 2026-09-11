package main

import (
	"testing"

	"github.com/seaswalker/spring-analysis/internal/deps/springframework/beans"
	"github.com/seaswalker/spring-analysis/internal/testsupport"
)

// TestRun covers the command's body from the repository root, where its
// default classpath resolves.
func TestRun(t *testing.T) {
	testsupport.ChdirRepoRoot(t)
	testsupport.CaptureStdout(t, func() {
		if err := run(); err != nil {
			t.Errorf("run: %v", err)
		}
	})
}

// TestRunWithoutConfiguration covers the failure path: no config.xml on the
// classpath means no container.
func TestRunWithoutConfiguration(t *testing.T) {
	saved := beans.ClassPath
	beans.ClassPath = t.TempDir()
	defer func() { beans.ClassPath = saved }()
	if err := run(); err == nil {
		t.Error("run succeeded with no configuration on the classpath")
	}
}
