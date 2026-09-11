package aop_test

import (
	"strings"
	"testing"

	demo "github.com/seaswalker/spring-analysis/internal/base/aop"
	"github.com/seaswalker/spring-analysis/internal/base/aop/annotation"
	springaop "github.com/seaswalker/spring-analysis/internal/deps/springframework/aop"
	"github.com/seaswalker/spring-analysis/internal/testsupport"
)

// TestAopDemoMessages covers the target's three methods.
func TestAopDemoMessages(t *testing.T) {
	d := demo.NewAopDemo()
	cases := []struct {
		call func()
		want string
	}{
		{d.Send, "send from aopdemo\n"},
		{d.Receive, "receive from aopdemo\n"},
		{d.Inter, "inter\n"},
	}
	for _, tc := range cases {
		if got := testsupport.CaptureStdout(t, tc.call); got != tc.want {
			t.Errorf("printed %q, want %q", got, tc.want)
		}
	}
	var _ demo.AopDemoInter = d
}

// TestAopDemoAdviceMessages covers the four advice methods.
func TestAopDemoAdviceMessages(t *testing.T) {
	a := demo.NewAopDemoAdvice()
	cases := []struct {
		call func()
		want string
	}{
		{a.BeforeSend, "before send\n"},
		{a.AfterSend, "after send\n"},
		{a.BeforeReceive, "before receive\n"},
		{a.AfterReceive, "after receive\n"},
	}
	for _, tc := range cases {
		if got := testsupport.CaptureStdout(t, tc.call); got != tc.want {
			t.Errorf("printed %q, want %q", got, tc.want)
		}
	}
}

// TestAopDemoAdviceSuppressesTheTarget covers the interceptor that never
// proceeds — which is what the original's invoke does.
func TestAopDemoAdviceSuppressesTheTarget(t *testing.T) {
	target := demo.NewAopDemo()
	proxy := springaop.NewProxy(target, "base.aop.AopDemo", []springaop.Advisor{{
		Pointcut: springaop.MustParseExecution("execution(* base.aop.AopDemo.*(..))"),
		Advice:   demo.NewAopDemoAdvice(),
	}}, false)

	out := testsupport.CaptureStdout(t, func() {
		if _, err := proxy.Invoke("send"); err != nil {
			t.Errorf("Invoke: %v", err)
		}
	})
	if out != "" {
		t.Errorf("the target ran despite the advice not proceeding: %q", out)
	}
}

// TestAspectDemoAdvisesOnlySend covers the aspect's pointcut and its before
// advice, both carried over verbatim from the annotations.
func TestAspectDemoAdvisesOnlySend(t *testing.T) {
	aspect := annotation.NewAspectDemo()
	if got := aspect.BeforeSend().Expression(); got != annotation.PointcutExpression {
		t.Errorf("pointcut = %q, want %q", got, annotation.PointcutExpression)
	}
	if out := testsupport.CaptureStdout(t, aspect.Before); out != "send之前\n" {
		t.Errorf("Before printed %q", out)
	}

	target := demo.NewAopDemo()
	proxy := springaop.NewProxy(target, "base.aop.AopDemo", []springaop.Advisor{aspect.Advisor()}, false)

	sendOut := testsupport.Lines(testsupport.CaptureStdout(t, func() {
		if _, err := proxy.Invoke("send"); err != nil {
			t.Errorf("Invoke(send): %v", err)
		}
	}))
	want := []string{"send之前", "send from aopdemo"}
	if strings.Join(sendOut, "\n") != strings.Join(want, "\n") {
		t.Errorf("send: got %v, want %v", sendOut, want)
	}

	receiveOut := testsupport.Lines(testsupport.CaptureStdout(t, func() {
		if _, err := proxy.Invoke("receive"); err != nil {
			t.Errorf("Invoke(receive): %v", err)
		}
	}))
	if strings.Join(receiveOut, "\n") != "receive from aopdemo" {
		t.Errorf("receive was advised: %v", receiveOut)
	}
}
