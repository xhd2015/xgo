package trap_set

import (
	"testing"
)

func __xgo_link_set_trap(trapImpl func(pkgPath string, identityName string, generic bool, pc uintptr, recv interface{}, args []interface{}, results []interface{}) (func(), bool)) {
	panic("WARNING: failed to link __xgo_link_set_trap.(xgo required)")
}

var haveCalledTrap bool

// go run ./cmd/xgo test -v -run TestTrapSet ./test/xgo_test/trap_set
func TestTrapSet(t *testing.T) {
	// must avoid infinite loop
	__xgo_link_set_trap(example_trap__xgo_trap_skip)
	run()

	if !haveCalledTrap {
		t.Fatalf("expect have called trap, actually not called")
	}
}

func run() {
}

// NOTE: must be skipped by the compiler
func example_trap__xgo_trap_skip(pkgPath, identityName string, generic bool, pc uintptr, recv interface{}, args, results []interface{}) (func(), bool) {
	haveCalledTrap = true
	return nil, false
}
