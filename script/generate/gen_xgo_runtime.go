package main

import (
	"os"
	"path/filepath"

	"github.com/xhd2015/xgo/support/filecopy"
	"github.com/xhd2015/xgo/support/fileutil"
)

func genXgoRuntime(cmd string, rootDir string) error {
	runtimeDir := filepath.Join(rootDir, "runtime")
	genRuntimeDir := filepath.Join(rootDir, "cmd", "xgo", "runtime_gen")

	traceRenderingStackModel := filepath.Join(rootDir, "cmd", "xgo", "trace", "render", "stack_model", "stack_model.go")
	runtimeStackModel := filepath.Join(runtimeDir, "trace", "stack_model", "stack_model.go")

	// copy stack model from xgo to runtime first
	err := copyStackTraceExport(cmd, traceRenderingStackModel, runtimeStackModel)
	if err != nil {
		return err
	}

	// then copy runtime to xgo/runtime_gen
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

func copyStackTraceExport(cmd string, srcFile string, dstFile string) error {
	content, err := fileutil.ReadFile(srcFile)
	if err != nil {
		return err
	}
	newCode := getCmdPrelude(cmd) + "// keep the same with cmd/xgo/trace/render/stack_model/stack_model.go\n" + string(content)

	return fileutil.WriteFile(dstFile, []byte(newCode))
}
