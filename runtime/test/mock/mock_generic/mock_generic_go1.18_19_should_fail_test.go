//go:build go1.18 && !go1.20
// +build go1.18,!go1.20

package mock_generic

import (
	"context"
	"fmt"
	"testing"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/mock"
)

func TestMockGenericFuncGo11819ShouldFail(t *testing.T) {
	var panicErr interface{}
	func() {
		defer func() {
			panicErr = recover()
		}()
		mock.Mock(ToString[int], func(ctx context.Context, fn *core.FuncInfo, args, results core.Object) error {
			results.GetFieldIndex(0).Set("hello world")
			return nil
		})
	}()
	if panicErr == nil {
		t.Fatalf("expect panic, actual nil")
	}

	msg := fmt.Sprint(panicErr)
	expectMsg := "func not instrumented by xgo, see https://github.com/xhd2015/xgo/tree/master/doc/ERR_NOT_INSTRUMENTED.md: github.com/xhd2015/xgo/runtime/test/mock/mock_generic.TestMockGenericFuncGo11819ShouldFail.func1.3"
	if msg != expectMsg {
		t.Fatalf("expect panic message %q, actual %q", expectMsg, msg)
	}
}

func TestMockGenericMethodGo11819ShouldFail(t *testing.T) {
	var panicErr interface{}
	formatter := &Formatter[int]{prefix: 1}
	func() {
		defer func() {
			panicErr = recover()
		}()
		mock.Mock(formatter.Format, func(ctx context.Context, fn *core.FuncInfo, args, results core.Object) error {
			results.GetFieldIndex(0).Set("hello world")
			return nil
		})
	}()
	if panicErr == nil {
		t.Fatalf("expect panic, actual nil")
	}

	msg := fmt.Sprint(panicErr)
	expectMsg := "func not instrumented by xgo, see https://github.com/xhd2015/xgo/tree/master/doc/ERR_NOT_INSTRUMENTED.md: github.com/xhd2015/xgo/runtime/test/mock/mock_generic.TestMockGenericMethodGo11819ShouldFail.func1.3"
	if msg != expectMsg {
		t.Fatalf("expect panic message %q, actual %q", expectMsg, msg)
	}
}
