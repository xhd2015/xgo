package service

import (
	"fmt"
	"strings"
	"testing"

	"github.com/xhd2015/xgo/runtime/test/patch/patch_third_party_var/third/mix"
	"github.com/xhd2015/xgo/runtime/test/patch/patch_third_party_var/third/sub"
	"github.com/xhd2015/xgo/runtime/trace"
)

func TestPatchMapWithFuncShouldFail(t *testing.T) {
	var panicErr interface{}
	func() {
		defer func() {
			panicErr = recover()
		}()
		trace.RecordCall(&mix.MapFunc, func() {
			mix.MapFunc[sub.ID{}] = func() {}
		})
	}()
	_ = mix.MapFunc
	if panicErr == nil {
		t.Fatalf("expect panic, but no panic")
	}
	panicMsg := fmt.Sprint(panicErr)
	expectIncludeMsg := "variable not instrumented by xgo"
	if !strings.Contains(panicMsg, expectIncludeMsg) {
		t.Fatalf("expect panic message to include %v, but actual: %v", expectIncludeMsg, panicMsg)
	}
}

func TestPatchMapWithoutFuncShouldSuccess(t *testing.T) {
	var recorded bool
	trace.RecordCall(&mix.MapInt, func(map[sub.ID]int) {
		recorded = true
	})
	_ = mix.MapInt
	if !recorded {
		t.Fatalf("expect called, but not called")
	}
}
