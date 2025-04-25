package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/xgo/script/generate/gen_defs"
	"github.com/xhd2015/xgo/support/filecopy"
	"github.com/xhd2015/xgo/support/fileutil"
)

func genXgoRuntime(rootDir string, needCopyTrace bool) error {
	if needCopyTrace {
		// copy stack model from xgo to runtime first
		err := copyTraceModel(rootDir)
		if err != nil {
			return err
		}
	}
	runtimeDir := filepath.Join(rootDir, "runtime")
	genRuntimeDir := filepath.Join(rootDir, "cmd", "xgo", "asset", "runtime_gen")

	return copyGoModule(runtimeDir, genRuntimeDir, []string{".xgo", "test"})
}

func copyGoModule(src string, dst string, ignores []string) error {
	// then copy runtime to xgo/runtime_gen
	err := filecopy.NewOptions().Ignore(ignores...).IncludeSuffix(".go", "go.mod").IgnoreSuffix("_test.go").CopyReplaceDir(src, dst)
	if err != nil {
		return err
	}
	err = os.Rename(filepath.Join(dst, "go.mod"), filepath.Join(dst, "go.mod.txt"))
	if err != nil {
		return err
	}
	return nil
}

func copyTraceModel(rootDir string) error {
	runtimeDir := filepath.Join(rootDir, "runtime")

	traceRenderingStackModel := filepath.Join(rootDir, "cmd", "xgo", "trace", "render", "stack_model", "stack_model.go")
	runtimeStackModel := filepath.Join(runtimeDir, "trace", "stack_model", "stack_model.go")

	// copy stack model from xgo to runtime
	return copyStackTraceExport(string(gen_defs.GenernateType_RuntimeTraceModel), traceRenderingStackModel, runtimeStackModel)
}

func copyStackTraceExport(cmd string, srcFile string, dstFile string) error {
	content, err := fileutil.ReadFile(srcFile)
	if err != nil {
		return err
	}
	newCode := getCmdPrelude(cmd) + "// keep the same with cmd/xgo/trace/render/stack_model/stack_model.go\n" + string(content)

	return fileutil.WriteFile(dstFile, []byte(newCode))
}

func replaceCoreFunc() error {
	return fileutil.UpdateFile(filepath.Join(funcPath...), func(content []byte) (bool, []byte, error) {
		newContent, err := replaceXgoFunc(string(content), "")
		if err != nil {
			return false, nil, err
		}
		return true, []byte(newContent), nil
	})
}

func replaceFuncInXgoTrap() error {
	return fileutil.UpdateFile(filepath.Join(xgoTrapTemplatePath...), func(content []byte) (bool, []byte, error) {
		newContent, err := replaceXgoFunc(string(content), "Xgo")
		if err != nil {
			return false, nil, err
		}
		return true, []byte(newContent), nil
	})
}

func replaceFuncInLegacyRuntimeLink() error {
	return fileutil.UpdateFile(filepath.Join(legacyRuntimeLinkTemplatePath...), func(content []byte) (bool, []byte, error) {
		newContent, err := replaceXgoFunc(string(content), "Xgo")
		if err != nil {
			return false, nil, err
		}
		return true, []byte(newContent), nil
	})
}

const xgoFuncStartMarker = "// ==start xgo func=="
const xgoFuncEndMarker = "// ==end xgo func=="

var funcTemplatePath = []string{"runtime", "core", "func_template.go"}
var funcPath = []string{"runtime", "core", "func.go"}
var xgoTrapTemplatePath = []string{"runtime", "internal", "runtime", "xgo_trap_template.go"}
var legacyRuntimeLinkTemplatePath = []string{"instrument", "instrument_xgo_runtime", "runtime_link_template_legacy_1_1_0.go"}

func replaceXgoFunc(content string, prefix string) (string, error) {
	cstart, cend, cTypeDefStartIdx, cLastBraceIdx, err := getFuncInfoRange(content)
	if err != nil {
		return "", err
	}

	funcTemplateFile := filepath.Join(funcTemplatePath...)
	templateContentBytes, err := fileutil.ReadFile(funcTemplateFile)
	if err != nil {
		return "", err
	}
	templateContent := string(templateContentBytes)
	start, end, typeDefStartIdx, lastBraceIdx, err := getFuncInfoRange(templateContent)
	if err != nil {
		return "", err
	}
	templateFuncContent := templateContent[typeDefStartIdx : lastBraceIdx+1]
	_ = cstart
	_ = cend
	_ = start
	_ = end

	funcContent := strings.ReplaceAll(templateFuncContent, "__PREFIX__", prefix)

	return content[:cTypeDefStartIdx] + funcContent + content[cLastBraceIdx+1:], nil
}

func getFuncInfoRange(content string) (start int, end int, typeDefStartIdx int, lastBraceIdx int, err error) {
	n := len(content)
	start = strings.Index(content, xgoFuncStartMarker)
	if start < 0 {
		return -1, -1, -1, -1, fmt.Errorf("start marker not found: %s", xgoFuncStartMarker)
	}
	base := start + len(xgoFuncStartMarker)
	for i := base; i < n; i++ {
		if isSpace(content[i]) {
			continue
		}
		typeDefStartIdx = i
		break
	}

	end = strings.Index(content[typeDefStartIdx:], xgoFuncEndMarker)
	if end < 0 {
		return -1, -1, -1, -1, fmt.Errorf("end marker not found: %s", xgoFuncEndMarker)
	}
	end += typeDefStartIdx
	lastBraceIdx = strings.LastIndex(content[:end], "}")
	if lastBraceIdx < 0 {
		return -1, -1, -1, -1, fmt.Errorf("brace not found")
	}
	return start, end, typeDefStartIdx, lastBraceIdx, nil
}

func isSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}
