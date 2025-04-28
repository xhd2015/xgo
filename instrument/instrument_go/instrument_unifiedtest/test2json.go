package instrument_unifiedtest

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/xhd2015/xgo/instrument/patch"
	"github.com/xhd2015/xgo/support/goinfo"
)

// src/cmd/internal/test2json/test2json.go
var test2jsonFile = patch.FilePath{"src", "cmd", "internal", "test2json", "test2json.go"}

func instrumentTest2json(goroot string, goVersion *goinfo.GoVersion) error {
	fileName := test2jsonFile.JoinPrefix()
	if goVersion.Major != 1 || (goVersion.Minor < 17 || goVersion.Minor > 24) {
		return fmt.Errorf("%s unsupported version: go%d.%d, available: go1.17~go1.24", fileName, goVersion.Major, goVersion.Minor)
	}
	file := filepath.Join(goroot, fileName)

	return patch.EditFile(file, func(content string) (string, error) {
		// func (c *Converter) writeEvent(e *event) {
		content = patch.UpdateContent(content,
			"/*<begin declare_XgoGetTestName>*/",
			"/*<end declare_XgoGetTestName>*/",
			[]string{
				"func (c *Converter) writeEvent(e *event) {",
			},
			0,
			patch.UpdatePosition_Before,
			"var XgoGetTestName func(testPkgPath string, name string) (pkgPath string, testName string);",
		)

		content = patch.UpdateContent(content,
			"/*<begin writeEvent_replace_pkg>*/",
			"/*<end writeEvent_replace_pkg>*/",
			[]string{
				"func (c *Converter) writeEvent(e *event) {",
				"js, err := json.Marshal(e)",
			},
			1,
			patch.UpdatePosition_Before,
			strings.Join([]string{
				`if XgoGetTestName != nil && e.Test != "" {`,
				"  tpkg, tname := XgoGetTestName(e.Package, e.Test);",
				`  if tpkg != "" {`,
				"    e.Package = tpkg;",
				"  };",
				`  if tname != "" {`,
				"    e.Test = tname;",
				"  };",
				"};",
			}, ""),
		)

		return content, nil
	})
}
