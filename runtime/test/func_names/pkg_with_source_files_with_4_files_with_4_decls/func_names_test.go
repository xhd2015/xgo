package pkg_with_source_files_with_4_files_with_4_decls

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
//   GoFiles[0]     = funcs.go                          → index 0
//   TestGoFiles[0] = closure_name_*_test.go (build-tagged)  → index 1 (no stubs, file has no trap targets)
//   TestGoFiles[1] = func_names_test.go (this file)     → index 2
//
// Thus testFileStubIndex = 2. If a new file is added BEFORE this file in
// the GoFiles or TestGoFiles list, this index shifts. To recalculate:
//
//   1. Run `ls *.go | wc -l` to count total source files (N)
//   2. The first GoFile is index 0; this file's index is its position
//      in the concatenated GoFiles + TestGoFiles list
//   3. Update testFileStubIndex accordingly
//
// Each file (regardless of build tag) occupies a position, even if that
// file is excluded by build constraints on the current Go version.
// const testFileStubIndex = 2

var topClosure = func() {}

// TestStubNames verifies that:
//   - Compiler autogen stubs (index 0 from funcs.go) are present
//   - Overlay-injected file stubs (this file's index) are present
//   - The top-level closure has the version-correct name
func TestStubNames(t *testing.T) {
	// Compiler autogen stubs (funcs.go, index 0)
	assertName(t, __xgo_register_0, "runtime.XgoRegister")
	assertName(t, __xgo_trap_0, "runtime.XgoTrap")
	assertName(t, __xgo_trap_var_0, "runtime.XgoTrapVar")
	assertName(t, __xgo_trap_varptr_0, "runtime.XgoTrapVarPtr")

	// Overlay-injected file stubs (this file, index = testFileStubIndex)
	// Using __xgo_register_2 because testFileStubIndex = 2
	assertName(t, __xgo_register_2, "runtime.XgoRegister")
	assertName(t, __xgo_trap_2, "runtime.XgoTrap")
	assertName(t, __xgo_trap_var_2, "runtime.XgoTrapVar")
	assertName(t, __xgo_trap_varptr_2, "runtime.XgoTrapVarPtr")

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
