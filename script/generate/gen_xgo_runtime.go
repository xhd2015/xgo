package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/xgo/support/filecopy"
	"github.com/xhd2015/xgo/support/fileutil"
)

func genXgoRuntime(rootDir string) error {
	runtimeDir := filepath.Join(rootDir, "runtime")
	genRuntimeDir := filepath.Join(rootDir, "cmd", "xgo", "runtime_gen")

	stackExportGen := filepath.Join(rootDir, "cmd", "xgo", "test-explorer", "trace_stack_export_gen.go")
	err := copyStackTraceExport(filepath.Join(runtimeDir, "trace", "stack_export.go"), stackExportGen, "test_explorer")
	if err != nil {
		return err
	}
	err = filecopy.NewOptions().Ignore("test").IncludeSuffix(".go", "go.mod").IgnoreSuffix("_test.go").CopyReplaceDir(runtimeDir, genRuntimeDir)
	if err != nil {
		return err
	}
	err = os.Rename(filepath.Join(genRuntimeDir, "go.mod"), filepath.Join(genRuntimeDir, "go.mod.txt"))
	if err != nil {
		return err
	}
	return nil
}

func copyStackTraceExport(srcFile string, dstFile string, dstPkg string) error {
	content, err := fileutil.ReadFile(srcFile)
	if err != nil {
		return err
	}
	s := string(content)
	const PKG = "package trace"
	if !strings.Contains(s, PKG) {
		return fmt.Errorf("package trace not found")
	}
	newCode := strings.Replace(s, PKG, prelude+"package "+dstPkg, 1)

	return fileutil.WriteFile(dstFile, []byte(newCode))
}
