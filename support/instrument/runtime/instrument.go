package runtime

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/xhd2015/xgo/support/goinfo"
	"github.com/xhd2015/xgo/support/instrument/patch"
)

//go:embed xgo_trap.go.txt
var xgoTrapFile string

const instrumentMark = "v0.0.1"

var instrumentMarkPath = patch.FilePath{"xgo_trap_instrument_mark.txt"}
var xgoTrapFilePath = patch.FilePath{"src", "runtime", "xgo_trap.go"}
var runtime2Path = patch.FilePath{"src", "runtime", "runtime2.go"}

var jsonEncodingPath = patch.FilePath{"src", "encoding", "json", "encode.go"}

// only support go1.19 for now
func InstrumentRuntime(goroot string) error {
	srcDir := filepath.Join(goroot, "src")

	srcStat, statErr := os.Stat(srcDir)
	if statErr != nil {
		if !os.IsNotExist(statErr) {
			return statErr
		}
		return fmt.Errorf("GOROOT/src does not exist, please use newer go distribution which contains src code for runtime: %w", statErr)
	}
	if !srcStat.IsDir() {
		return fmt.Errorf("GOROOT/src is not a directory, please use newer go distribution: %s", srcDir)
	}

	markFile := instrumentMarkPath.JoinPrefix(goroot)
	markContent, statErr := os.ReadFile(markFile)
	if statErr != nil {
		if !os.IsNotExist(statErr) {
			return statErr
		}
	}
	if string(markContent) == instrumentMark {
		return nil
	}
	goVersionStr, err := goinfo.GetGoVersionOutput(filepath.Join(goroot, "bin", "go"))
	if err != nil {
		return err
	}
	goVersion, err := goinfo.ParseGoVersion(goVersionStr)
	if err != nil {
		return err
	}

	err = instrumentRuntime2(goroot, goVersion.Major, goVersion.Minor)
	if err != nil {
		return fmt.Errorf("instrument runtime2: %w", err)
	}
	err = instrumentJsonEncoding(goroot, goVersion.Major, goVersion.Minor)
	if err != nil {
		return fmt.Errorf("instrument json encoding: %w", err)
	}

	err = os.WriteFile(xgoTrapFilePath.JoinPrefix(goroot), []byte(xgoTrapFile), 0644)
	if err != nil {
		return err
	}

	return os.WriteFile(markFile, []byte(instrumentMark), 0644)
}

func instrumentRuntime2(goroot string, goMajor int, goMinor int) error {
	if goMajor != 1 || goMinor != 19 {
		return fmt.Errorf("unsupported version: go%d.%d, available: go1.19", goMajor, goMinor)
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
			"__xgo_g *__xgo_g;",
		)
		return content, nil
	})
}

func instrumentJsonEncoding(goroot string, goMajor int, goMinor int) error {
	if goMajor != 1 || goMinor != 19 {
		return fmt.Errorf("unsupported version: go%d.%d, available: go1.19", goMajor, goMinor)
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
