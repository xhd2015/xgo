package instrument_runtime

import (
	"fmt"

	"github.com/xhd2015/xgo/instrument/patch"
)

func instrumentRuntime2(goroot string, goMajor int, goMinor int) error {
	if goMajor != 1 || (goMinor < 17 || goMinor > 24) {
		// src/runtime/runtime2.go
		return fmt.Errorf("%s unsupported version: go%d.%d, available: go1.17~go1.24", runtime2Path.JoinPrefix(""), goMajor, goMinor)
	}
	runtime2File := runtime2Path.JoinPrefix(goroot)

	return patch.EditFile(runtime2File, func(content string) (string, error) {
		content = patch.UpdateContent(content,
			"/*<begin instrument_runtime2_xgo_g>*/",
			"/*<end instrument_runtime2_xgo_g>*/",
			[]string{
				"type g struct {",
				"}\n",
			},
			1,
			patch.UpdatePosition_Before,
			"__xgo_g __xgo_g;",
		)
		return content, nil
	})
}
