package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// assume go 1.20
func patchGoSrc(goroot string, xgoSrc string) error {
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
	err = copyReplaceDir(filepath.Join(xgoSrc, "patch"), filepath.Join(goroot, "src/cmd/compile/internal/xgo_rewrite_internal/patch"))
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

	extraDef := `
// xgo
func __xgo_getcurg() unsafe.Pointer
func __xgo_trap(interface{}, []interface{}, []interface{}) (func(), bool)
func __xgo_register_func(fn interface{}, recvName string, argNames []string, resNames []string)
func __xgo_for_each_func(f func(pkgName string,funcName string, pc uintptr, fn interface{}, recvName string, argNames []string, resNames []string))`
	fullFile := filepath.Join(goRoot, runtimeDefFile)

	file, err := os.OpenFile(fullFile, os.O_APPEND|os.O_WRONLY, 0755)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("missing _builtin/runtime.go: %w", err)
		}
		return err
	}
	defer file.Close()

	_, err = file.WriteString(extraDef)
	if err != nil {
		return err
	}
	return nil
}

func copyReplaceDir(srcDir string, targetDir string) error {
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
	return exec.Command("cp", "-R", srcDir, targetAbsDir).Run()
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
	return ioutil.WriteFile(dstFile, content, 0755)
}

func patchCompilerNoder(goroot string) error {
	file := "src/cmd/compile/internal/noder/noder.go"
	return editFile(filepath.Join(goroot, file), func(content string) (string, error) {
		content = addImport(content, "cmd/compile/internal/xgo_rewrite_internal/patch/syntax", "xgo_syntax")
		content = addImport(content, "io", "")
		content = addCodeAfterLine(content, `base.Timer.AddEvent(int64(lines), "lines")`, `	// auto gen
		files := make([]*syntax.File, 0, len(noders))
		for _, n := range noders {
			files = append(files, n.file)
		}
		xgo_syntax.AfterFilesParsed(files, func(name string, r io.Reader) {
			p := &noder{}
			fbase := syntax.NewFileBase(name)
			file, err := syntax.Parse(fbase, r, nil, p.pragma, syntax.CheckBranches) // errors are tracked via p.error
			if err != nil {
				e := err.(syntax.Error)
				base.ErrorfAt(p.makeXPos(e.Pos), "%s", e.Msg)
				return
			}
			p.file = file
			noders = append(noders, p)
		})`)
		return content, nil
	})
}

func patchGcMain(goroot string) error {
	file := "src/cmd/compile/internal/gc/main.go"
	return editFile(filepath.Join(goroot, file), func(content string) (string, error) {
		content = addImport(content, "cmd/compile/internal/xgo_rewrite_internal/patch", "xgo_patch")
		content = addCodeAfterLine(content, `ssagen.InitConfig()`, `	// insert trap points
		xgo_patch.Patch()
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

func addImport(code string, pkgPath string, alias string) string {
	idx := indexSeq(code, []string{"import", "(", "\n"})
	if idx < 0 {
		panic(fmt.Errorf("import not found"))
	}
	clause := strconv.Quote(pkgPath)
	if alias != "" {
		clause = alias + " " + clause
	}
	return code[:idx] + clause + "\n" + code[idx:]
}

func addCodeAfterLine(code string, line string, addCode string) string {
	idx := indexSeq(code, []string{line, "\n"})
	if idx < 0 {
		panic(fmt.Errorf("line not found: %s", line))
	}
	return code[:idx] + addCode + "\n" + code[idx:]
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
