package main

import (
	"fmt"
	"testing"
)

func __xgo_link_on_test_start(fn func(t *testing.T, fn func(t *testing.T))) {
	panic("WARNING: failed to link __xgo_link_on_test_start.(xgo required)")
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
