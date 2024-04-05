//go:build go1.18
// +build go1.18

package exec_tool

import (
	"github.com/xhd2015/xgo/support/fileutil"
)

func patchJSONPretty(settingsFile string, fn func(conf *VscodeLaunchConfigMap) error) error {
	return fileutil.PatchJSONPretty(settingsFile, fn)
}
