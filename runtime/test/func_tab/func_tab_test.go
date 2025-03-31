package functab

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/functab"
)

var SomeVar int

var SomeVarWithoutType = 0

func Example() {
	var _ = &SomeVar
	var _ = &SomeVarWithoutType
}

type SomeType struct {
}

func (s SomeType) ValueMethod() {

}

func (s *SomeType) PtrMethod() {

}

func TestFuncTab(t *testing.T) {
	getPC := func(fn interface{}) uintptr {
		return reflect.ValueOf(fn).Pointer()
	}
	type TestCase struct {
		WantKind     core.Kind
		WantFullName string
		WantPC       uintptr
		WantVar      interface{}
	}
	expectFullNames := []TestCase{
		{
			WantKind:     core.Kind_Var,
			WantFullName: "github.com/xhd2015/xgo/runtime/test/func_tab.SomeVar",
			WantVar:      &SomeVar,
		},
		{
			WantKind:     core.Kind_Func,
			WantFullName: "github.com/xhd2015/xgo/runtime/test/func_tab.Example",
			WantPC:       getPC(Example),
		},
		{
			WantKind:     core.Kind_Func,
			WantFullName: "github.com/xhd2015/xgo/runtime/test/func_tab.SomeType.ValueMethod",
			WantPC:       getPC(SomeType.ValueMethod),
		},
		{
			WantKind:     core.Kind_Func,
			WantFullName: "github.com/xhd2015/xgo/runtime/test/func_tab.(*SomeType).PtrMethod",
			WantPC:       getPC((*SomeType).PtrMethod),
		},
		{
			WantKind:     core.Kind_Func,
			WantFullName: "github.com/xhd2015/xgo/runtime/test/func_tab.TestFuncTab",
			WantPC:       getPC(TestFuncTab),
		},
	}
	funcInfos := functab.GetFuncs()
	if len(funcInfos) != len(expectFullNames) {
		t.Errorf("funcInfos length mismatch: %d != %d", len(funcInfos), len(expectFullNames))
	}
	for i, funcInfo := range funcInfos {
		var expectKind core.Kind
		var expectFullName string
		var expectPC uintptr
		var expectVar interface{}
		if i < len(expectFullNames) {
			expectKind = expectFullNames[i].WantKind
			expectFullName = expectFullNames[i].WantFullName
			expectPC = expectFullNames[i].WantPC
			expectVar = expectFullNames[i].WantVar
		}
		if funcInfo.Kind != expectKind {
			t.Errorf("funcInfo[%d] kind mismatch, want %s, got %s", i, expectKind, funcInfo.Kind)
		}
		if funcInfo.FullName != expectFullName {
			t.Errorf("funcInfo[%d] mismatch, want %s, got %s", i, expectFullName, funcInfo.FullName)
		}
		if funcInfo.PC != expectPC {
			t.Errorf("funcInfo[%d] pc mismatch, want %d, got %d", i, expectPC, funcInfo.PC)
		}
		if expectVar != nil && funcInfo.Var != expectVar {
			t.Errorf("funcInfo[%d] var mismatch, want %v, got %v", i, expectVar, funcInfo.Var)
		}
	}

	funcInfo := functab.InfoFunc(Example)
	if funcInfo == nil {
		t.Fatal(fmt.Errorf("func not found"))
	}
	funcInfo2 := functab.InfoPC(getPC(Example))
	if funcInfo2 == nil {
		t.Fatal(fmt.Errorf("func not found"))
	}
	if funcInfo2.PC != funcInfo.PC {
		t.Errorf("funcInfo pc mismatch, want %d, got %d", funcInfo.PC, funcInfo2.PC)
	}

	varInfo := functab.InfoVar(&SomeVar)
	if varInfo == nil {
		t.Fatal(fmt.Errorf("var not found"))
	}
	if varInfo.Var != &SomeVar {
		t.Errorf("varInfo var mismatch, want %v, got %v", &SomeVar, varInfo.Var)
	}
}
