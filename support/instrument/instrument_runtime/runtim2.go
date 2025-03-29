package instrument_runtime

import (
	"fmt"

	"github.com/xhd2015/xgo/support/instrument/patch"
)

func instrumentRuntime2(goroot string, goMajor int, goMinor int) error {
	if goMajor != 1 || (goMinor != 18 && goMinor != 19 && goMinor != 20 && goMinor != 21 && goMinor != 22 && goMinor != 23 && goMinor != 24) {
		// src/runtime/runtime2.go
		return fmt.Errorf("%s unsupported version: go%d.%d, available: go1.18~go1.24", runtime2Path.JoinPrefix(""), goMajor, goMinor)
	}
	runtime2File := runtime2Path.JoinPrefix(goroot)

	// bytes, err := fileutil.ReadFile(runtime2File)
	// if err != nil {
	// 	return err
	// }
	// before := string(bytes)
	// content := patch.CleanPatch(before)

	// diff := assert.Diff(before, content)
	// fmt.Fprintln(os.Stderr, "diff:", diff)
	// // panic("debug")
	// time.Sleep(100 * time.Second)

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
