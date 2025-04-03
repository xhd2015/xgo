//go:build !go1.18
// +build !go1.18

package main

import (
	"github.com/xhd2015/xgo/support/fileutil"
)

func patchJSONPretty(settingsFile string, fn func(settings *map[string]interface{}) error) error {
	init := func() interface{} {
		m := make(map[string]interface{})
		return &m
	}
	return fileutil.PatchJSONPretty(settingsFile, init, func(v interface{}) error {
		f := *(v.(*interface{}))
		return fn(f.(*map[string]interface{}))
	})
}
