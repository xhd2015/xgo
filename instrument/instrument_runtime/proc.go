package instrument_runtime

import (
	"strings"

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

	// Create callback runs off systemstack. Race context must be set up AFTER
	// the callback so parent writes to child __xgo_g happen-before racegostart
	// (see https://github.com/xhd2015/xgo/issues/341).
	//
	// The raceenabled block body differs by Go version:
	// - go1.17–1.18: racegostart only
	// - go1.19 early: racegostart + labels release
	// - go1.19.13+/1.20+: racegostart + raceignore + labels release
	// Detect which pieces exist in newproc1 so we re-emit a compiling block.
	raceSetupAfterCallback := buildRaceSetupAfterCreateCallback(procContent)
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
		";__xgo_newg=newg});__xgo_callback_on_create_g(__xgo_curg,__xgo_newg);systemstack(func(){newg:=__xgo_newg;"+raceSetupAfterCallback,
	)

	// Disable race setup inside newproc1 (moved after create callback above).
	// Anchor without relying on "// Set up race context." (added only in go1.22+).
	//
	// Anchoring note (legacy path; not upgraded further):
	// SequenceOffset takes the first "if raceenabled {" after "func newproc1(",
	// then requires "racegostart(" later. That is enough for stock Go, which has a
	// single raceenabled block owning racegostart. It is weaker than the file-based
	// patch (match racegostart → find_for_replace last preceding if raceenabled),
	// which survives an earlier decoy if raceenabled. Prefer that DSL if newproc1
	// gains multiple raceenabled blocks; this live path is outdated relative to it.
	// See instrument/patch apply_engine evalMatch(forReplace) and
	// TestApplyPatch_DeferRacegostartAnchorsOnRacegostart.
	procContent = patch.UpdateContent(procContent,
		"/*<begin xgo_proc_defer_racegostart>*/",
		"/*<end xgo_proc_defer_racegostart>*/",
		[]string{
			"func newproc1(",
			"if raceenabled {",
			"racegostart(",
		},
		1,
		patch.UpdatePosition_Replace,
		"if false { // xgo: race setup moved after create callback (#341)",
	)
	return procContent, nil
}

// buildRaceSetupAfterCreateCallback returns an inlined race-setup statement
// that matches the original newproc1 raceenabled block for this GOROOT.
func buildRaceSetupAfterCreateCallback(procContent string) string {
	// Minimal setup present since early race support.
	// pc is newproc's caller PC (GetCallerPC / getcallerpc).
	setup := "if raceenabled { newg.racectx = racegostart(pc);"
	// raceignore was added mid-go1.19 (present in go1.19.13+, not in go1.19.0).
	if containsInFunc(procContent, "func newproc1(", "newg.raceignore") {
		setup += " newg.raceignore = 0;"
	}
	// labels release present from go1.19+.
	if containsInFunc(procContent, "func newproc1(", "racereleasemergeg(newg") {
		setup += " if newg.labels != nil { racereleasemergeg(newg, unsafe.Pointer(&labelSync)) };"
	}
	setup += " };"
	return setup
}

// containsInFunc reports whether needle appears inside the Go function that
// starts at startMark (e.g. "func newproc1("). The scan stops at the next
// top-level "\nfunc " (or EOF), so later unrelated occurrences cannot flip
// feature detection for race setup emission.
func containsInFunc(content, startMark, needle string) bool {
	idx := strings.Index(content, startMark)
	if idx < 0 {
		return false
	}
	rest := content[idx:]
	// Skip past startMark so we do not match a nested/next "func " incorrectly.
	nextRel := strings.Index(rest[len(startMark):], "\nfunc ")
	if nextRel < 0 {
		return strings.Contains(rest, needle)
	}
	return strings.Contains(rest[:len(startMark)+nextRel], needle)
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
