package local_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/seaswalker/spring-analysis/internal/local"
	"github.com/seaswalker/spring-analysis/internal/testsupport"
)

// TestResourceBoundle is the original's Local.resourceBoundle: the same key
// read in two locales, with the trailing space each value carries preserved
// and the Chinese file's \uXXXX escapes decoded.
func TestResourceBoundle(t *testing.T) {
	classPath := filepath.Join(testsupport.RepoRoot(t), "resources")
	var err error
	lines := testsupport.Lines(testsupport.CaptureStdout(t, func() {
		err = local.ResourceBundleDemo(classPath)
	}))
	if err != nil {
		t.Fatalf("ResourceBundleDemo: %v", err)
	}
	want := []string{"US: How are you! ", "CN: 您好！ "}
	if strings.Join(lines, "\n") != strings.Join(want, "\n") {
		t.Errorf("got:\n%q\nwant:\n%q", lines, want)
	}
}
