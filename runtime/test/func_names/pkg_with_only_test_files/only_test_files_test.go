package pkg_with_only_test_files

import (
	"reflect"
	"runtime"
	"testing"

	_ "github.com/xhd2015/xgo/runtime/trap"
)

var topClosure = func() {}

func ExportedFunc() {}

func TestStubNames(t *testing.T) {
	// With only one file in this package, file index is 0 for all stubs.
	assertName(t, __xgo_register_0, "runtime.XgoRegister")
	assertName(t, __xgo_trap_0, "runtime.XgoTrap")
	assertName(t, __xgo_trap_var_0, "runtime.XgoTrapVar")
	assertName(t, __xgo_trap_varptr_0, "runtime.XgoTrapVarPtr")
}

func assertName(t *testing.T, fn interface{}, expected string) {
	t.Helper()
	v := reflect.ValueOf(fn)
	name := runtime.FuncForPC(v.Pointer()).Name()
	if name != expected {
		t.Errorf("expect %s, actual: %s", expected, name)
	}
}
