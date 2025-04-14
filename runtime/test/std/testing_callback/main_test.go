package main

import (
	"fmt"
	"testing"

	"github.com/xhd2015/xgo/runtime/internal/legacy"
)

func __xgo_link_on_test_start(fn func(t *testing.T, fn func(t *testing.T))) {
	if !legacy.V1_0_0 {
		return
	}
	panic("WARNING: failed to link __xgo_link_on_test_start(requires xgo).")
	// link by compiler
}
func init() {
	__xgo_link_on_test_start(func(t *testing.T, fn func(t *testing.T)) {
		fmt.Printf("TEST STARTED: %v\n", t.Name())
	})
}
func TestExample(t *testing.T) {
	t.Logf("TEST EXAMPLE")
}
