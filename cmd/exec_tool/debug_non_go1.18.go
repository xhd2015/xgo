//go:build !go1.18
// +build !go1.18

package main

import (
	"github.com/xhd2015/xgo/support/fileutil"
)

func patchJSONPretty(settingsFile string, fn func(conf *VscodeLaunchConfigMap) error) error {
	init := func() interface{} {
		m := VscodeLaunchConfigMap{}
		return &m
	}
	return fileutil.PatchJSONPretty(settingsFile, init, func(v interface{}) error {
		f := *(v.(*interface{}))
		return fn(f.(*VscodeLaunchConfigMap))
	})
}
