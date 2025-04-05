package instrument_xgo_runtime

import (
	_ "embed"
	"fmt"
	"strconv"
	"strings"

	"github.com/xhd2015/xgo/instrument/patch"
)

//go:embed runtime_link_template.go
var runtimeLinkTemplate string

func GetLinkRuntimeCode() string {
	code, err := patch.RemoveBuildIgnore(runtimeLinkTemplate)
	if err != nil {
		panic(err)
	}
	return code
}

func ReplaceActualXgoVersion(versionCode string, xgoVersion string, xgoRevision string, xgoNumber int) string {
	versionCode = replaceByLine(versionCode, `const XGO_VERSION = `, `const XGO_VERSION = "`+xgoVersion+`"`)
	versionCode = replaceByLine(versionCode, `const XGO_REVISION = `, `const XGO_REVISION = "`+xgoRevision+`"`)
	versionCode = replaceByLine(versionCode, `const XGO_NUMBER = `, `const XGO_NUMBER = `+strconv.Itoa(xgoNumber))
	return versionCode
}

func ParseCoreVersion(versionCode string) (string, error) {
	anchor := `const VERSION =`
	idx := strings.Index(versionCode, anchor)
	if idx < 0 {
		return "", fmt.Errorf("VERSION not found")
	}
	idx += len(anchor)
	endIdx := strings.Index(versionCode[idx:], "\n")
	if endIdx < 0 {
		return "", fmt.Errorf("VERSION not found")
	}
	endIdx += idx
	versionStr := strings.TrimSpace(versionCode[idx:endIdx])
	ver, err := strconv.Unquote(versionStr)
	if err != nil {
		return "", fmt.Errorf("parse VERSION: %w", err)
	}
	if ver == "" {
		return "", fmt.Errorf("VERSION is empty")
	}
	return ver, nil
}

func InjectFlags(flagsCode string, collectTestTrace bool, collectTestTraceDir string) string {
	flagsCode = replaceByLine(flagsCode, `const COLLECT_TEST_TRACE = `, fmt.Sprintf(`const COLLECT_TEST_TRACE = %t`, collectTestTrace))
	flagsCode = replaceByLine(flagsCode, `const COLLECT_TEST_TRACE_DIR = `, fmt.Sprintf(`const COLLECT_TEST_TRACE_DIR = %q`, collectTestTraceDir))
	return flagsCode
}

// replaceByLine allows re-entrant replacement
func replaceByLine(code string, linePattern string, replacement string) string {
	idx := strings.Index(code, linePattern)
	if idx == -1 {
		return code
	}
	base := idx + len(linePattern)
	endIdx := strings.Index(code[base:], "\n")
	if endIdx == -1 {
		return code
	}
	endIdx += base
	return code[:idx] + replacement + "\n" + code[endIdx:]
}
