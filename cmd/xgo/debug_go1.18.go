//go:build go1.18
// +build go1.18

package main

import "github.com/xhd2015/xgo/support/fileutil"

func patchJSONPretty(settingsFile string, fn func(settings *map[string]interface{}) error) error {
	return fileutil.PatchJSONPretty(settingsFile, fn)
}
