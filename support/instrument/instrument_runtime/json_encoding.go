package instrument_runtime

import (
	"fmt"

	"github.com/xhd2015/xgo/support/instrument/patch"
)

const enableCyclicDetect = false

func instrumentJsonEncoding(goroot string, goMajor int, goMinor int) error {
	if goMajor != 1 || (goMinor < 17 || goMinor > 24) {
		// src/encoding/json/encode.go
		return fmt.Errorf("%s unsupported version: go%d.%d, available: go1.17~go1.24", jsonEncodingPath.JoinPrefix(""), goMajor, goMinor)
	}
	return patch.EditFile(jsonEncodingPath.JoinPrefix(goroot), func(content string) (string, error) {
		content = patch.UpdateContent(content,
			"/*<begin instrument_json_encoding_import_runtime>*/",
			"/*<end instrument_json_encoding_import_runtime>*/",
			[]string{
				"package json",
				"\n",
			},
			0,
			patch.UpdatePosition_After,
			";import __xgo_runtime \"runtime\"",
		)
		content = patch.UpdateContent(content,
			"/*<begin instrument_json_encoding_unsupported_type_encoder>*/",
			"/*<end instrument_json_encoding_unsupported_type_encoder>*/",
			[]string{
				"func unsupportedTypeEncoder(",
				") {",
				"\n",
			},
			1,
			patch.UpdatePosition_After,
			unsupportedTypeIgnore,
		)

		if enableCyclicDetect {
			encoders := []string{"mapEncoder", "sliceEncoder", "ptrEncoder"}
			for _, encoder := range encoders {
				content = patch.UpdateContent(content,
					fmt.Sprintf("/*<begin instrument_json_encoding_ignore_cyclic_%s>*/", encoder),
					fmt.Sprintf("/*<end instrument_json_encoding_ignore_cyclic_%s>*/", encoder),
					[]string{
						fmt.Sprintf("%s) encode", encoder),
						"if",
						"e.ptrLevel > startDetectingCyclesAfter",
						"if _, ok := e.ptrSeen[ptr]; ok {",
						"\n",
					},
					3,
					patch.UpdatePosition_After,
					cyclicIgnore,
				)
			}
		}
		return content, nil
	})
}

const unsupportedTypeIgnore = `if __xgo_runtime.XgoIsLooseJsonMarshaling() {` +
	`    e.WriteString(fmt.Sprintf("{%q:%q}", v.Type().String(), "?"));` +
	`    return;` +
	`}`

const cyclicIgnore = `if __xgo_runtime.XgoIsLooseJsonMarshaling() {` +
	`   e.WriteString(fmt.Sprintf("{%q:%q}", v.Type().String(), "cyclic..."));` +
	`   return;` +
	`}`
