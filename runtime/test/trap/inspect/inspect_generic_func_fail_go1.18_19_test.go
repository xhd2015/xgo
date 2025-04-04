//go:build go1.18 && !go1.20
// +build go1.18,!go1.20

package inspect

import (
	"fmt"
	"testing"

	"github.com/xhd2015/xgo/runtime/internal/trap"
)

func TestGenericFuncFailWithGo118_19(t *testing.T) {
	var panicErr interface{}
	func() {
		defer func() {
			panicErr = recover()
		}()
		trap.Inspect(GenericFunc[int])
	}()
	if panicErr == nil {
		t.Fatalf("expect panic, actual nil")
	}
	msg := fmt.Sprint(panicErr)
	expectMsg := "func not instrumented by xgo, see https://github.com/xhd2015/xgo/tree/master/doc/ERR_NOT_INSTRUMENTED.md: github.com/xhd2015/xgo/runtime/test/trap/inspect.TestGenericFuncFailWithGo118_19.func1.2"
	if msg != expectMsg {
		t.Fatalf("expect panic message %q, actual %q", expectMsg, msg)
	}
}

func TestInspectPCGenericGo118_19ShouldFail(t *testing.T) {
	var panicErr interface{}
	func() {
		defer func() {
			panicErr = recover()
		}()
		trap.Inspect(ToString[int])
	}()
	if panicErr == nil {
		t.Fatalf("expect panic, actual nil")
	}
	msg := fmt.Sprint(panicErr)
	expectMsg := "func not instrumented by xgo, see https://github.com/xhd2015/xgo/tree/master/doc/ERR_NOT_INSTRUMENTED.md: github.com/xhd2015/xgo/runtime/test/trap/inspect.TestInspectPCGenericGo118_19ShouldFail.func1.2"
	if msg != expectMsg {
		t.Fatalf("expect panic message %q, actual %q", expectMsg, msg)
	}
}
