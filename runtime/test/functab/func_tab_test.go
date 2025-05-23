package functab

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/functab"
)

var SomeVar int

// xgo can trap var without type
// if the var is defined as literal
var SomeVarWithoutType = 0

func SomeFunc(s string) {
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
		WantKind         core.Kind
		WantFullName     string
		WantName         string
		WantIdentityName string
		WantArgs         []string
		WantPC           uintptr
		WantVar          interface{}
	}
	expectFullNames := []TestCase{
		{
			WantKind:         core.Kind_Var,
			WantFullName:     "github.com/xhd2015/xgo/runtime/test/functab.SomeVar",
			WantName:         "SomeVar",
			WantIdentityName: "SomeVar",
			WantVar:          &SomeVar,
		},
		{
			WantKind:         core.Kind_Var,
			WantFullName:     "github.com/xhd2015/xgo/runtime/test/functab.SomeVarWithoutType",
			WantName:         "SomeVarWithoutType",
			WantIdentityName: "SomeVarWithoutType",
			WantVar:          &SomeVarWithoutType,
		},
		{
			WantKind:         core.Kind_Func,
			WantFullName:     "github.com/xhd2015/xgo/runtime/test/functab.SomeFunc",
			WantName:         "SomeFunc",
			WantIdentityName: "SomeFunc",
			WantPC:           getPC(SomeFunc),
			WantArgs:         []string{"s"},
		},
		{
			WantKind:         core.Kind_Func,
			WantFullName:     "github.com/xhd2015/xgo/runtime/test/functab.SomeType.ValueMethod",
			WantName:         "ValueMethod",
			WantIdentityName: "SomeType.ValueMethod",
			WantPC:           getPC(SomeType.ValueMethod),
		},
		{
			WantKind:         core.Kind_Func,
			WantFullName:     "github.com/xhd2015/xgo/runtime/test/functab.(*SomeType).PtrMethod",
			WantName:         "PtrMethod",
			WantIdentityName: "(*SomeType).PtrMethod",
			WantPC:           getPC((*SomeType).PtrMethod),
		},
		{
			WantKind:         core.Kind_Func,
			WantFullName:     "github.com/xhd2015/xgo/runtime/test/functab.TestFuncTab",
			WantName:         "TestFuncTab",
			WantIdentityName: "TestFuncTab",
			WantPC:           getPC(TestFuncTab),
		},
	}
	allFuncInfos := functab.GetFuncs()
	funcInfos := make([]*core.FuncInfo, 0, len(allFuncInfos))
	for _, fnInfo := range allFuncInfos {
		// because we added --trap-stdlib=true,..., so filter them out
		if fnInfo.Stdlib {
			continue
		}
		funcInfos = append(funcInfos, fnInfo)
	}

	if len(funcInfos) != len(expectFullNames) {
		t.Errorf("funcInfos length mismatch: %d != %d", len(funcInfos), len(expectFullNames))
	}
	for i, funcInfo := range funcInfos {
		var expectKind core.Kind
		var expectFullName string
		var expectName string
		var expectIdentityName string
		var expectPC uintptr
		var expectVar interface{}
		var expectArgs []string
		if i < len(expectFullNames) {
			expectKind = expectFullNames[i].WantKind
			expectFullName = expectFullNames[i].WantFullName
			expectName = expectFullNames[i].WantName
			expectIdentityName = expectFullNames[i].WantIdentityName
			expectPC = expectFullNames[i].WantPC
			expectVar = expectFullNames[i].WantVar
			expectArgs = expectFullNames[i].WantArgs
		}
		if funcInfo.Kind != expectKind {
			t.Errorf("funcInfo[%d] kind mismatch, want %s, got %s", i, expectKind, funcInfo.Kind)
		}
		if funcInfo.FullName != expectFullName {
			t.Errorf("funcInfo[%d] fullName mismatch, want %s, got %s", i, expectFullName, funcInfo.FullName)
		}
		if funcInfo.Name != expectName {
			t.Errorf("funcInfo[%d] name mismatch, want %s, got %s", i, expectName, funcInfo.Name)
		}
		if funcInfo.IdentityName != expectIdentityName {
			t.Errorf("funcInfo[%d] identityName mismatch, want %s, got %s", i, expectIdentityName, funcInfo.IdentityName)
		}
		if funcInfo.PC != expectPC {
			t.Errorf("funcInfo[%d] pc mismatch, want %d, got %d", i, expectPC, funcInfo.PC)
		}
		if expectVar != nil && funcInfo.Var != expectVar {
			t.Errorf("funcInfo[%d] var mismatch, want %v, got %v", i, expectVar, funcInfo.Var)
		}
		if expectArgs != nil && !reflect.DeepEqual(expectArgs, funcInfo.ArgNames) {
			t.Errorf("funcInfo[%d] args mismatch, want %v, got %v", i, expectArgs, funcInfo.ArgNames)
		}
	}

	funcInfo := functab.InfoFunc(SomeFunc)
	if funcInfo == nil {
		t.Fatal(fmt.Errorf("func not found"))
	}
	funcInfo2 := functab.InfoPC(getPC(SomeFunc))
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
