package trace_runtime

import (
	"fmt"
	"os"
	"unsafe"
)

func Runtime_XgoSetTrap(trap func(recvName string, recvPtr interface{}, argNames []string, args []interface{}, resultNames []string, results []interface{}) (func(), bool)) {
	fmt.Fprintln(os.Stderr, "WARNING: failed to link Runtime_XgoSetTrap(requires xgo).")
}

func Runtime_XgoGetCurG() unsafe.Pointer {
	fmt.Fprintln(os.Stderr, "WARNING: failed to link Runtime_XgoGetCurG(requires xgo).")
	return nil
}

func Runtime_XgoPeekPanic() interface{} {
	fmt.Fprintln(os.Stderr, "WARNING: failed to link Runtime_XgoPeekPanic(requires xgo).")
	return nil
}

func Runtime_XgoGetFullPCName(pc uintptr) string {
	fmt.Fprintln(os.Stderr, "WARNING: failed to link Runtime_XgoGetFullPCName(requires xgo).")
	return ""
}
