package instrument_go

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/xhd2015/xgo/instrument/build"
	"github.com/xhd2015/xgo/instrument/instrument_go/instrument_gc"
	"github.com/xhd2015/xgo/instrument/instrument_go/instrument_unifiedtest"
	"github.com/xhd2015/xgo/instrument/patch"
	"github.com/xhd2015/xgo/support/goinfo"
)

// instrument the `go` command to fix coverage with -overlay
// see https://github.com/xhd2015/xgo/issues/300

//go:embed xgo_main_template.go
var xgoMainTemplate string

//go:embed xgo_work_sum_template.go
var xgoWorkSumTemplate string

// src/cmd/go/internal/work/shell_test.go

var srcCmdGoPath = patch.FilePath{"src", "cmd", "go"}          // src/cmd/go
var internalWorkPath = srcCmdGoPath.Append("internal", "work") // src/cmd/go/internal/work

var mainFilePath = srcCmdGoPath.Append("main.go")
var xgoMainFilePath = srcCmdGoPath.Append("xgo_main.go")

func BuildInstrument(goroot string, goVersion *goinfo.GoVersion, allowDebugCompile bool) error {
	err := instrumentExec(goroot, goVersion)
	if err != nil {
		return err
	}
	err = instrumentGoMain(goroot, goVersion)
	if err != nil {
		return err
	}
	err = copyXgoMain(goroot)
	if err != nil {
		return err
	}
	err = copyWorkSum(goroot)
	if err != nil {
		return err
	}
	err = instrumentPkgLoad(goroot, goVersion)
	if err != nil {
		return err
	}
	if allowDebugCompile {
		err = instrument_gc.InstrumentGc(goroot, goVersion)
		if err != nil {
			return err
		}
	}

	// go test
	err = instrument_unifiedtest.Unify(goroot, goVersion)
	if err != nil {
		return err
	}

	// build go command
	return build.RebuildGoBinary(goroot)
}

func copyXgoMain(goroot string) error {
	code, err := patch.RemoveBuildIgnore(xgoMainTemplate)
	if err != nil {
		return err
	}
	xgoMainFile := xgoMainFilePath.JoinPrefix(goroot)
	return os.WriteFile(xgoMainFile, []byte(code), 0644)
}

func copyWorkSum(goroot string) error {
	code, err := patch.RemoveBuildIgnore(xgoWorkSumTemplate)
	if err != nil {
		return err
	}
	sumFile := xgoWorkSumFilePath.JoinPrefix(goroot)
	return os.WriteFile(sumFile, []byte(code), 0644)
}

func instrumentGoMain(goroot string, goVersion *goinfo.GoVersion) error {
	if goVersion.Major != 1 || (goVersion.Minor < 17 || goVersion.Minor > 24) {
		// src/cmd/go/main.go
		return fmt.Errorf("%s unsupported version: go%d.%d, available: go1.17~go1.24", execFilePath.JoinPrefix(""), goVersion.Major, goVersion.Minor)
	}
	mainFile := mainFilePath.JoinPrefix(goroot)
	return patch.EditFile(mainFile, func(content string) (string, error) {
		content = patch.UpdateContent(content,
			"/*<begin call_xgo_precheck>*/",
			"/*<end call_xgo_precheck>*/",
			[]string{
				"\nfunc main() {",
				// before the first command
				"if args[0] ==",
			},
			1,
			patch.UpdatePosition_Before,
			"if xgoPrecheck(args[0], args[1:]) { return; };",
		)
		return content, nil
	})
}
