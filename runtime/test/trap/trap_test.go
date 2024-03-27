package trap

import (
	"context"
	"fmt"
	"testing"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/trap"
)

var hasTrapA bool
var hasCalledA bool
var hasAbortB bool
var hasCalledB bool

func init() {
	trap.AddInterceptor(&trap.Interceptor{
		Pre: func(ctx context.Context, f *core.FuncInfo, args core.Object, results core.Object) (interface{}, error) {
			trap.Skip()
			if f.Name == "A" {
				hasTrapA = true
				return nil, nil
			}
			if f.Name == "B" {
				hasAbortB = true
				return nil, trap.ErrAbort
			}
			return nil, nil
		},
	})
}

// go run ./script/run-test/ --include go1.17.13 --xgo-runtime-test-only -run TestTrap -v ./test/trap
func TestTrap(t *testing.T) {
	run()
	if !hasCalledA {
		t.Fatalf("expect hasCalledA, actually not called")
	}
	if !hasTrapA {
		t.Fatalf("expect hasTrapA, actually not set")
	}
	if !hasAbortB {
		t.Fatalf("expect hasAbortB, actually not set")
	}
	if hasCalledB {
		t.Fatalf("expect not hasCalledB, actually called")
	}
}

func run() {
	A()
	B()
}

func A() {
	hasCalledA = true
	fmt.Printf("A\n")
}

func B() {
	hasCalledB = true
	fmt.Printf("B\n")
}
