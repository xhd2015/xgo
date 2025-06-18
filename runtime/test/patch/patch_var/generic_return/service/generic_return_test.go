//go:build go1.18
// +build go1.18

package service

import (
	"testing"

	"github.com/xhd2015/xgo/runtime/test/patch/patch_var/generic_return/third"
	"github.com/xhd2015/xgo/runtime/test/patch/patch_var/generic_return/third/fn2"
)

var Tree = third.MustBuild[int](1)
var Tree2 = fn2.MustBuild2[int, string](1, "data")

// this should build successfully
func TestGenericReturnSingleParam(t *testing.T) {
	var _ = Tree
}

// this should build successfully
func TestGenericReturnDoubleParam(t *testing.T) {
	var _ = Tree2
}
