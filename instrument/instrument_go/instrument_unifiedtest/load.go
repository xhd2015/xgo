package instrument_unifiedtest

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/xhd2015/xgo/instrument/patch"
	"github.com/xhd2015/xgo/support/goinfo"
)

// src/cmd/go/internal/load/test.go
var loadTestFile = patch.FilePath{"src", "cmd", "go", "internal", "load", "test.go"}

func instrumentLoadTest(goroot string, goVersion *goinfo.GoVersion) error {
	fileName := loadTestFile.JoinPrefix()
	if goVersion.Major != 1 || (goVersion.Minor < 17 || goVersion.Minor > 26) {
		return fmt.Errorf("%s unsupported version: go%d.%d, available: go1.17~go1.26", fileName, goVersion.Major, goVersion.Minor)
	}
	file := filepath.Join(goroot, fileName)

	return patch.EditFile(file, func(content string) (string, error) {
		// In go1.26, TestPackagesAndErrors signature changed to include loaderstate param
		testPkgsAnchor := "func TestPackagesAndErrors(ctx context.Context,"
		if goVersion.Minor >= 26 {
			testPkgsAnchor = "func TestPackagesAndErrors(loaderstate *modload.State, ctx context.Context,"
		}
		content = patch.UpdateContent(content,
			"/*<begin declare_XgoAfterGenerateTestMain>*/",
			"/*<end declare_XgoAfterGenerateTestMain>*/",
			[]string{
				testPkgsAnchor,
			},
			0,
			patch.UpdatePosition_Before,
			"var XgoAfterGenerateTestMain func(p *Package, pmain *Package, code []byte) []byte;",
		)

		content = patch.UpdateContent(content,
			"/*<begin call_XgoAfterGenerateTestMain>*/",
			"/*<end call_XgoAfterGenerateTestMain>*/",
			[]string{
				testPkgsAnchor,
				", err := formatTestmain(t)",
			},
			1,
			patch.UpdatePosition_After,
			strings.Join([]string{
				";if err==nil && XgoAfterGenerateTestMain!=nil {",
				"  data = XgoAfterGenerateTestMain(p, pmain, data);",
				"};",
			}, ""),
		)

		// loadTestFuncs
		content = patch.UpdateContent(content,
			"/*<begin export_XgoLoadTestFuncs>*/",
			"/*<end export_XgoLoadTestFuncs>*/",
			[]string{
				"func loadTestFuncs(ptest *Package)",
			},
			0,
			patch.UpdatePosition_Before,
			strings.Join([]string{
				"type XgoTestFunc = testFunc;",
				"type XgoTestFuncs struct {",
				"  Tests    []XgoTestFunc;",
				"  TestMain *XgoTestFunc;",
				"  Package  *Package;",
				"};",
				"func XgoLoadTestFuncs(ptest *Package) (*XgoTestFuncs, error) {",
				"  res, err := loadTestFuncs(ptest);",
				"  if err != nil {",
				"    return nil, err;",
				"  };",
				"  return &XgoTestFuncs{",
				"    Tests:    res.Tests,",
				"    TestMain: res.TestMain,",
				"    Package:  ptest,",
				"  }, nil;",
				"};",
			}, ""),
		)

		return content, nil
	})
}
