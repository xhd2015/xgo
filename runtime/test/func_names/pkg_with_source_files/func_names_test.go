package pkg_with_source_files

import (
	"reflect"
	"runtime"
	"testing"

	_ "github.com/xhd2015/xgo/runtime/trap"
)

var topClosure = func() {}

// TestStubNames verifies that with two files (funcs.go + this file),
// the test file stubs use file index 1, while the compiler autogen
// stubs use index 0. Both sets are present.
func TestStubNames(t *testing.T) {
	// Compiler autogen stubs (index 0)
	assertName(t, __xgo_register_0, "runtime.XgoRegister")
	assertName(t, __xgo_trap_0, "runtime.XgoTrap")
	assertName(t, __xgo_trap_var_0, "runtime.XgoTrapVar")
	assertName(t, __xgo_trap_varptr_0, "runtime.XgoTrapVarPtr")

	// Overlay-injected file stubs (index 1, since funcs.go gets index 0)
	assertName(t, __xgo_register_1, "runtime.XgoRegister")
	assertName(t, __xgo_trap_1, "runtime.XgoTrap")
	assertName(t, __xgo_trap_var_1, "runtime.XgoTrapVar")
	assertName(t, __xgo_trap_varptr_1, "runtime.XgoTrapVarPtr")

	assertName(t, topClosure, expectTopLevelClosure)
}

func assertName(t *testing.T, fn interface{}, expected string) {
	t.Helper()
	v := reflect.ValueOf(fn)
	name := runtime.FuncForPC(v.Pointer()).Name()
	if name != expected {
		t.Errorf("expect %s, actual: %s", expected, name)
	}
}
