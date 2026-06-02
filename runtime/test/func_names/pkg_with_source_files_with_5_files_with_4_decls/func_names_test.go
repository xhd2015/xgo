package pkg_with_source_files_with_5_files_with_4_decls

import (
	"reflect"
	"runtime"
	"testing"

	_ "github.com/xhd2015/xgo/runtime/trap"
)

// testFileStubIndex is the file index that xgo's overlay assigns to this
// test file's stub variables (__xgo_register_N, __xgo_trap_N, etc.).
//
// The index is the zero-based position of this file in the package's file
// list: GoFiles (non-test) come first, followed by TestGoFiles. In this
// package:
//
//   GoFiles[0]     = aaa_placeholder.go                   → index 0 (no stubs, no trap-worthy declarations)
//   GoFiles[1]     = funcs.go                              → index 1
//   TestGoFiles[0] = closure_name_*_test.go (build-tagged) → index 2 (no stubs, file has no trap targets)
//   TestGoFiles[1] = func_names_test.go (this file)        → index 3
//
// Thus testFileStubIndex = 3. If a new file is added BEFORE this file in
// the GoFiles or TestGoFiles list, this index shifts. To recalculate:
//
//   1. Run `ls *.go | wc -l` to count total source files (N)
//   2. The first GoFile is index 0; this file's index is its position
//      in the concatenated GoFiles + TestGoFiles list
//   3. Update testFileStubIndex accordingly
//
// Each file (regardless of build tag) occupies a position, even if that
// file is excluded by build constraints on the current Go version.
// const testFileStubIndex = 3

var topClosure = func() {}

// TestStubNames verifies that:
//   - Compiler autogen stubs for funcs.go (index 1) are present
//   - Overlay-injected file stubs for this file (index 3) are present
//   - The top-level closure has the version-correct name (func5 — unchanged
//     because aaa_placeholder.go has no trap targets, so no stub closures
//     are injected for it)
func TestStubNames(t *testing.T) {
	// funcs.go stubs, index 1 (was 0 in the 4-files variant)
	assertName(t, __xgo_register_1, "runtime.XgoRegister")
	assertName(t, __xgo_trap_1, "runtime.XgoTrap")
	assertName(t, __xgo_trap_var_1, "runtime.XgoTrapVar")
	assertName(t, __xgo_trap_varptr_1, "runtime.XgoTrapVarPtr")

	// this file's stubs, index = testFileStubIndex = 3 (was 2 in the 4-files variant)
	assertName(t, __xgo_register_3, "runtime.XgoRegister")
	assertName(t, __xgo_trap_3, "runtime.XgoTrap")
	assertName(t, __xgo_trap_var_3, "runtime.XgoTrapVar")
	assertName(t, __xgo_trap_varptr_3, "runtime.XgoTrapVarPtr")

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
