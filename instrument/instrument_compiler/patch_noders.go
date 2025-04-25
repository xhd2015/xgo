package instrument_compiler

import (
	"fmt"

	"github.com/xhd2015/xgo/instrument/patch"
	"github.com/xhd2015/xgo/support/goinfo"
)

var NoderFile = patch.FilePath{"src", "cmd", "compile", "internal", "noder", "noder.go"}
var NoderFile16 = patch.FilePath{"src", "cmd", "compile", "internal", "gc", "noder.go"}

func PatchNoder(goroot string, goVersion *goinfo.GoVersion) error {
	filePath := NoderFile
	var noderFiles string
	if goVersion.Major == goinfo.GO_MAJOR_1 {
		minor := goVersion.Minor
		if minor == goinfo.GO_VERSION_16 {
			filePath = NoderFile16
			noderFiles = NoderFiles_1_17
		} else if minor == goinfo.GO_VERSION_17 {
			noderFiles = NoderFiles_1_17
		} else if minor == goinfo.GO_VERSION_18 {
			noderFiles = NoderFiles_1_17
		} else if minor == goinfo.GO_VERSION_19 {
			noderFiles = NoderFiles_1_17
		} else if minor == goinfo.GO_VERSION_20 {
			noderFiles = NoderFiles_1_20
		} else if minor >= goinfo.GO_VERSION_21 && minor <= goinfo.GO_VERSION_24 {
			noderFiles = NoderFiles_1_21
		} else {
			return fmt.Errorf("patch compiler noder:unsupported: %v", goVersion)
		}
	}
	if noderFiles == "" {
		return fmt.Errorf("patch compiler noder:unsupported: %v", goVersion)
	}

	return patch.EditFile(filePath.JoinPrefix(goroot), func(content string) (string, error) {
		content = patch.AddImportAfterName(content,
			"/*<begin import_xgo_syntax>*/",
			"/*<end import_xgo_syntax>*/",
			"xgo_syntax",
			"cmd/compile/internal/xgo_rewrite_internal/patch/syntax",
		)
		content = patch.AddImportAfterName(content,
			"/*<begin import_io>*/",
			"/*<end import_io>*/",
			"xgo_io",
			"io",
		)

		var anchors []string
		var idx int
		if goVersion.Major == 1 && goVersion.Minor <= 16 {
			anchors = []string{
				"func parseFiles(filenames []string)",
				"for _, p := range noders {",
				"localpkg.Height = myheight",
				"\n",
			}
			idx = 3
		} else {
			anchors = []string{
				`func LoadPackage`,
				`for _, p := range noders {`,
				`base.Timer.AddEvent(int64(lines), "lines")`,
				"\n",
			}
			idx = 3
		}
		content = patch.UpdateContent(content,
			"/*<begin file_autogen>*/",
			"/*<end file_autogen>*/",
			anchors,
			idx,
			patch.UpdatePosition_After,
			noderFiles,
		)

		// expose the trimFilename func for recording
		if goVersion.Major == 1 && goVersion.Minor <= 17 {
			content = patch.UpdateContent(content,
				"/*<begin expose_abs_filename>*/", "/*<end expose_abs_filename>*/",
				[]string{
					`func absFilename(name string) string {`,
				},
				0,
				patch.UpdatePosition_Before,
				"func init(){ xgo_syntax.AbsFilename = absFilename;};",
			)
		} else {
			content = patch.UpdateContentLines(content,
				"/*<begin expose_trim_filename>*/", "/*<end expose_trim_filename>*/",
				[]string{
					`func trimFilename(b *syntax.PosBase) string {`,
				},
				0,
				patch.UpdatePosition_Before,
				"func init(){ xgo_syntax.TrimFilename = trimFilename;};",
			)
		}

		return content, nil
	})
}
