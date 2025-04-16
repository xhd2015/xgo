//go:build go1.18
// +build go1.18

package service

import (
	"testing"

	"github.com/xhd2015/xgo/runtime/test/patch/patch_var/generic_return/third"
)

var Tree = third.MustBuild[int](1)

// this should build successfully
func TestGenericReturn(t *testing.T) {
	var _ = Tree
}
