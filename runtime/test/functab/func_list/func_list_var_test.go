package func_list

import (
	"testing"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/functab"
	"github.com/xhd2015/xgo/runtime/internal/legacy"
	"github.com/xhd2015/xgo/runtime/test/functab/func_list/sub"
)

const a int = 10

var b struct{}

const subPkgPath = testPkgPath + "/sub"

func TestFuncListConst(t *testing.T) {
	fna := functab.Info(testPkgPath, "a")
	if fna != nil {
		t.Fatalf("expect a to be nil since xgo v1.1.0 drops supports for const, actual not nil")
	}
}

func TestFuncListVar(t *testing.T) {
	fnaPtr := functab.Info(testPkgPath, "*a")
	if fnaPtr != nil {
		t.Fatalf("expect aptr to be nil, actual: %v", fnaPtr)
	}

	_ = b // make b referenced so that it is instrumented
	fnb := functab.Info(testPkgPath, "b")
	if fnb.Kind != core.Kind_Var {
		t.Fatalf("expect b.Kind to be %v, actual: %v", core.Kind_Var, fnb.Kind)
	}

	fnbPtr := functab.Info(testPkgPath, "*b")
	if !legacy.V1_0_0 {
		if fnbPtr != nil {
			t.Fatalf("expect bptr to be nil, actual: %v", fnbPtr)
		}
	} else {
		if fnbPtr.Kind != core.Kind_VarPtr {
			t.Fatalf("expect bptr.Kind to be %v, actual: %v", core.Kind_VarPtr, fnbPtr.Kind)
		}
	}

}

var _ = sub.A

func TestFuncListSubPkgVar(t *testing.T) {
	_ = sub.A // make A referenced so that it is instrumented
	fnA := functab.Info(subPkgPath, "A")
	if fnA.Kind != core.Kind_Var {
		t.Fatalf("expect fnA.Kind to be %v, actual: %v", core.Kind_Var, fnA.Kind)
	}
}
