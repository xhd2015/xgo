package instrument_runtime

import (
	"fmt"
	"strings"

	"github.com/xhd2015/xgo/instrument/constants"
	"github.com/xhd2015/xgo/instrument/patch"
)

var timePath = patch.FilePath{"src", "time", "time.go"}
var timeSleepPath = patch.FilePath{"src", "time", "sleep.go"}
var runtimeTimePath = patch.FilePath{"src", "runtime", "time.go"}

func instrumentTimeNow(goroot string, goMajor int, goMinor int) error {
	if goMajor != 1 || (goMinor < 17 || goMinor > 24) {
		// src/time/time.go
		return fmt.Errorf("%s unsupported version: go%d.%d, available: go1.17~go1.24", timePath.JoinPrefix(""), goMajor, goMinor)
	}
	timeFile := timePath.JoinPrefix(goroot)
	return patch.EditFile(timeFile, func(content string) (string, error) {
		const anchor = "func Now() Time {"
		const endAnchor = "\n}\n"
		idx := strings.Index(content, anchor)
		if idx < 0 {
			panic(fmt.Sprintf("%s not found", anchor))
		}
		base := idx + len(anchor)
		endIdx := strings.Index(content[base:], endAnchor)
		if endIdx < 0 {
			panic(fmt.Sprintf("end %s of %s not found", endAnchor, anchor))
		}
		endIdx += base + len(endAnchor)
		body := content[idx:endIdx]

		// insert to end of file
		xgoBody := strings.ReplaceAll(body, anchor, "func "+constants.XGO_REAL_NOW+"() Time {")
		xgoBodyWithMarker := "/*<begin add_xgo_real_time_now>*/" + xgoBody + "/*<end add_xgo_real_time_now>*/"
		if !strings.HasSuffix(content, "\n") {
			xgoBodyWithMarker = "\n" + xgoBodyWithMarker
		}
		return content + xgoBodyWithMarker, nil
	})
}

func instrumentTimeSleep(goroot string, goMajor int, goMinor int) error {
	if goMajor != 1 || (goMinor < 17 || goMinor > 24) {
		// src/time/sleep.go
		return fmt.Errorf("%s unsupported version: go%d.%d, available: go1.17~go1.24", timeSleepPath.JoinPrefix(""), goMajor, goMinor)
	}
	timeFile := timeSleepPath.JoinPrefix(goroot)
	return patch.EditFile(timeFile, func(content string) (string, error) {
		anchor := "func Sleep(d Duration)"
		replaceWith := "func " + constants.XGO_REAL_SLEEP + "(d Duration);func /*xgo_instrumented*/Sleep(d Duration){ " + constants.XGO_REAL_SLEEP + "(d); }"
		idx := strings.Index(content, anchor)
		if idx < 0 {
			// already instrumented
			if strings.Index(content, replaceWith) > 0 {
				return content, nil
			}
			panic(fmt.Sprintf("%s not found", anchor))
		}
		content = content[:idx] + replaceWith + content[idx+len(anchor):]
		return content, nil
	})
}

func instrumentRuntimeTimeSleep(goroot string, goMajor int, goMinor int) error {
	if goMajor != 1 || (goMinor < 17 || goMinor > 24) {
		// src/runtime/time.go
		return fmt.Errorf("%s unsupported version: go%d.%d, available: go1.17~go1.24", runtimeTimePath.JoinPrefix(""), goMajor, goMinor)
	}
	timeFile := runtimeTimePath.JoinPrefix(goroot)
	return patch.EditFile(timeFile, func(content string) (string, error) {
		anchor := "//go:linkname timeSleep time.Sleep"
		replaceWith := "//go:linkname timeSleep time." + constants.XGO_REAL_SLEEP
		idx := strings.Index(content, anchor)
		if idx < 0 {
			// already instrumented
			if strings.Index(content, replaceWith) > 0 {
				return content, nil
			}
			panic(fmt.Sprintf("%s not found", anchor))
		}
		content = content[:idx] + replaceWith + content[idx+len(anchor):]

		return content, nil
	})
}
