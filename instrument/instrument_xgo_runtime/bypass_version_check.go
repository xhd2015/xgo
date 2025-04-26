package instrument_xgo_runtime

import (
	"github.com/xhd2015/xgo/instrument/patch"
	"github.com/xhd2015/xgo/support/goinfo"
)

func checkBypassVersionCheck(versionCode string, runtimeVersion string) string {
	if goinfo.CompareSemVer("v"+runtimeVersion, "v1.1.4") < 0 {
		// xgo v1.1.2 and v1.1.4 has addressed several issues that
		// affect correctness. older v1.1.0 need to upgrade
		//
		// newer xgo can work with runtime v1.1.0/1.1.1/... with no trouble
		// NOTE: how to decide bypass or not?
		// you'd check the xgo/runtime code diff across these versions
		// and find a way to migrate from old version to new version
		// by instrumenting the code, which is basically enhancing
		// the code here
		// also add test under runtime/test/build
		//  - runtime/test/build/legacy_depend_1_0_52
		//  - runtime/test/build/legacy_depend_1_1_0
		//  - runtime/test/build/legacy_depend_1_1_1
		versionCode = bypassVersionCheck(versionCode)
	}

	return versionCode
}

func bypassVersionCheck(versionCode string) string {
	return patch.UpdateContent(versionCode,
		"/*<begin bypass_version_check>*/",
		"/*<end bypass_version_check>*/",
		[]string{
			"func checkVersion() error {",
		},
		0,
		patch.UpdatePosition_After,
		"if true { return nil; }",
	)
}
