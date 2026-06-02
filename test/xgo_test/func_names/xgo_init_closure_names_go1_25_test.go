//go:build go1.25
// +build go1.25

package func_names

import (
	"reflect"
	"runtime"
	"testing"
)

// TestXgoInitClosureNames verifies that on go1.25, the xgo
// instrumentation (RegisterFuncTab overlay) correctly:
//
//  1. Shifts user top-level closures by 4 init slots — the 4
//     synthetic func init() blocks injected before user code
//     consume init.func1-4, pushing var c to init.func5.
//
//  2. Declares 4 stub template closures (__xgo_register_1,
//     __xgo_trap_1, __xgo_trap_var_1, __xgo_trap_varptr_1)
//     that reference runtime dispatch functions. These are
//     NOT init closures — the init shift comes from the
//     func init() blocks, not from these variable closures.
func TestXgoInitClosureNames(t *testing.T) {
	pkgPrefix := "github.com/xhd2015/xgo/test/xgo_test/func_names"

	// User top-level closure shifted to init.func5 (4 xgo func init()
	// blocks injected by RegisterFuncTab consume init.func1-4)
	assertName(t, c, pkgPrefix+".init.func5")

	// The 4 xgo stub variables are template closures that reference
	// runtime functions for trap dispatch. They are not init closures —
	// the init.func1-4 shift comes from synthetic func init() blocks.
	assertName(t, __xgo_register_1, "runtime.XgoRegister")
	assertName(t, __xgo_trap_1, "runtime.XgoTrap")
	assertName(t, __xgo_trap_var_1, "runtime.XgoTrapVar")
	assertName(t, __xgo_trap_varptr_1, "runtime.XgoTrapVarPtr")
}

func assertName(t *testing.T, fn interface{}, expected string) {
	t.Helper()
	v := reflect.ValueOf(fn)
	name := runtime.FuncForPC(v.Pointer()).Name()
	if name != expected {
		t.Errorf("expect %s, actual: %s", expected, name)
	}
}
