// func names are taken from
// runtime's symbol section,
// they may not support in gcflags="all=-N -l" mode

package main

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"unsafe"
)

func __xgo_link_get_pc_name(pc uintptr) string {
	fmt.Fprintf(os.Stderr, "WARNING: failed to link __xgo_link_get_pc_name(requires xgo).\n")
	return ""
}

func F() {}

// NOTE: this test is expected be run in
// both normal and debug mode(i.e. gcflags="all=-N -l")
// go run ./cmd/xgo test -run TestFuncNameSlice -v ./test/xgo_test/func_name_slice
// go run ./cmd/xgo test -run TestFuncNameSlice -v -gcflags="all=-N -l" ./test/xgo_test/func_name_slice
func TestFuncNameSlice(t *testing.T) {
	type _func struct {
		pc uintptr
	}
	f := F
	px := (**_func)(unsafe.Pointer(&f))
	x := *px

	// t.Logf("pc=0x%x", x.pc)

	expectFuncName := "github.com/xhd2015/xgo/test/xgo_test/func_name_slice.F"
	funcName := __xgo_link_get_pc_name(x.pc)

	if funcName != expectFuncName {
		t.Fatalf("expect func name to be %s, actual: %s", expectFuncName, funcName)
	}

	idx := strings.LastIndex(funcName, ".")
	if idx < 0 {
		t.Fatalf("expect func name contains ., actula not found: %s", funcName)
	}

	pkgPath := funcName[:idx]
	expectPkgPath := "github.com/xhd2015/xgo/test/xgo_test/func_name_slice"
	if pkgPath != expectPkgPath {
		t.Fatalf("expect pkg: %s, actual: %s", expectPkgPath, pkgPath)
	}
}
