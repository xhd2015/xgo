package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/xhd2015/xgo/cmd/xgo/patch"
	"github.com/xhd2015/xgo/support/filecopy"
	"github.com/xhd2015/xgo/support/goinfo"
)

// assume go 1.20
// the patch should be idempotent
func patchGoSrc(goroot string, xgoSrc string, goVersion *goinfo.GoVersion, noInstrument bool, syncWithLink bool) error {
	if goroot == "" {
		return fmt.Errorf("requires goroot")
	}
	if xgoSrc == "" {
		return fmt.Errorf("requries xgoSrc")
	}

	err := addRuntimeTrap(goroot, goVersion, xgoSrc)
	if err != nil {
		return err
	}

	if noInstrument {
		return nil
	}

	err = patchRuntimeDef(goroot)
	if err != nil {
		return err
	}

	// copy compiler internal dependencies
	err = importCompileInternalPatch(goroot, xgoSrc, syncWithLink)
	if err != nil {
		return err
	}

	err = patchCompiler(goroot, goVersion)
	if err != nil {
		return err
	}

	return nil
}

func importCompileInternalPatch(goroot string, xgoSrc string, syncWithLink bool) error {
	dstDir := filepath.Join(goroot, "src", "cmd", "compile", "internal", "xgo_rewrite_internal", "patch")
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

func patchRuntimeDef(goRoot string) error {
	err := prepareRuntimeDefs(goRoot)
	if err != nil {
		return err
	}

	// run mkbuiltin
	cmd := exec.Command("go", "run", "mkbuiltin.go")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Dir = filepath.Join(goRoot, "src/cmd/compile/internal/typecheck")
	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func prepareRuntimeDefs(goRoot string) error {
	runtimeDefFile := "src/cmd/compile/internal/typecheck/_builtin/runtime.go"
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

func patchCompiler(goroot string, goVersion *goinfo.GoVersion) error {
	// src/cmd/compile/internal/noder/noder.go
	err := patchCompilerNoder(goroot, goVersion)
	if err != nil {
		return fmt.Errorf("patching noder: %w", err)
	}
	err = patchGcMain(goroot, goVersion)
	if err != nil {
		return fmt.Errorf("patching gc main:%w", err)
	}
	return nil
}

func addRuntimeTrap(goroot string, goVersion *goinfo.GoVersion, xgoSrc string) error {
	srcFile := filepath.Join(xgoSrc, "runtime", "trap_runtime", "xgo_trap.go")
	dstFile := filepath.Join(goroot, "src", "runtime", "xgo_trap.go")
	content, err := ioutil.ReadFile(srcFile)
	if err != nil {
		return err
	}
	content = bytes.Replace(content, []byte("//go:build ignore\n"), nil, 1)

	// TODO: remove the patch
	if goVersion.Major == 1 && goVersion.Minor == 20 {
		content = append(content, []byte(patch.RuntimeFuncNamePatch)...)
	}
	return ioutil.WriteFile(dstFile, content, 0755)
}

func patchCompilerNoder(goroot string, goVersion *goinfo.GoVersion) error {
	file := "src/cmd/compile/internal/noder/noder.go"
	var noderFiles string
	if goVersion.Major == 1 {
		minor := goVersion.Minor
		if minor == 20 {
			noderFiles = patch.NoderFiles_1_20
		} else if minor == 21 {
			noderFiles = patch.NoderFiles_1_22
		} else if minor == 22 {
			noderFiles = patch.NoderFiles_1_22
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
		content = addContentAfter(content, "/*<begin file_autogen>*/", "/*<end file_autogen>*/", []string{
			`func LoadPackage`,
			`for _, p := range noders {`,
			`base.Timer.AddEvent(int64(lines), "lines")`,
			"\n",
		},
			noderFiles)
		return content, nil
	})
}

func patchGcMain(goroot string, goVersion *goinfo.GoVersion) error {
	file := "src/cmd/compile/internal/gc/main.go"
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

		content = addContentAfter(content,
			"/*<begin patch>*/", "/*<end patch>*/",
			[]string{`noder.LoadPackage(flag.Args())`, `dwarfgen.RecordPackageName()`, `ssagen.InitConfig()`, "\n"},
			`	// insert trap points
		if os.Getenv("XGO_COMPILER_ENABLE")=="true" {
		    xgo_patch.Patch()
		}
`)

		// there are two ways to turn off inline
		// - 1. by not calling to inline.InlinePackage
		// - 2. by override base.Flag.LowerL to 0
		// prefer 1 because it is more focused
		if go120 || go121 {
			// go1.20 does not respect rewritten content when inlined
			content = replaceContentAfter(content,
				"/*<begin prevent_inline>*/", "/*<end prevent_inline>*/",
				[]string{`base.Timer.Start("fe", "inlining")`, `if base.Flag.LowerL != 0 {`, "\n"},
				`inline.InlinePackage(profile)`,
				`	// NOTE: turn off inline if there is any rewrite
		if !xgo_record.HasRewritten() {
			inline.InlinePackage(profile)
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
	idx := indexSeq(code, []string{"import", "(", "\n"})
	if idx < 0 {
		panic(fmt.Errorf("import not found"))
	}
	return insertConentNoDudplicate(code, beginMark, endMark, idx, strings.Join(contents, "\n")+"\n")
}

func addContentAfter(content string, beginMark string, endMark string, seq []string, addContent string) string {
	idx := indexSeq(content, seq)
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
	idx := indexSeq(content, seq)
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

func indexSeq(s string, seqs []string) int {
	base := 0
	for _, seq := range seqs {
		idx := strings.Index(s, seq)
		if idx < 0 {
			return -1
		}
		s = s[idx+len(seq):]
		base += idx + len(seq)
	}
	return base
}

func syncGoroot(goroot string, dstDir string) error {
	// check if src goroot has src/runtime
	srcRuntimeDir := filepath.Join(goroot, "src", "runtime")
	err := assertDir(srcRuntimeDir)
	if err != nil {
		return err
	}
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
	// need copy, delete target dst dir first
	// TODO: use git worktree add if .git exists
	return filecopy.CopyReplaceDir(goroot, dstDir, false)
}
