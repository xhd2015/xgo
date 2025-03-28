package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/xgo/cmd/xgo/patch"
	"github.com/xhd2015/xgo/support/filecopy"
	"github.com/xhd2015/xgo/support/fileutil"
	"github.com/xhd2015/xgo/support/goinfo"
	instrument_patch "github.com/xhd2015/xgo/support/instrument/patch"
	runtime_patch "github.com/xhd2015/xgo/support/instrument/runtime"
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
	bytes, err := readXgoSrc(xgoSrc, []string{"trap_runtime", "xgo_trap.go"})
	if err != nil {
		return false, err
	}
	content := string(bytes)

	content, err = instrument_patch.RemoveBuildIgnore(content)
	if err != nil {
		return false, fmt.Errorf("file %s: %w", filepath.Base(dstFile), err)
	}

	// the func.entry is a field, not a function
	if goVersion.Major == 1 && goVersion.Minor <= 17 {
		entryPatch := "fn.entry() /*>=go1.18*/"
		idx := strings.Index(content, entryPatch)
		if idx < 0 {
			return false, fmt.Errorf("expect %q in xgo_trap.go, actually not found", entryPatch)
		}
		content = strings.ReplaceAll(content, entryPatch, "fn.entry")
	}

	content = runtime_patch.AppendGetFuncNameImpl(goVersion, content)

	return true, os.WriteFile(dstFile, []byte(content), 0755)
}

func patchRuntimeProc(goroot string, goVersion *goinfo.GoVersion) error {
	procFile := filepath.Join(goroot, filepath.Join(runtimeProc...))
	err := instrument_patch.EditFile(procFile, func(content string) (string, error) {
		content = instrument_patch.AddContentAfter(content, "/*<begin set_init_finished_mark>*/", "/*<end set_init_finished_mark>*/", []string{
			"func main() {",
			"doInit(", "runtime_inittask", ")", // first doInit for runtime
			"doInit(", // second init for main
			"close(main_init_done)",
			"\n",
		}, patch.RuntimeProcPatch)

		// goexit1() is called for every exited goroutine
		content = instrument_patch.AddContentAfter(content,
			"/*<begin add_go_exit_callback>*/", "/*<end add_go_exit_callback>*/",
			[]string{"func goexit1() {", "\n"},
			patch.RuntimeProcGoroutineExitPatch,
		)

		return runtime_patch.InstrumentGoroutineCreation(goVersion, content)
	})
	if err != nil {
		return err
	}
	return nil
}

func patchRuntimeTesting(origGoroot string, goroot string, goVersion *goinfo.GoVersion) error {
	if goVersion.Major == goinfo.GO_MAJOR_1 && goVersion.Minor <= goinfo.GO_VERSION_22 {
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
