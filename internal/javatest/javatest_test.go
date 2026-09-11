package javatest_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/seaswalker/spring-analysis/internal/javatest"
	"github.com/seaswalker/spring-analysis/internal/testsupport"
)

// TestClasspath is the original's JavaTest.classpath: it prints the code units
// the running program was built from. The value is environment-specific in
// both languages, so what is asserted is that it is produced and names the
// module under test.
func TestClasspath(t *testing.T) {
	out := testsupport.CaptureStdout(t, javatest.Classpath)
	if strings.TrimSpace(out) == "" {
		t.Fatal("classpath printed nothing")
	}
	if !strings.Contains(out, "spring-analysis") {
		t.Errorf("classpath does not name the module:\n%s", out)
	}
}

// TestFindClass is the original's JavaTest.findClass: the resource lookup is
// by literal path, so the asterisk matches nothing and the demonstration
// prints no lines.
func TestFindClass(t *testing.T) {
	root := testsupport.RepoRoot(t)
	out := testsupport.CaptureStdout(t, func() {
		javatest.FindClass(filepath.Join(root, "resources"))
	})
	if out != "" {
		t.Errorf("findClass printed %q, want nothing: a literal lookup must not glob", out)
	}
}

// TestIntro is the original's JavaTest.intro: it prints Student's property
// descriptors, in property-name order, with a missing setter printed as null.
func TestIntro(t *testing.T) {
	lines := testsupport.Lines(testsupport.CaptureStdout(t, javatest.Intro))
	if len(lines)%2 != 0 {
		t.Fatalf("expected a read/write line pair per property, got %d lines", len(lines))
	}

	// age, class, id and name — the bean's three properties plus the
	// read-only one the root of the hierarchy contributes, alphabetically.
	wantProperties := []string{"GetAge", "Type", "GetId", "GetName"}
	if len(lines) != 2*len(wantProperties) {
		t.Fatalf("got %d lines for %d properties:\n%s", len(lines), len(wantProperties), strings.Join(lines, "\n"))
	}
	for i, want := range wantProperties {
		if !strings.Contains(lines[2*i], want) {
			t.Errorf("property %d read method = %q, want it to mention %q", i, lines[2*i], want)
		}
	}
	// The class property has no setter, so its write line is null.
	if lines[3] != "null" {
		t.Errorf("write method of the read-only property = %q, want %q", lines[3], "null")
	}
	for _, i := range []int{1, 5, 7} {
		if !strings.HasPrefix(lines[i], "func ") {
			t.Errorf("line %d = %q, want a setter signature", i, lines[i])
		}
	}
}

// TestSplit is the original's JavaTest.split: the record splits on tabs, the
// parts render with ", " between them, printing the slice itself prints an
// identity rather than the contents, and the two labelled columns are read by
// index.
func TestSplit(t *testing.T) {
	lines := testsupport.Lines(testsupport.CaptureStdout(t, javatest.Split))
	want := []string{
		"[1, 2, aug, fri, 14.7, 66, 2.7, 0, 0]",
		"", // identity — asserted below, since it differs per run
		"月份: aug",
		"天气: 14.7",
	}
	if len(lines) != len(want) {
		t.Fatalf("got %d lines, want %d:\n%s", len(lines), len(want), strings.Join(lines, "\n"))
	}
	for i, w := range want {
		if i == 1 {
			continue
		}
		if lines[i] != w {
			t.Errorf("line %d = %q, want %q", i, lines[i], w)
		}
	}
	if !strings.HasPrefix(lines[1], "[]string@") {
		t.Errorf("identity line = %q, want a type descriptor and an address", lines[1])
	}
	if strings.Contains(lines[1], "aug") {
		t.Errorf("identity line %q printed the contents; the demonstration is that it does not", lines[1])
	}
}

// TestDeclaredMethods covers the original's JavaTest.main: it prints the
// declared methods of the narrowed list type.
//
// Java reports two — javac emits a bridge method beside a covariant override.
// Go has no bridge methods, so the narrowed method simply shadows the embedded
// one and exactly one is reported. The demonstration prints what the language
// reports, so the assertion is the Go answer.
func TestDeclaredMethods(t *testing.T) {
	got := javatest.DeclaredMethods()
	want := []string{"name: Get, return: string"}
	if len(got) != len(want) {
		t.Fatalf("got %d declared methods, want %d: %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("method %d = %q, want %q", i, got[i], want[i])
		}
	}
}

// TestNarrowedAccessorShadowsTheEmbeddedOne covers the other half of what the
// bridge method exists for in Java: both accessors are present, the narrowed
// one on the type itself and the widened one on the value it embeds.
func TestNarrowedAccessorShadowsTheEmbeddedOne(t *testing.T) {
	m := javatest.NewMyList()

	var narrowed string = m.Get(0)
	if narrowed != "" {
		t.Errorf("narrowed accessor returned %q, want the empty string", narrowed)
	}

	var widened any = m.List.Get(0)
	if widened != nil {
		t.Errorf("widened accessor returned %v, want nil", widened)
	}
}
