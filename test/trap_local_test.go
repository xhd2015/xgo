package test

import (
	"testing"
)

// go test -run TestTrapLocal -v ./test
func TestTrapLocal(t *testing.T) {
	t.Parallel()
	origExpect := "hello\nhello\n"
	expectOut := "global trap: main\nglobal trap: A\nlocal trap from A: printHello\nglobal trap: printHello\nhello\nglobal trap: B\nlocal trap from B: printHello\nglobal trap: printHello\nhello\n"
	testTrap(t, "./testdata/trap_local", origExpect, expectOut)
}

// go test -run TestTrapGoroutineLocal -v ./test
func TestTrapGoroutineLocal(t *testing.T) {
	t.Parallel()
	origExpect := "printHello: goroutine\nmain: goroutine\n"
	expectOut := "call: goroutineStr\nprintHello: goroutine\nmain: goroutine\n"
	testTrap(t, "./testdata/goroutine_trap", origExpect, expectOut)
}

// go test -run TestTrapNestedShouldAvoidRecursive -v ./test
func TestTrapNestedShouldAvoidRecursive(t *testing.T) {
	t.Parallel()
	origExpect := "hello world\n"
	expectOut := "trap pre: hello\ncall from trap\nhello world\n"
	testTrap(t, "./testdata/trap_nested", origExpect, expectOut)
}
