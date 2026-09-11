package main

import (
	"strings"
	"testing"

	"github.com/seaswalker/spring-analysis/internal/testsupport"
)

// TestRunFailsWithoutTheBean covers the command's recorded behaviour:
// config.xml declares no TransactionBean, so the lookup fails and names the
// type it could not find.
func TestRunFailsWithoutTheBean(t *testing.T) {
	testsupport.ChdirRepoRoot(t)
	var err error
	testsupport.CaptureStdout(t, func() { err = run() })
	if err == nil {
		t.Fatal("run succeeded; config.xml declares no TransactionBean")
	}
	want := "No qualifying bean of type 'base.transaction.TransactionBean' available"
	if err.Error() != want {
		t.Errorf("error = %q, want %q", err, want)
	}
	if strings.Contains(err.Error(), "\n") {
		t.Errorf("error spans several lines: %q", err)
	}
}
