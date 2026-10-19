package mock_test

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/internal/runtime"
	"github.com/xhd2015/xgo/runtime/mock"
	"github.com/xhd2015/xgo/runtime/trace"
	"github.com/xhd2015/xgo/runtime/trap"
)

// Check during variable initialization, before this package's init functions.
var instrumentedAtInit = core.Instrumented()

type target struct{}

func (target) Method() {}

func TestNotInstrumented(t *testing.T) {
	if core.Instrumented() {
		t.Skip("plain-Go diagnostic test; xgo-built behavior is covered by runtime/test/core/instrumented")
	}
	if instrumentedAtInit {
		t.Fatal("plain go test must not report xgo runtime linkage during init")
	}
	fn := func() {}
	tests := []struct {
		name string
		call func()
	}{
		{"Patch", func() { mock.Patch(fn, fn) }},
		{"PatchVar", func() { mock.Patch(new(int), func() int { return 1 }) }},
		{"PatchByName", func() { mock.PatchByName("missing/package", "Fn", fn) }},
		{"PatchMethodByName", func() { mock.PatchMethodByName(target{}, "Method", fn) }},
		{"Mock", func() { mock.Mock(fn, nil) }},
		{"MockByName", func() { mock.MockByName("missing/package", "Fn", nil) }},
		{"MockMethodByName", func() { mock.MockMethodByName(target{}, "Method", nil) }},
		{"AddInterceptor", func() { trap.AddInterceptor(&trap.Interceptor{}) }},
		{"AddFuncInterceptor", func() { trap.AddFuncInterceptor(fn, &trap.Interceptor{}) }},
		{"Record", func() { trace.Record(fn, fn, nil) }},
		{"RecordCall", func() { trace.RecordCall(fn, fn) }},
		{"RecordResult", func() { trace.RecordResult(fn, fn) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if got := recover(); got != runtime.NotInstrumentedMessage {
					t.Fatalf("panic = %v, want actionable xgo build message", got)
				}
			}()
			tt.call()
		})
	}
}

func TestPlainGoPatchFailure(t *testing.T) {
	if core.Instrumented() {
		t.Skip("requires a plain-Go subprocess")
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestPlainGoPatchChild$")
	cmd.Env = append(os.Environ(), "XGO_PATCH_CHILD=1")
	out, err := cmd.CombinedOutput()
	if err == nil || !strings.Contains(string(out), "panic: "+runtime.NotInstrumentedMessage) {
		t.Fatalf("expected actionable panic, exit = %v, output:\n%s", err, out)
	}
	if strings.Contains(string(out), "nil pointer dereference") {
		t.Fatalf("unexpected deeper panic:\n%s", out)
	}
}

func TestPlainGoPatchChild(t *testing.T) {
	if os.Getenv("XGO_PATCH_CHILD") == "1" {
		mock.Patch(func() {}, func() {})
	}
}
