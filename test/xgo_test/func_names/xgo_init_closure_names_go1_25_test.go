//go:build go1.25
// +build go1.25

package func_names

import (
	"reflect"
	"runtime"
	"testing"
)

// TestXgoInitClosureNames verifies that on go1.25, the xgo
// instrumentation (RegisterFuncTab overlay) shifts user top-level
// closures by 4 init slots — the 4 synthetic func init() blocks
// injected before user code consume init.func1-4, pushing var c
// to init.func5.
//
// When run via xgo test (overlay active), the 4 stub template
// closures are also present:
//
//	__xgo_register_1   → runtime.XgoRegister
//	__xgo_trap_1       → runtime.XgoTrap
//	__xgo_trap_var_1   → runtime.XgoTrapVar
//	__xgo_trap_varptr_1 → runtime.XgoTrapVarPtr
//
// These are NOT init closures — the init shift comes from the
// synthetic func init() blocks, not from these variable closures.
func TestXgoInitClosureNames(t *testing.T) {
	pkgPrefix := "github.com/xhd2015/xgo/test/xgo_test/func_names"
	assertName(t, c, pkgPrefix+".init.func5")
}

func assertName(t *testing.T, fn interface{}, expected string) {
	t.Helper()
	v := reflect.ValueOf(fn)
	name := runtime.FuncForPC(v.Pointer()).Name()
	if name != expected {
		t.Errorf("expect %s, actual: %s", expected, name)
	}
}
