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
)

// assume go 1.20
// the patch should be idempotent
func patchGoSrc(goroot string, xgoSrc string, syncWithLink bool) error {
	if goroot == "" {
		return fmt.Errorf("requires goroot")
	}
	if xgoSrc == "" {
		return fmt.Errorf("requries xgoSrc")
	}
	err := patchRuntimeDef(goroot)
	if err != nil {
		return err
	}

	// copy compiler internal dependencies
	err = importCompileInternalPatch(goroot, xgoSrc, syncWithLink)
	if err != nil {
		return err
	}

	err = patchCompiler(goroot)
	if err != nil {
		return err
	}

	err = addRuntimeTrap(goroot, xgoSrc)
	if err != nil {
		return err
	}

	return nil
}

func importCompileInternalPatch(goroot string, xgoSrc string, syncWithLink bool) error {
	dstDir := filepath.Join(goroot, "src", "cmd", "compile", "internal", "xgo_rewrite_internal", "patch")
	// copy compiler internal dependencies
	err := copyReplaceDir(filepath.Join(xgoSrc, "patch"), dstDir, syncWithLink)
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

func copyReplaceDir(srcDir string, targetDir string, useLink bool) error {
	if srcDir == "" {
		return fmt.Errorf("requires srcDir")
	}
	targetAbsDir, err := filepath.Abs(targetDir)
	if err != nil {
		return err
	}
	if targetAbsDir == "/" {
		return fmt.Errorf("cannot replace /")
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	if !strings.HasPrefix(targetAbsDir, homeDir+"/.xgo") && !strings.HasPrefix(targetAbsDir, "/tmp/") {
		return fmt.Errorf("replace not permitted:%s", targetDir)
	}
	err = os.RemoveAll(targetAbsDir)
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Dir(targetAbsDir), 0755)
	if err != nil {
		return err
	}
	if useLink {
		return linkFiles(srcDir, targetAbsDir)
	}
	return exec.Command("cp", "-R", srcDir, targetAbsDir).Run()
}

func linkFiles(srcDir string, targetDir string) error {
	absDir, err := filepath.Abs(srcDir)
	if err != nil {
		return err
	}
	return filepath.WalkDir(absDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !strings.HasPrefix(path, absDir) {
			return fmt.Errorf("invalid path: %s", path)
		}
		relPath := path[len(absDir):]
		if strings.HasPrefix(relPath, string(os.PathSeparator)) {
			relPath = relPath[1:]
		}
		// if relPath is "", it is root dir
		targetPath := filepath.Join(targetDir, relPath)
		if d.IsDir() {
			err := os.MkdirAll(targetPath, 0755)
			if err != nil {
				return err
			}
			return nil
		}
		// link file
		return os.Symlink(path, targetPath)
	})
}

func patchCompiler(goroot string) error {
	// src/cmd/compile/internal/noder/noder.go
	err := patchCompilerNoder(goroot)
	if err != nil {
		return err
	}
	err = patchGcMain(goroot)
	if err != nil {
		return err
	}
	return nil
}

func addRuntimeTrap(goroot string, xgoSrc string) error {
	srcFile := filepath.Join(xgoSrc, "runtime", "trap_runtime", "xgo_trap.go")
	dstFile := filepath.Join(goroot, "src", "runtime", "xgo_trap.go")
	content, err := ioutil.ReadFile(srcFile)
	if err != nil {
		return err
	}
	content = bytes.Replace(content, []byte("//go:build ignore\n"), nil, 1)

	// TODO: limit the patch for go1.20
	content = append(content, []byte(patch.RuntimeFuncNamePatch)...)
	return ioutil.WriteFile(dstFile, content, 0755)
}

func patchCompilerNoder(goroot string) error {
	file := "src/cmd/compile/internal/noder/noder.go"
	return editFile(filepath.Join(goroot, file), func(content string) (string, error) {
		content = addCodeAfterImports(content,
			"/*<begin file_autogen_import>*/", "/*<end file_autogen_import>*/",
			[]string{
				`xgo_syntax "cmd/compile/internal/xgo_rewrite_internal/patch/syntax"`,
				`"io"`,
			},
		)
		content = addContentAfter(content, "/*<begin file_autogen>*/", "/*<end file_autogen>*/", []string{
			`for _, p := range noders {`,
			`base.Timer.AddEvent(int64(lines), "lines")`,
			"\n",
		},
			patch.NoderFiles)
		return content, nil
	})
}

func patchGcMain(goroot string) error {
	file := "src/cmd/compile/internal/gc/main.go"
	return editFile(filepath.Join(goroot, file), func(content string) (string, error) {
		content = addCodeAfterImports(content,
			"/*<begin gc_import>*/", "/*<end gc_import>*/",
			[]string{
				`xgo_patch "cmd/compile/internal/xgo_rewrite_internal/patch"`,
				`xgo_record "cmd/compile/internal/xgo_rewrite_internal/patch/record"`,
			},
		)

		content = addContentAfter(content,
			"/*<begin patch>*/", "/*<end patch>*/",
			[]string{`noder.LoadPackage(flag.Args())`, `ssagen.InitConfig()`, "\n"},
			`	// insert trap points
		if os.Getenv("XGO_COMPILER_ENABLE")=="true" {
		    xgo_patch.Patch()
		}
`)

		content = replaceContentAfter(content,
			"/*<begin prevent_inline>*/", "/*<end prevent_inline>*/",
			[]string{`base.Timer.Start("fe", "inlining")`, `if base.Flag.LowerL != 0 {`, "\n"},
			`inline.InlinePackage(profile)`,
			`	// NOTE: turn off inline for go1.20 if there is any rewrite
		if !xgo_record.HasRewritten() {
			inline.InlinePackage(profile)
		}
`)
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
	return copyReplaceDir(goroot, dstDir, false)
}
