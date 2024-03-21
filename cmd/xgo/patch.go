package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/xhd2015/xgo/cmd/xgo/patch"
	"github.com/xhd2015/xgo/support/filecopy"
	"github.com/xhd2015/xgo/support/goinfo"
	"github.com/xhd2015/xgo/support/strutil"
)

// assume go 1.20
// the patch should be idempotent
func patchRuntimeAndCompiler(goroot string, xgoSrc string, goVersion *goinfo.GoVersion, syncWithLink bool, revisionChanged bool) error {
	if goroot == "" {
		return fmt.Errorf("requires goroot")
	}
	if isDevelopment && xgoSrc == "" {
		return fmt.Errorf("requries xgoSrc")
	}
	if !isDevelopment && !revisionChanged {
		return nil
	}

	// runtime
	err := patchRuntimeAndTesting(goroot)
	if err != nil {
		return err
	}

	// compiler
	err = patchCompiler(goroot, goVersion, xgoSrc, revisionChanged, syncWithLink)
	if err != nil {
		return err
	}

	return nil
}

func patchRuntimeAndTesting(goroot string) error {
	err := patchRuntimeProc(goroot)
	if err != nil {
		return err
	}
	err = patchRuntimeTesting(goroot)
	if err != nil {
		return err
	}
	return nil
}

func patchRuntimeProc(goroot string) error {
	anchors := []string{
		"func main() {",
		"doInit(", "runtime_inittask", ")", // first doInit for runtime
		"doInit(", // second init for main
		"close(main_init_done)",
		"\n",
	}
	procGo := filepath.Join(goroot, "src", "runtime", "proc.go")
	err := editFile(procGo, func(content string) (string, error) {
		content = addContentAfter(content, "/*<begin set_init_finished_mark>*/", "/*<end set_init_finished_mark>*/", anchors, patch.RuntimeProcPatch)

		// goexit1() is called for every exited goroutine
		content = addContentAfter(content,
			"/*<begin add_go_exit_callback>*/", "/*<end add_go_exit_callback>*/",
			[]string{"func goexit1() {", "\n"},
			patch.RuntimeProcGoroutineExitPatch,
		)
		return content, nil
	})
	if err != nil {
		return err
	}
	return nil
}

func patchRuntimeTesting(goroot string) error {
	testingFile := filepath.Join(goroot, "src", "testing", "testing.go")
	return editFile(testingFile, func(content string) (string, error) {
		// func tRunner(t *T, fn func(t *T)) {
		anchor := []string{"func tRunner(t *T", "{", "\n"}
		content = addContentBefore(content,
			"/*<begin declare_testing_callback>*/", "/*<end declare_testing_callback>*/",
			anchor,
			patch.TestingCallbackDeclarations,
		)
		content = addContentAfter(content,
			"/*<begin call_testing_callback>*/", "/*<end call_testing_callback>*/",
			anchor,
			patch.TestingStart,
		)
		return content, nil
	})
}
func importCompileInternalPatch(goroot string, xgoSrc string, forceReset bool, syncWithLink bool) error {
	dstDir := filepath.Join(goroot, "src", "cmd", "compile", "internal", "xgo_rewrite_internal", "patch")
	if isDevelopment {
		// copy compiler internal dependencies
		err := filecopy.CopyReplaceDir(filepath.Join(xgoSrc, "patch"), dstDir, syncWithLink)
		if err != nil {
			return err
		}

		// remove patch/go.mod
		err = os.RemoveAll(filepath.Join(dstDir, "go.mod"))
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return nil
			}
			return err
		}
		return nil
	}

	if forceReset {
		// -a causes repatch
		err := os.RemoveAll(dstDir)
		if err != nil {
			return err
		}
	} else {
		// check if already copied
		_, statErr := os.Stat(dstDir)
		if statErr == nil {
			// skip copy if already exists
			return nil
		}
	}

	// read from embed
	err := fs.WalkDir(patchEmbed, "patch_compiler", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == "patch_compiler" {
			return os.MkdirAll(dstDir, 0755)
		}
		// TODO: test on windows if "/" works
		dstPath := filepath.Join(dstDir, strings.TrimPrefix(path, "patch_compiler/"))
		if d.IsDir() {
			return os.MkdirAll(dstPath, 0755)
		}

		content, err := patchEmbed.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(dstPath, content, 0755)
	})
	if err != nil {
		return err
	}

	return nil
}

func patchRuntimeDef(goRoot string, goVersion *goinfo.GoVersion) error {
	err := prepareRuntimeDefs(goRoot, goVersion)
	if err != nil {
		return err
	}

	// run mkbuiltin
	cmd := exec.Command(filepath.Join(goRoot, "bin", "go"), "run", "mkbuiltin.go")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	dir := "src/cmd/compile/internal/typecheck"
	if goVersion.Major == 1 && goVersion.Minor <= 16 {
		dir = "src/cmd/compile/internal/gc"
	}
	cmd.Dir = filepath.Join(goRoot, dir)

	cmd.Env = os.Environ()
	cmd.Env, err = patchEnvWithGoroot(cmd.Env, goRoot)
	if err != nil {
		return err
	}

	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func prepareRuntimeDefs(goRoot string, goVersion *goinfo.GoVersion) error {
	runtimeDefFile := "src/cmd/compile/internal/typecheck/_builtin/runtime.go"
	if goVersion.Major == 1 && goVersion.Minor <= 19 {
		if goVersion.Minor > 16 {
			// in go1.19 and below, builtin has no _ prefix
			runtimeDefFile = "src/cmd/compile/internal/typecheck/builtin/runtime.go"
		} else {
			runtimeDefFile = "src/cmd/compile/internal/gc/builtin/runtime.go"
		}
	}
	fullFile := filepath.Join(goRoot, runtimeDefFile)

	extraDef := patch.RuntimeExtraDef
	return editFile(fullFile, func(content string) (string, error) {
		content = addContentAfter(content,
			`/*<begin extra_runtime_func>*/`, `/*<end extra_runtime_func>*/`,
			[]string{`var x86HasFMA bool`, `var armHasVFPv4 bool`, `var arm64HasATOMICS bool`},
			extraDef,
		)
		return content, nil
	})
}

func patchCompiler(goroot string, goVersion *goinfo.GoVersion, xgoSrc string, forceReset bool, syncWithLink bool) error {
	// copy compiler internal dependencies
	err := importCompileInternalPatch(goroot, xgoSrc, forceReset, syncWithLink)
	if err != nil {
		return err
	}
	runtimeDefUpdated, err := addRuntimeFunctions(goroot, goVersion, xgoSrc)
	if err != nil {
		return err
	}

	if runtimeDefUpdated {
		err = patchRuntimeDef(goroot, goVersion)
		if err != nil {
			return err
		}
	}

	err = patchCompilerInternal(goroot, goVersion)
	if err != nil {
		return err
	}
	return nil
}

func patchCompilerInternal(goroot string, goVersion *goinfo.GoVersion) error {
	// src/cmd/compile/internal/noder/noder.go
	err := patchCompilerNoder(goroot, goVersion)
	if err != nil {
		return fmt.Errorf("patching noder: %w", err)
	}
	if goVersion.Major == 1 && (goVersion.Minor == 18 || goVersion.Minor == 19) {
		err := poatchIRGenericGen(goroot, goVersion)
		if err != nil {
			return fmt.Errorf("patching generic trap: %w", err)
		}
	}
	err = patchGcMain(goroot, goVersion)
	if err != nil {
		return fmt.Errorf("patching gc main:%w", err)
	}
	return nil
}

// addRuntimeFunctions always copy file
func addRuntimeFunctions(goroot string, goVersion *goinfo.GoVersion, xgoSrc string) (updated bool, err error) {
	dstFile := filepath.Join(goroot, "src", "runtime", "xgo_trap.go")
	var content []byte
	if isDevelopment {
		srcFile := filepath.Join(xgoSrc, "runtime", "trap_runtime", "xgo_trap.go")
		content, err = ioutil.ReadFile(srcFile)
	} else {
		content, err = patchEmbed.ReadFile("patch_compiler/trap_runtime/xgo_trap.go")
	}
	if err != nil {
		return false, err
	}

	content = bytes.Replace(content, []byte("//go:build ignore\n"), nil, 1)

	// the func.entry is a field, not a function
	if goVersion.Major == 1 && goVersion.Minor <= 17 {
		entryPatch := "fn.entry() /*>=go1.18*/"
		entryPatchBytes := []byte(entryPatch)
		idx := bytes.Index(content, entryPatchBytes)
		if idx < 0 {
			return false, fmt.Errorf("expect %q in xgo_trap.go, actually not found", entryPatch)
		}
		oldContent := content
		content = append(content[:idx], []byte("fn.entry")...)
		content = append(content, oldContent[idx+len(entryPatchBytes):]...)
	}

	// TODO: remove the patch
	if goVersion.Major == 1 && goVersion.Minor == 20 {
		content = append(content, []byte(patch.RuntimeFuncNamePatch)...)
	}
	return true, ioutil.WriteFile(dstFile, content, 0755)
}

func patchCompilerNoder(goroot string, goVersion *goinfo.GoVersion) error {
	file := "src/cmd/compile/internal/noder/noder.go"
	var noderFiles string
	if goVersion.Major == 1 {
		minor := goVersion.Minor
		if minor == 16 {
			file = "src/cmd/compile/internal/gc/noder.go"
			noderFiles = patch.NoderFiles_1_17
		} else if minor == 17 {
			noderFiles = patch.NoderFiles_1_17
		} else if minor == 18 {
			noderFiles = patch.NoderFiles_1_17
		} else if minor == 19 {
			noderFiles = patch.NoderFiles_1_17
		} else if minor == 20 {
			noderFiles = patch.NoderFiles_1_20
		} else if minor == 21 {
			noderFiles = patch.NoderFiles_1_21
		} else if minor == 22 {
			noderFiles = patch.NoderFiles_1_21
		}
	}
	if noderFiles == "" {
		return fmt.Errorf("unsupported: %v", goVersion)
	}
	return editFile(filepath.Join(goroot, file), func(content string) (string, error) {
		content = addCodeAfterImports(content,
			"/*<begin file_autogen_import>*/", "/*<end file_autogen_import>*/",
			[]string{
				`xgo_syntax "cmd/compile/internal/xgo_rewrite_internal/patch/syntax"`,
				`"io"`,
			},
		)
		var anchors []string
		if goVersion.Major == 1 && goVersion.Minor <= 16 {
			anchors = []string{
				"func parseFiles(filenames []string)",
				"for _, p := range noders {",
				"localpkg.Height = myheight",
				"\n",
			}
		} else {
			anchors = []string{
				`func LoadPackage`,
				`for _, p := range noders {`,
				`base.Timer.AddEvent(int64(lines), "lines")`,
				"\n",
			}
		}
		content = addContentAfter(content, "/*<begin file_autogen>*/", "/*<end file_autogen>*/", anchors,
			noderFiles)
		return content, nil
	})
}

func poatchIRGenericGen(goroot string, goVersion *goinfo.GoVersion) error {
	file := "src/cmd/compile/internal/noder/irgen.go"
	return editFile(filepath.Join(goroot, file), func(content string) (string, error) {
		imports := []string{
			`xgo_patch "cmd/compile/internal/xgo_rewrite_internal/patch"`,
		}
		if goVersion.Major == 1 && goVersion.Minor >= 19 {
			imports = append(imports, `"os"`)
		}
		content = addCodeAfterImports(content,
			"/*<begin irgen_autogen_import>*/", "/*<end irgen_autogen_import>*/",
			imports,
		)
		content = addContentAfter(content, "/*<begin irgen_generic_trap_autogen>*/", "/*<end irgen_generic_trap_autogen>*/", []string{
			`func (g *irgen) generate(noders []*noder) {`,
			`types.DeferCheckSize()`,
			`base.ExitIfErrors()`,
			`typecheck.DeclareUniverse()`,
			"\n",
		},
			patch.GenericTrapForGo118And119)
		return content, nil
	})
}

func patchGcMain(goroot string, goVersion *goinfo.GoVersion) error {
	file := "src/cmd/compile/internal/gc/main.go"
	go116AndUnder := goVersion.Major == 1 && goVersion.Minor <= 16
	go117 := goVersion.Major == 1 && goVersion.Minor == 17
	go118 := goVersion.Major == 1 && goVersion.Minor == 18
	go119 := goVersion.Major == 1 && goVersion.Minor == 19
	go119AndUnder := goVersion.Major == 1 && goVersion.Minor <= 19
	go120 := goVersion.Major == 1 && goVersion.Minor == 20
	go121 := goVersion.Major == 1 && goVersion.Minor == 21
	go122 := goVersion.Major == 1 && goVersion.Minor == 22

	return editFile(filepath.Join(goroot, file), func(content string) (string, error) {
		imports := []string{
			`xgo_patch "cmd/compile/internal/xgo_rewrite_internal/patch"`,
			`xgo_record "cmd/compile/internal/xgo_rewrite_internal/patch/record"`,
		}
		content = addCodeAfterImports(content,
			"/*<begin gc_import>*/", "/*<end gc_import>*/",
			imports,
		)
		initRuntimeTypeCheckGo117 := `typecheck.InitRuntime()`

		var beforePatchContent string
		var patchAnchors []string

		if go116AndUnder {
			// go1.16 is pretty old
			patchAnchors = []string{
				"loadsys()",
				"parseFiles(flag.Args())",
				"finishUniverse()",
				"recordPackageName()",
			}
		} else {
			patchAnchors = []string{`noder.LoadPackage(flag.Args())`, `dwarfgen.RecordPackageName()`}
			if !go117 {
				patchAnchors = append(patchAnchors, `ssagen.InitConfig()`)
			} else {
				// go 1.17 needs to call typecheck.InitRuntime() before patch
				beforePatchContent = initRuntimeTypeCheckGo117 + "\n"
			}
		}
		patchAnchors = append(patchAnchors, "\n")
		content = addContentAfter(content,
			"/*<begin patch>*/", "/*<end patch>*/",
			patchAnchors,
			`	// insert trap points
		if os.Getenv("XGO_COMPILER_ENABLE")=="true" {
		    `+beforePatchContent+`xgo_patch.Patch()
		}
`)

		if go117 {
			// go1.17 needs to adjust typecheck.InitRuntime before patch
			content = replaceContentAfter(content,
				"/*<begin patch_init_runtime_type>*/", "/*<end patch_init_runtime_type>*/",
				[]string{`escape.Funcs(typecheck.Target.Decls)`, `if base.Flag.CompilingRuntime {`, "}", "\n"},
				initRuntimeTypeCheckGo117,
				`if os.Getenv("XGO_COMPILER_ENABLE")!="true" {
					`+initRuntimeTypeCheckGo117+`
				}`,
			)
		}

		// there are two ways to turn off inline
		// - 1. by not calling to inline.InlinePackage
		// - 2. by override base.Flag.LowerL to 0
		// prefer 1 because it is more focused
		if go116AndUnder {
			inlineGuard := `if Debug.l != 0 {`
			inlineAnchors := []string{
				`fninit(xtop)`,
				`Curfn = nil`,
				`// Phase 5: Inlining`,
				`if Debug_typecheckinl != 0 {`,
				"\n",
			}
			content = replaceContentAfter(content,
				"/*<begin prevent_inline>*/", "/*<end prevent_inline>*/",
				inlineAnchors,
				inlineGuard,
				`	// NOTE: turn off inline if there is any rewrite
		`+strings.TrimSuffix(inlineGuard, " {")+` && !xgo_record.HasRewritten() {`)
		} else if go117 || go118 || go119 || go120 || go121 {
			inlineCall := `inline.InlinePackage(profile)`
			if go119AndUnder {
				// go1.19 and under does not hae PGO
				inlineCall = `inline.InlinePackage()`
			}
			// go1.20 does not respect rewritten content when inlined
			content = replaceContentAfter(content,
				"/*<begin prevent_inline>*/", "/*<end prevent_inline>*/",
				[]string{`base.Timer.Start("fe", "inlining")`, `if base.Flag.LowerL != 0 {`, "\n"},
				inlineCall,
				`	// NOTE: turn off inline if there is any rewrite
		if !xgo_record.HasRewritten() {
			`+inlineCall+`
		}
`)
		} else if go122 {
			// go1.22 also does not respect rewritten content when inlined
			// NOTE: the override of LowerL is inserted after xgo_patch.Patch()
			content = addContentAfter(content,
				"/*<begin prevent_inline_by_override_flag>*/", "/*<end prevent_inline_by_override_flag>*/",
				[]string{`if base.Flag.LowerL <= 1 {`, `base.Flag.LowerL = 1 - base.Flag.LowerL`, "}", "xgo_patch.Patch()", "}", "\n"},
				`	// NOTE: turn off inline if there is any rewrite
						if xgo_record.HasRewritten() {
							base.Flag.LowerL = 0
						}
				`)
		} else {
			return "", fmt.Errorf("inline for %v not defined", goVersion)
		}
		return content, nil
	})
}

func editFile(file string, callback func(content string) (string, error)) error {
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	content := string(bytes)
	newContent, err := callback(content)
	if err != nil {
		return err
	}
	if newContent == content {
		return nil
	}
	return ioutil.WriteFile(file, []byte(newContent), 0755)
}

func addCodeAfterImports(code string, beginMark string, endMark string, contents []string) string {
	idx := indexSeq(code, []string{"import", "(", "\n"}, false)
	if idx < 0 {
		panic(fmt.Errorf("import not found"))
	}
	return insertConentNoDudplicate(code, beginMark, endMark, idx, strings.Join(contents, "\n")+"\n")
}

func addContentBefore(content string, beginMark string, endMark string, seq []string, addContent string) string {
	return addContentAt(content, beginMark, endMark, seq, addContent, true)
}

func addContentAfter(content string, beginMark string, endMark string, seq []string, addContent string) string {
	return addContentAt(content, beginMark, endMark, seq, addContent, false)
}

func addContentAt(content string, beginMark string, endMark string, seq []string, addContent string, begin bool) string {
	idx := indexSeq(content, seq, begin)
	if idx < 0 {
		panic(fmt.Errorf("sequence not found: %v", seq))
	}
	return insertConentNoDudplicate(content, beginMark, endMark, idx, addContent)
}

func replaceContentAfter(content string, beginMark string, endMark string, seq []string, target string, replaceContent string) string {
	if replaceContent == "" {
		return content
	}
	closuerContent := beginMark + "\n" + replaceContent + "\n" + endMark + "\n"
	idx := indexSeq(content, seq, false)
	if idx < 0 {
		panic(fmt.Errorf("sequence not found: %v", seq))
	}
	if strings.Contains(content, closuerContent) {
		return content
	}
	content, ok := tryReplaceWithMark(content, beginMark, endMark, closuerContent)
	if ok {
		return content
	}
	targetIdx := strings.Index(content[idx:], target)
	if targetIdx < 0 {
		panic(fmt.Errorf("not found: %s", target))
	}
	return content[:idx+targetIdx] + closuerContent + content[idx+targetIdx+len(target):]
}

// signature example: /*<begin ident>*/ {content} /*<end ident>*/
func insertConentNoDudplicate(content string, beginMark string, endMark string, idx int, insertContent string) string {
	if insertContent == "" {
		return content
	}
	closuerContent := beginMark + "\n" + insertContent + "\n" + endMark + "\n"
	content, ok := tryReplaceWithMark(content, beginMark, endMark, closuerContent)
	if ok {
		return content
	}
	if strings.Contains(content, closuerContent) {
		return content
	}
	return content[:idx] + closuerContent + content[idx:]
}

func tryReplaceWithMark(content string, beginMark string, endMark string, closureContent string) (string, bool) {
	beginIdx := strings.Index(content, beginMark)
	if beginIdx < 0 {
		return content, false
	}
	endIdx := strings.Index(content, endMark)
	if endIdx < 0 {
		return content, false
	}
	lastIdx := endIdx + len(endMark)
	if lastIdx+1 < len(content) && content[lastIdx+1] == '\n' {
		lastIdx++
	}
	return content[:beginIdx] + closureContent + content[lastIdx:], true
}

func indexSeq(s string, sequence []string, begin bool) int {
	return strutil.IndexSequenceAt(s, sequence, begin)
}

func checkRevisionChanged(revisionFile string, currentRevision string) (bool, error) {
	savedRevision, err := readOrEmpty(revisionFile)
	if err != nil {
		return false, err
	}
	if savedRevision == "" || savedRevision != currentRevision {
		return true, nil
	}
	return false, nil
}

func readOrEmpty(file string) (string, error) {
	version, err := os.ReadFile(file)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil
		}
		return "", err
	}
	s := string(version)
	s = strings.TrimSuffix(s, "\n")
	s = strings.TrimSuffix(s, "\r")
	return s, nil
}

// NOTE: flagA never cause goroot to reset
func syncGoroot(goroot string, dstDir string, forceCopy bool) error {
	// check if src goroot has src/runtime
	srcRuntimeDir := filepath.Join(goroot, "src", "runtime")
	err := assertDir(srcRuntimeDir)
	if err != nil {
		return err
	}
	if !forceCopy {
		srcGoBin := filepath.Join(goroot, "bin", "go")
		dstGoBin := filepath.Join(dstDir, "bin", "go")

		srcFile, err := os.Stat(srcGoBin)
		if err != nil {
			return nil
		}
		if srcFile.IsDir() {
			return fmt.Errorf("bad goroot: %s", goroot)
		}

		dstFile, err := os.Stat(dstGoBin)
		if err != nil {
			if !os.IsNotExist(err) {
				return err
			}
			err = nil
		}

		if dstFile != nil && !dstFile.IsDir() && dstFile.Size() == srcFile.Size() {
			// already copied
			return nil
		}
	}

	// need copy, delete target dst dir first
	// TODO: use git worktree add if .git exists
	return filecopy.CopyReplaceDir(goroot, dstDir, false)
}
