package instrument_xgo_runtime

import "strings"

func checkBypassVersionCheck(versionCode string, coreVersion string) string {
	if coreVersion == "1.1.2" {
		// xgo v1.1.2 has addressed the several issues that
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
	return strings.Replace(versionCode,
		"func checkVersion() error {",
		"func checkVersion() error { if true { return nil; }",
		1,
	)
}
