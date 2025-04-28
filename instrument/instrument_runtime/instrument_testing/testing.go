package instrument_testing

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/xhd2015/xgo/instrument/patch"
	"github.com/xhd2015/xgo/support/goinfo"
)

// src/testing/testing.go

var testingFile = patch.FilePath{"src", "testing", "testing.go"}

func Instrument(goroot string, goVersion *goinfo.GoVersion) error {
	fileName := testingFile.JoinPrefix()
	if goVersion.Major != 1 || (goVersion.Minor < 17 || goVersion.Minor > 24) {
		return fmt.Errorf("%s unsupported version: go%d.%d, available: go1.17~go1.24", fileName, goVersion.Major, goVersion.Minor)
	}

	return patch.EditFile(filepath.Join(goroot, fileName), func(content string) (string, error) {
		content = patch.UpdateContent(content,
			"/*<begin declare_XgoTestNameToDir>*/",
			"/*<end declare_XgoTestNameToDir>*/",
			[]string{
				"func runTests(",
			},
			0,
			patch.UpdatePosition_Before,
			"var XgoTestNameToDir map[string]string;",
		)

		// TODO: disable parallel because it is unsafe for os.Chdir
		content = patch.UpdateContent(content,
			"/*<begin check_XgoTestNameToDir>*/",
			"/*<end check_XgoTestNameToDir>*/",

			// tRunner(t, func(t *T) {
			// 	for _, test := range tests {
			// 		t.Run(test.Name, test.F)
			// 	}
			// })
			[]string{
				"func runTests(",
				"tRunner(t, func(t *T) {",
				"t.Run(test.Name,",
			},
			2,
			patch.UpdatePosition_Before,
			strings.Join([]string{
				`if xgoDir := XgoTestNameToDir[test.Name]; xgoDir != "" {`,
				`  os.Chdir(xgoDir);`,
				`};`,
			}, ""),
		)
		return content, nil
	})
}
