package instrument_runtime

import (
	"github.com/xhd2015/xgo/instrument/instrument_runtime/template"
	"github.com/xhd2015/xgo/instrument/patch"
	"github.com/xhd2015/xgo/support/goinfo"
)

func instrumentProc(goroot string, goVersion *goinfo.GoVersion) error {
	file := procPath.JoinPrefix(goroot)

	return patch.EditFile(file, func(content string) (string, error) {
		procContent, err := instrumentNewGroutineV2(goVersion, content)
		if err != nil {
			return "", err
		}
		procContent, err = instrumentGoexit(goVersion, procContent)
		if err != nil {
			return "", err
		}
		procContent = instrumentInitFinished(procContent)
		return procContent, nil
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
		"__xgo_curg := gp.m.curg;var __xgo_newg *g;",
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
		// avoid executing xgo code on system stack
		";__xgo_newg=newg});__xgo_callback_on_create_g(__xgo_curg,__xgo_newg);systemstack(func(){newg:=__xgo_newg;",
	)
	return procContent, nil
}

func instrumentInitFinished(content string) string {
	content = patch.UpdateContent(content,
		"/*<begin set_init_finished_mark>*/",
		"/*<end set_init_finished_mark>*/",
		[]string{
			"func main() {",
			"doInit(",
			"runtime_inittask",
			")",       // first doInit for runtime
			"doInit(", // second init for main
			"close(main_init_done)",
			"\n",
		},
		5,
		patch.UpdatePosition_Before,
		"__xgo_callback_on_init_finished();",
	)
	return content
}

func instrumentGoexit(goVersion *goinfo.GoVersion, procContent string) (string, error) {
	// goexit1() is called for every exited goroutine
	procContent = patch.UpdateContent(procContent,
		"/*<begin add_go_exit1_callback>*/", "/*<end add_go_exit1_callback>*/",
		[]string{
			"func goexit1() {",
			"\n",
		},
		0,
		patch.UpdatePosition_After,
		";__xgo_callback_on_exit_g()",
	)
	return procContent, nil
}

// InstrumentGoroutineCreation instruments newproc, which is
// called when `go func(){}` is executed
// `procContent` is the content of runtime/proc.go
// Deprecated
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
		template.RuntimeProcGoroutineCreatedPatch,
	)
	return procContent, nil
}
