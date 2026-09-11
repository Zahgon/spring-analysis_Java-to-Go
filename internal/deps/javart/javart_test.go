package javart_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/seaswalker/spring-analysis/internal/deps/javart"
	"github.com/seaswalker/spring-analysis/internal/testsupport"
)

// TestArraysToString covers the rendering the bean-definition listing and the
// split demonstration both print through — ", " between elements, not Go's
// single space.
func TestArraysToString(t *testing.T) {
	cases := []struct {
		name  string
		input any
		want  string
	}{
		{"empty", []string{}, "[]"},
		{"single", []string{"a"}, "[a]"},
		{"several", []string{"1", "2", "aug"}, "[1, 2, aug]"},
		{"ints", []int{3, 1, 4}, "[3, 1, 4]"},
		{"nil slice", []string(nil), "null"},
		{"nil", nil, "null"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := javart.ArraysToString(tc.input); got != tc.want {
				t.Errorf("ArraysToString(%v) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

// TestToStringRendersNullForAbsentValues checks the four characters a null
// prints as, which SimpleModel.toString depends on.
func TestToStringRendersNullForAbsentValues(t *testing.T) {
	var absent *int
	present := 30
	cases := []struct {
		input any
		want  string
	}{
		{nil, "null"},
		{absent, "null"},
		{&present, "30"},
		{"text", "text"},
		{30, "30"},
	}
	for _, tc := range cases {
		if got := javart.ToString(tc.input); got != tc.want {
			t.Errorf("ToString(%v) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

// TestIdentityStringDoesNotPrintContents covers the demonstration's point:
// printing an array reference shows an identity, not the elements.
func TestIdentityStringDoesNotPrintContents(t *testing.T) {
	got := javart.IdentityString([]string{"aug", "fri"})
	if !strings.HasPrefix(got, "[]string@") {
		t.Errorf("IdentityString = %q, want a type descriptor and an address", got)
	}
	if strings.Contains(got, "aug") {
		t.Errorf("IdentityString = %q, want it not to print the contents", got)
	}
}

// TestLoadProperties covers the .properties rules the localisation
// demonstration relies on: whitespace around the separator is dropped but a
// trailing space in the value survives, and \uXXXX escapes decode.
func TestLoadProperties(t *testing.T) {
	input := strings.Join([]string{
		"# a comment",
		"! another comment",
		"",
		"greeting.common=How are you! ",
		"greeting.morning = Good morning! ",
		"chinese=\\u60a8\\u597d\\uff01 ",
		"student.name=  skywalker",
		"escaped\\=key=value",
		"colon:separated",
		"continued=one\\",
		"two",
		"tabbed=a\\tb",
	}, "\n")

	props, err := javart.LoadProperties(strings.NewReader(input))
	if err != nil {
		t.Fatalf("LoadProperties: %v", err)
	}
	want := map[string]string{
		"greeting.common":  "How are you! ",
		"greeting.morning": "Good morning! ",
		"chinese":          "您好！ ",
		"student.name":     "skywalker",
		"escaped=key":      "value",
		"colon":            "separated",
		"continued":        "onetwo",
		"tabbed":           "a\tb",
	}
	for k, w := range want {
		if got := props[k]; got != w {
			t.Errorf("props[%q] = %q, want %q", k, got, w)
		}
	}
	if len(props) != len(want) {
		t.Errorf("got %d entries, want %d: %v", len(props), len(want), props)
	}
}

// TestGetBundleLocaleFallback checks the candidate order: an exact
// language_COUNTRY match wins, and an unknown locale falls back to the base
// bundle.
func TestGetBundleLocaleFallback(t *testing.T) {
	root := filepath.Join(testsupport.RepoRoot(t), "resources")

	us, err := javart.GetBundle(root, "resource/resource", javart.LocaleUS)
	if err != nil {
		t.Fatalf("GetBundle(en_US): %v", err)
	}
	if got, _ := us.GetString("greeting.common"); got != "How are you! " {
		t.Errorf("en_US greeting = %q", got)
	}
	if len(us.Keys()) != 3 {
		t.Errorf("en_US bundle has %d keys, want 3", len(us.Keys()))
	}
	if us.Locale().String() != "en_US" {
		t.Errorf("locale = %q, want %q", us.Locale(), "en_US")
	}

	cn, err := javart.GetBundle(root, "resource/resource", javart.LocaleChina)
	if err != nil {
		t.Fatalf("GetBundle(zh_CN): %v", err)
	}
	if got, _ := cn.GetString("greeting.common"); got != "您好！ " {
		t.Errorf("zh_CN greeting = %q", got)
	}
	if _, err := cn.GetString("absent"); err == nil {
		t.Error("GetString for an absent key succeeded")
	}

	if _, err := javart.GetBundle(root, "resource/missing", javart.LocaleUS); err == nil {
		t.Error("GetBundle for a missing base name succeeded")
	} else if _, ok := err.(*javart.MissingResourceError); !ok {
		t.Errorf("error is %T, want *javart.MissingResourceError", err)
	}
}

// TestParseLocale covers the "lang_COUNTRY" round trip.
func TestParseLocale(t *testing.T) {
	for _, s := range []string{"en_US", "zh_CN", "fr"} {
		if got := javart.ParseLocale(s).String(); got != s {
			t.Errorf("ParseLocale(%q).String() = %q", s, got)
		}
	}
}

// TestClassLoaderDoesNotGlob is the behaviour the findClass demonstration
// prints nothing because of.
func TestClassLoaderDoesNotGlob(t *testing.T) {
	root := filepath.Join(testsupport.RepoRoot(t), "resources")
	loader := javart.NewClassLoader(root)

	if got := loader.GetResources("resource/*"); got != nil {
		t.Errorf("GetResources with a wildcard returned %v, want nothing", got)
	}
	if got := loader.GetResource("resource/resource_en_US.properties"); got == "" {
		t.Error("a literal resource path resolved to nothing")
	} else if !strings.HasPrefix(got, "file:") {
		t.Errorf("resource URL = %q, want a file: URL", got)
	}
	if got := loader.GetResources("resource"); len(got) != 2 {
		t.Errorf("directory enumeration returned %d entries, want 2: %v", len(got), got)
	}
	if got := loader.GetResource("absent"); got != "" {
		t.Errorf("an absent resource resolved to %q", got)
	}
}

// TestGoMethodName covers the camelCase mapping that keeps a printed method
// name in the source language's spelling while dispatch reaches the Go one.
func TestGoMethodName(t *testing.T) {
	cases := map[string]string{"testB": "TestB", "boo": "Boo", "printName": "PrintName", "": ""}
	for in, want := range cases {
		if got := javart.GoMethodName(in); got != want {
			t.Errorf("GoMethodName(%q) = %q, want %q", in, got, want)
		}
		if got := javart.JavaMethodName(want); got != in {
			t.Errorf("JavaMethodName(%q) = %q, want %q", want, got, in)
		}
	}
}

// TestCodePathNamesTheModule checks the classpath analogue reports the running
// program's code units in every build mode — including a test binary, where
// build info carries no module path under some toolchains.
func TestCodePathNamesTheModule(t *testing.T) {
	got := javart.CodePath()
	if strings.TrimSpace(got) == "" {
		t.Fatal("CodePath is empty")
	}
	if !strings.Contains(got, javart.ModulePath()) {
		t.Errorf("CodePath %q does not name the module %q", got, javart.ModulePath())
	}
}

// TestModulePath pins the module path the trimmed package path yields. Moving
// this package without updating its declared suffix fails here rather than
// silently reporting a truncated path.
func TestModulePath(t *testing.T) {
	got := javart.ModulePath()
	if got != "github.com/seaswalker/spring-analysis" {
		t.Errorf("ModulePath = %q, want the module path with no package suffix left on it", got)
	}
}

// TestMissingResourceErrorMessage keeps the wording java.util.ResourceBundle
// reports for a bundle it cannot find.
func TestMissingResourceErrorMessage(t *testing.T) {
	err := &javart.MissingResourceError{BaseName: "resource/resource", Locale: javart.LocaleUS}
	want := "Can't find bundle for base name resource/resource, locale en_US"
	if err.Error() != want {
		t.Errorf("message = %q, want %q", err.Error(), want)
	}
}

// TestShippedPropertyFile reads the resource the repository ships but never
// loads, checking that its leading whitespace is handled the way
// java.util.Properties handles it: dropped from the front of a value.
func TestShippedPropertyFile(t *testing.T) {
	f, err := os.Open(filepath.Join(testsupport.RepoRoot(t), "resources", "property.properties"))
	if err != nil {
		t.Fatalf("opening the shipped property file: %v", err)
	}
	defer f.Close()

	props, err := javart.LoadProperties(f)
	if err != nil {
		t.Fatalf("LoadProperties: %v", err)
	}
	want := map[string]string{
		"student.name": "skywalker",
		"student.age":  "20",
		"student.id":   "id",
	}
	for k, w := range want {
		if got := props[k]; got != w {
			t.Errorf("props[%q] = %q, want %q", k, got, w)
		}
	}
}
