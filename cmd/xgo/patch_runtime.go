package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/xgo/cmd/xgo/patch"
	"github.com/xhd2015/xgo/support/filecopy"
	"github.com/xhd2015/xgo/support/goinfo"
)

var xgoAutoGenRegisterFuncHelper = _FilePath{"src", "runtime", "__xgo_autogen_register_func_helper.go"}
var xgoTrap = _FilePath{"src", "runtime", "xgo_trap.go"}
var runtimeProc = _FilePath{"src", "runtime", "proc.go"}
var runtimeTime _FilePath = _FilePath{"src", "runtime", "time.go"}
var timeSleep _FilePath = _FilePath{"src", "time", "sleep.go"}

var testingFilePatch = &FilePatch{
	FilePath: _FilePath{"src", "testing", "testing.go"},
	Patches: []*Patch{
		{
			Mark:         "declare_testing_callback_v2",
			InsertIndex:  0,
			InsertBefore: true,
			Anchors: []string{
				"func tRunner(t *T, fn func",
				"{",
				"\n",
			},
			Content: patch.TestingCallbackDeclarations,
		},
		{
			Mark:         "call_testing_callback_v2",
			InsertIndex:  4,
			InsertBefore: true,
			Anchors: []string{
				"func tRunner(t *T, fn func",
				"{",
				"\n",
				`t.start = time.Now()`,
				"fn(t",
			},
			Content: patch.TestingStart,
		},
	},
}

var runtimeFiles = []_FilePath{
	xgoAutoGenRegisterFuncHelper,
	xgoTrap,
	runtimeProc,
	testingFilePatch.FilePath,
	runtimeTime,
	timeSleep,
}

func patchRuntimeAndTesting(goroot string, goVersion *goinfo.GoVersion) error {
	err := patchRuntimeProc(goroot, goVersion)
	if err != nil {
		return err
	}
	err = patchRuntimeTesting(goroot)
	if err != nil {
		return err
	}
	err = patchRuntimeTime(goroot)
	if err != nil {
		return err
	}
	return nil
}

// addRuntimeFunctions always copy file
func addRuntimeFunctions(goroot string, goVersion *goinfo.GoVersion, xgoSrc string) (updated bool, err error) {
	if false {
		// seems unnecessary
		// TODO: needs to debug to see what will happen with auto generated files
		// we need to skip when debugging

		// add debug file
		//   rational: when debugging, dlv will jump to __xgo_autogen_register_func_helper.go
		// previously this file does not exist, making the debugging blind
		runtimeAutoGenFile := filepath.Join(goroot, filepath.Join(xgoAutoGenRegisterFuncHelper...))
		srcAutoGen := getInternalPatch(goroot, "syntax", "helper_code.go")
		err = filecopy.CopyFile(srcAutoGen, runtimeAutoGenFile)
		if err != nil {
			return false, err
		}
	}

	dstFile := filepath.Join(goroot, filepath.Join(xgoTrap...))
	content, err := readXgoSrc(xgoSrc, []string{"trap_runtime", "xgo_trap.go"})
	if err != nil {
		return false, err
	}

	content, err = replaceBuildIgnore(content)
	if err != nil {
		return false, fmt.Errorf("file %s: %w", filepath.Base(dstFile), err)
	}

	// the func.entry is a field, not a function
	if goVersion.Major == 1 && goVersion.Minor <= 17 {
		entryPatch := "fn.entry() /*>=go1.18*/"
		entryPatchBytes := []byte(entryPatch)
		idx := bytes.Index(content, entryPatchBytes)
		if idx < 0 {
			return false, fmt.Errorf("expect %q in xgo_trap.go, actually not found", entryPatch)
		}
		content = bytes.ReplaceAll(content, entryPatchBytes, []byte("fn.entry"))
	}

	// func name patch
	if goVersion.Major > 1 || goVersion.Minor > 22 {
		panic("should check the implementation of runtime.FuncForPC(pc).Name() to ensure __xgo_get_pc_name is not wrapped in print format above go1.22")
	}
	if goVersion.Major > 1 || goVersion.Minor >= 21 {
		content = append(content, []byte(patch.RuntimeGetFuncName_Go121)...)
	} else if goVersion.Major == 1 {
		if goVersion.Minor >= 17 {
			// go1.17,go1.18,go1.19
			content = append(content, []byte(patch.RuntimeGetFuncName_Go117_120)...)
		}
	}

	return true, os.WriteFile(dstFile, content, 0755)
}

func patchRuntimeProc(goroot string, goVersion *goinfo.GoVersion) error {
	procFile := filepath.Join(goroot, filepath.Join(runtimeProc...))
	anchors := []string{
		"func main() {",
		"doInit(", "runtime_inittask", ")", // first doInit for runtime
		"doInit(", // second init for main
		"close(main_init_done)",
		"\n",
	}
	err := editFile(procFile, func(content string) (string, error) {
		content = addContentAfter(content, "/*<begin set_init_finished_mark>*/", "/*<end set_init_finished_mark>*/", anchors, patch.RuntimeProcPatch)

		// goexit1() is called for every exited goroutine
		content = addContentAfter(content,
			"/*<begin add_go_exit_callback>*/", "/*<end add_go_exit_callback>*/",
			[]string{"func goexit1() {", "\n"},
			patch.RuntimeProcGoroutineExitPatch,
		)

		procDecl := `func newproc(fn`
		newProc := `newg := newproc1(fn, gp, pc)`
		if goVersion.Major == 1 && goVersion.Minor <= 17 {
			procDecl = `func newproc(siz int32`
			newProc = `newg := newproc1(fn, argp, siz, gp, pc)`
		}

		// see https://github.com/xhd2015/xgo/issues/67
		content = addContentAtIndex(
			content,
			"/*<begin declare_xgo_newg>*/", "/*<end declare_xgo_newg>*/",
			[]string{
				procDecl,
				`systemstack(func() {`,
				newProc,
			},
			1,
			true,
			"var xgo_newg *g",
		)
		content = addContentAtIndex(
			content,
			"/*<begin set_xgo_newg>*/", "/*<end set_xgo_newg>*/",
			[]string{
				procDecl,
				`systemstack(func() {`,
				newProc,
				"\n",
			},
			3,
			false,
			"xgo_newg = newg",
		)

		content = addContentAtIndex(content,
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
			true,
			patch.RuntimeProcGoroutineCreatedPatch,
		)
		return content, nil
	})
	if err != nil {
		return err
	}
	return nil
}

func patchRuntimeTesting(goroot string) error {
	return testingFilePatch.Apply(goroot, nil)
}

// only required if need to mock time.Sleep
func patchRuntimeTime(goroot string) error {
	runtimeTimeFile := filepath.Join(goroot, filepath.Join(runtimeTime...))
	timeSleepFile := filepath.Join(goroot, filepath.Join(timeSleep...))

	err := editFile(runtimeTimeFile, func(content string) (string, error) {
		content = replaceContentAfter(content,
			"/*<begin redirect_runtime_sleep>*/", "/*<end redirect_runtime_sleep>*/",
			[]string{},
			"//go:linkname timeSleep time.Sleep\nfunc timeSleep(ns int64) {",
			"//go:linkname timeSleep time.runtimeSleep\nfunc timeSleep(ns int64) {",
		)
		return content, nil
	})
	if err != nil {
		return err
	}

	err = editFile(timeSleepFile, func(content string) (string, error) {
		content = replaceContentAfter(content,
			"/*<begin replace_sleep_with_runtimesleep>*/", "/*<end replace_sleep_with_runtimesleep>*/",
			[]string{},
			"func Sleep(d Duration)",
			strings.Join([]string{
				"func runtimeSleep(d Duration)",
				"func Sleep(d Duration){",
				"    runtimeSleep(d)",
				"}",
			}, "\n"),
		)
		return content, nil
	})
	if err != nil {
		return err
	}

	return nil
}
