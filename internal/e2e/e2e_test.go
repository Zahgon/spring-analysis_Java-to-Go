// Package e2e runs the built commands and asserts on what they print, which
// is the application's whole observable contract.
package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// binaries holds the path of each command, built once for the package.
var binaries map[string]string

// repoRoot is the directory the commands run from, so that the default
// "resources" and "web" paths resolve.
var repoRoot string

func TestMain(m *testing.M) {
	var err error
	repoRoot, err = findRepoRoot()
	if err != nil {
		panic(err)
	}
	dir, err := os.MkdirTemp("", "spring-analysis-e2e")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	binaries = map[string]string{}
	for _, cmd := range []string{"aop", "base", "javaconfig", "jdkproxy", "javatest"} {
		out := filepath.Join(dir, cmd)
		build := exec.Command("go", "build", "-o", out, "./cmd/"+cmd)
		build.Dir = repoRoot
		if output, err := build.CombinedOutput(); err != nil {
			panic("building ./cmd/" + cmd + ": " + err.Error() + "\n" + string(output))
		}
		binaries[cmd] = out
	}
	os.Exit(m.Run())
}

func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}

// run executes a command from the repository root and returns its streams and
// exit status.
func run(t *testing.T, name string) (stdout, stderr string, exitCode int) {
	t.Helper()
	cmd := exec.Command(binaries[name])
	cmd.Dir = repoRoot
	var out, errOut strings.Builder
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	err := cmd.Run()
	if err != nil {
		var exit *exec.ExitError
		if !asExitError(err, &exit) {
			t.Fatalf("running %s: %v", name, err)
		}
		exitCode = exit.ExitCode()
	}
	return out.String(), errOut.String(), exitCode
}

func asExitError(err error, out **exec.ExitError) bool {
	e, ok := err.(*exec.ExitError)
	if ok {
		*out = e
	}
	return ok
}

// TestAopBootstrap covers the AOP demonstration end to end: the interleaving
// of interceptor and target output, and that the last line names a generated
// proxy rather than the target class.
func TestAopBootstrap(t *testing.T) {
	stdout, stderr, code := run(t, "aop")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr:\n%s", code, stderr)
	}
	lines := strings.Split(strings.TrimSuffix(stdout, "\n"), "\n")
	want := []string{
		"SimpleMethodInterceptor被调用: testB",
		"testB执行",
		"SimpleMethodInterceptor被调用: testC",
		"testC执行",
	}
	if len(lines) != len(want)+1 {
		t.Fatalf("got %d lines, want %d:\n%s", len(lines), len(want)+1, stdout)
	}
	for i, w := range want {
		if lines[i] != w {
			t.Errorf("line %d = %q, want %q", i, lines[i], w)
		}
	}
	if !strings.HasPrefix(lines[4], "SimpleAopBean$$") {
		t.Errorf("last line = %q, want a generated proxy class name", lines[4])
	}
}

// TestBaseBootstrap covers the demonstration that fails: config.xml declares
// no TransactionBean, so the lookup fails, names the type, and exits 1.
func TestBaseBootstrap(t *testing.T) {
	stdout, stderr, code := run(t, "base")
	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want nothing", stdout)
	}
	want := "No qualifying bean of type 'base.transaction.TransactionBean' available"
	if !strings.Contains(stderr, want) {
		t.Errorf("stderr = %q, want it to contain %q", stderr, want)
	}
}

// TestJavaConfigBootstrap covers the annotation-configuration demonstration,
// including the bean-definition registry it prints.
func TestJavaConfigBootstrap(t *testing.T) {
	stdout, stderr, code := run(t, "javaconfig")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr:\n%s", code, stderr)
	}
	want := strings.Join([]string{
		"importaware",
		"skywalker",
		"[org.springframework.context.annotation.internalConfigurationAnnotationProcessor, " +
			"org.springframework.context.annotation.internalAutowiredAnnotationProcessor, " +
			"org.springframework.context.annotation.internalRequiredAnnotationProcessor, " +
			"org.springframework.context.annotation.internalCommonAnnotationProcessor, " +
			"org.springframework.context.event.internalEventListenerProcessor, " +
			"org.springframework.context.event.internalEventListenerFactory, " +
			"simpleBeanConfig, java_config.StudentConfig, student, simpleBean]",
	}, "\n") + "\n"
	if stdout != want {
		t.Errorf("stdout:\ngot:\n%s\nwant:\n%s", stdout, want)
	}
}

// TestJDKProxy covers the dynamic-proxy demonstration, including the absence
// of a report for the target's self-call.
func TestJDKProxy(t *testing.T) {
	stdout, stderr, code := run(t, "jdkproxy")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr:\n%s", code, stderr)
	}
	want := "Method printName is proxyed.\nName is XXX\nAge: 18\n"
	if stdout != want {
		t.Errorf("stdout:\ngot:\n%q\nwant:\n%q", stdout, want)
	}
}

// TestJavaTest covers the reflection demonstration.
func TestJavaTest(t *testing.T) {
	stdout, stderr, code := run(t, "javatest")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr:\n%s", code, stderr)
	}
	if stdout != "name: Get, return: string\n" {
		t.Errorf("stdout = %q", stdout)
	}
}
