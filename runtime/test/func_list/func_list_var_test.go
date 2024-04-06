package func_list

import (
	"testing"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/functab"
	"github.com/xhd2015/xgo/runtime/test/func_list/sub"
)

const a int = 10

var b struct{}

const subPkgPath = testPkgPath + "/sub"

func TestFuncListVar(t *testing.T) {
	fna := functab.Info(testPkgPath, "a")
	if fna.Kind != core.Kind_Const {
		t.Fatalf("expect a.Kind to be %v, actual: %v", core.Kind_Const, fna.Kind)
	}

	fnaPtr := functab.Info(testPkgPath, "*a")
	if fnaPtr != nil {
		t.Fatalf("expect aptr to be nil, actual: %v", fnaPtr)
	}

	fnb := functab.Info(testPkgPath, "b")
	if fnb.Kind != core.Kind_Var {
		t.Fatalf("expect b.Kind to be %v, actual: %v", core.Kind_Var, fnb.Kind)
	}

	fnbPtr := functab.Info(testPkgPath, "*b")
	if fnbPtr.Kind != core.Kind_VarPtr {
		t.Fatalf("expect bptr.Kind to be %v, actual: %v", core.Kind_VarPtr, fnbPtr.Kind)
	}
}

var _ = sub.A

func TestFuncListSubPkgVar(t *testing.T) {
	fnA := functab.Info(subPkgPath, "A")
	if fnA.Kind != core.Kind_Var {
		t.Fatalf("expect fnA.Kind to be %v, actual: %v", core.Kind_Var, fnA.Kind)
	}
}
