package main

import (
	"flag"
	"os"
	"strings"
	"testing"
)

// TestRunReportsAnUnassemblableApplication covers the command's body up to the
// point where it would block on the listener: bad flags in, error out.
func TestRunReportsAnUnassemblableApplication(t *testing.T) {
	savedArgs, savedFlags := os.Args, flag.CommandLine
	defer func() { os.Args, flag.CommandLine = savedArgs, savedFlags }()

	flag.CommandLine = flag.NewFlagSet("webapp", flag.ContinueOnError)
	flag.CommandLine.SetOutput(os.Stderr)
	os.Args = []string{"webapp", "-port", "0", "-resources", t.TempDir(), "-views", t.TempDir()}

	err := run()
	if err == nil {
		t.Fatal("run succeeded with no configuration on the classpath")
	}
	if !strings.Contains(err.Error(), "spring-servlet.xml") {
		t.Errorf("error = %q, want it to name the missing configuration", err)
	}
}
