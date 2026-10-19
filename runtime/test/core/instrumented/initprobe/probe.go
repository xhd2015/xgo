// Package initprobe checks linkage without importing mock or trap.
package initprobe

import "github.com/xhd2015/xgo/runtime/core"

var Instrumented = core.Instrumented()

func Untrapped() {}

func init() {
	if !Instrumented {
		panic("xgo linkage must be available before package init")
	}
}
