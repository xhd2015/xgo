package instrument_runtime

import (
	"github.com/xhd2015/xgo/support/fileutil"
	"github.com/xhd2015/xgo/support/goinfo"
	"github.com/xhd2015/xgo/support/instrument/patch"
)

func instrumentProc(goroot string, goVersion *goinfo.GoVersion) error {
	file := procPath.JoinPrefix(goroot)
	return fileutil.UpdateFile(file, func(content []byte) (bool, []byte, error) {
		procContent, err := instrumentNewGroutineV2(goVersion, string(content))
		if err != nil {
			return false, nil, err
		}
		return true, []byte(procContent), nil
	})
}

func getProcAnchor(goVersion *goinfo.GoVersion) (procDecl, newProc string) {
	procDecl = `func newproc(fn`
	newProc = `newg := newproc1(fn, gp, pc, false, waitReasonZero)`
	if goVersion.Major == goinfo.GO_MAJOR_1 {
		if goVersion.Minor <= goinfo.GO_VERSION_17 {
			// to bypass typo check
			const size = "s" + "i" + "z"
			procDecl = `func newproc(` + size + ` int32`
			newProc = `newg := newproc1(fn, argp, ` + size + `, gp, pc)`
		} else if goVersion.Minor <= goinfo.GO_VERSION_22 {
			newProc = `newg := newproc1(fn, gp, pc)`
		} else if goVersion.Minor <= goinfo.GO_VERSION_23 {
			newProc = `newg := newproc1(fn, gp, pc, false, waitReasonZero)`
		}
	}
	return procDecl, newProc
}

func instrumentNewGroutineV2(goVersion *goinfo.GoVersion, procContent string) (string, error) {
	procDecl, newProc := getProcAnchor(goVersion)
	// capture curg before switching to system goroutine
	procContent = patch.UpdateContent(procContent,
		"/*<begin declare_xgo_curg>*/",
		"/*<end declare_xgo_curg>*/",
		[]string{
			procDecl,
			`systemstack(func() {`,
			newProc,
			"\n",
			"})",
			"}",
		},
		1,
		patch.UpdatePosition_Before,
		"__xgo_curg := gp.m.curg;",
	)
	procContent = patch.UpdateContent(procContent,
		"/*<begin add_go_newproc_callback_v2>*/",
		"/*<end add_go_newproc_callback_v2>*/",
		[]string{
			procDecl,
			`systemstack(func() {`,
			newProc,
			"\n",
			"})",
			"}",
		},
		2,
		patch.UpdatePosition_After,
		";__xgo_callback_on_create_g(__xgo_curg,newg);",
	)
	return procContent, nil
}

// InstrumentGoroutineCreation instruments newproc, which is
// called when `go func(){}` is executed
// `procContent` is the content of runtime/proc.go
func InstrumentGoroutineCreation(goVersion *goinfo.GoVersion, procContent string) (string, error) {
	procDecl, newProc := getProcAnchor(goVersion)
	// see https://github.com/xhd2015/xgo/issues/67
	procContent = patch.UpdateContent(
		procContent,
		"/*<begin declare_xgo_newg>*/", "/*<end declare_xgo_newg>*/",
		[]string{
			procDecl,
			`systemstack(func() {`,
			newProc,
		},
		1,
		patch.UpdatePosition_Before,
		"var xgo_newg *g;",
	)
	procContent = patch.UpdateContentLines(
		procContent,
		"/*<begin set_xgo_newg>*/", "/*<end set_xgo_newg>*/",
		[]string{
			procDecl,
			`systemstack(func() {`,
			newProc,
			"\n",
		},
		3,
		patch.UpdatePosition_After,
		"xgo_newg = newg",
	)

	procContent = patch.UpdateContentLines(procContent,
		"/*<begin add_go_newproc_callback>*/", "/*<end add_go_newproc_callback>*/",
		[]string{
			procDecl,
			`systemstack(func() {`,
			newProc,
			"\n",
			"})",
			"}",
		},
		5,
		patch.UpdatePosition_Before,
		RuntimeProcGoroutineCreatedPatch,
	)
	return procContent, nil
}
