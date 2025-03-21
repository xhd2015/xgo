package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/xgo/cmd/xgo/patch"
	"github.com/xhd2015/xgo/support/filecopy"
	"github.com/xhd2015/xgo/support/fileutil"
	"github.com/xhd2015/xgo/support/goinfo"
	instrument_patch "github.com/xhd2015/xgo/support/instrument/patch"
	ast_patch "github.com/xhd2015/xgo/support/transform/patch"
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
			Mark:           "declare_testing_callback_v2",
			InsertIndex:    0,
			UpdatePosition: true,
			Anchors: []string{
				"func tRunner(t *T, fn func",
				"{",
				"\n",
			},
			Content: patch.TestingCallbackDeclarations + patch.TestingEndCallbackDeclarations,
		},
		{
			Mark:           "call_testing_callback_v2",
			InsertIndex:    4,
			UpdatePosition: true,
			Anchors: []string{
				"func tRunner(t *T, fn func",
				"{",
				"\n",
				`t.start = time.Now()`,
				"fn(t",
			},
			Content: patch.TestingStart + patch.TestingEnd,
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

func patchRuntimeAndTesting(origGoroot string, goroot string, goVersion *goinfo.GoVersion) error {
	err := patchRuntimeProc(goroot, goVersion)
	if err != nil {
		return err
	}
	err = patchRuntimeTesting(origGoroot, goroot, goVersion)
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
	if goVersion.Major > GO_MAJOR_1 || goVersion.Minor > GO_VERSION_23 {
		panic("should check the implementation of runtime.FuncForPC(pc).Name() to ensure __xgo_get_pc_name is not wrapped in print format above go1.23,it is confirmed that in go1.21,go1.22 and go1.23 the name is wrapped in funcNameForPrint(...).")
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
	err := instrument_patch.EditFile(procFile, func(content string) (string, error) {
		content = instrument_patch.AddContentAfter(content, "/*<begin set_init_finished_mark>*/", "/*<end set_init_finished_mark>*/", anchors, patch.RuntimeProcPatch)

		// goexit1() is called for every exited goroutine
		content = instrument_patch.AddContentAfter(content,
			"/*<begin add_go_exit_callback>*/", "/*<end add_go_exit_callback>*/",
			[]string{"func goexit1() {", "\n"},
			patch.RuntimeProcGoroutineExitPatch,
		)

		procDecl := `func newproc(fn`
		newProc := `newg := newproc1(fn, gp, pc, false, waitReasonZero)`
		if goVersion.Major == GO_MAJOR_1 {
			if goVersion.Minor <= GO_VERSION_17 {
				// to avoid typo check
				const size = "s" + "i" + "z"
				procDecl = `func newproc(` + size + ` int32`
				newProc = `newg := newproc1(fn, argp, ` + size + `, gp, pc)`
			} else if goVersion.Minor <= GO_VERSION_22 {
				newProc = `newg := newproc1(fn, gp, pc)`
			} else if goVersion.Minor <= GO_VERSION_23 {
				newProc = `newg := newproc1(fn, gp, pc, false, waitReasonZero)`
			}
		}
		// see https://github.com/xhd2015/xgo/issues/67
		content = instrument_patch.UpdateContent(
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
		content = instrument_patch.UpdateContent(
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

		content = instrument_patch.UpdateContent(content,
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

func patchRuntimeTesting(origGoroot string, goroot string, goVersion *goinfo.GoVersion) error {
	if goVersion.Major == GO_MAJOR_1 && goVersion.Minor <= GO_VERSION_22 {
		return testingFilePatch.Apply(goroot, nil)
	}
	// go 1.23
	srcFile := testingFilePatch.FilePath.JoinPrefix(origGoroot)
	srcCode, err := fileutil.ReadFile(srcFile)
	if err != nil {
		return err
	}
	newCode, err := ast_patch.Patch(string(srcCode), `package testing
//prepend <define_callbacks>`+toInsertCode(patch.TestingCallbackDeclarations+patch.TestingEndCallbackDeclarations)+`
func tRunner(t *T, fn func(t *T)) {
   //...
   t.start = highPrecisionTimeNow()
   
   //prepend <apply_callbacks>`+toInsertCode(patch.TestingStart+patch.TestingEnd)+`
   fn(t)
}
	`)
	if err != nil {
		return err
	}
	err = fileutil.WriteFile(testingFilePatch.FilePath.JoinPrefix(goroot), []byte(newCode))
	if err != nil {
		return err
	}
	return nil
}

func toInsertCode(code string) string {
	lines := strings.Split(code, "\n")
	for i, line := range lines {
		if i == 0 {
			lines[i] = " " + line
		} else {
			lines[i] = "// " + line
		}
	}
	return strings.Join(lines, "\n")
}

// only required if need to mock time.Sleep
func patchRuntimeTime(goroot string) error {
	runtimeTimeFile := filepath.Join(goroot, filepath.Join(runtimeTime...))
	timeSleepFile := filepath.Join(goroot, filepath.Join(timeSleep...))

	err := instrument_patch.EditFile(runtimeTimeFile, func(content string) (string, error) {
		content = instrument_patch.ReplaceContentAfter(content,
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

	err = instrument_patch.EditFile(timeSleepFile, func(content string) (string, error) {
		content = instrument_patch.ReplaceContentAfter(content,
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
