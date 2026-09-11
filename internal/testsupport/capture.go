// Package testsupport holds helpers the demonstration tests share.
//
// The application's contract is what it prints, so almost every test needs to
// run something and read back stdout.
package testsupport

import (
	"io"
	"os"
	"strings"
	"testing"
)

// CaptureStdout runs fn with os.Stdout redirected and returns what it wrote.
func CaptureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("creating pipe: %v", err)
	}
	saved := os.Stdout
	os.Stdout = w

	done := make(chan string, 1)
	go func() {
		var sb strings.Builder
		_, _ = io.Copy(&sb, r)
		done <- sb.String()
	}()

	func() {
		defer func() {
			os.Stdout = saved
			_ = w.Close()
		}()
		fn()
	}()

	out := <-done
	_ = r.Close()
	return out
}

// Lines splits captured output into lines, dropping the trailing empty one a
// final newline leaves behind.
func Lines(out string) []string {
	out = strings.TrimSuffix(out, "\n")
	if out == "" {
		return nil
	}
	return strings.Split(out, "\n")
}

// ChdirRepoRoot moves the test process to the repository root for the
// duration of the test, which is where the commands expect to be run from.
func ChdirRepoRoot(t *testing.T) {
	t.Helper()
	saved, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(RepoRoot(t)); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(saved); err != nil {
			t.Errorf("restoring the working directory: %v", err)
		}
	})
}

// RepoRoot returns the repository root, so a test can reach the resources and
// views regardless of which package directory it runs in.
func RepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(dir + "/go.mod"); err == nil {
			return dir
		}
		parent := dir[:strings.LastIndex(dir, "/")]
		if parent == "" || parent == dir {
			t.Fatal("could not find the repository root")
		}
		dir = parent
	}
}
