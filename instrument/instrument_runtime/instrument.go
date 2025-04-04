package instrument_runtime

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/xhd2015/xgo/instrument/patch"
	"github.com/xhd2015/xgo/support/goinfo"
)

//go:embed xgo_trap_template.go
var xgoTrapFile string

var instrumentMarkPath = patch.FilePath{"xgo_trap_instrument_mark.txt"}
var xgoTrapFilePath = patch.FilePath{"src", "runtime", "xgo_trap.go"}
var runtime2Path = patch.FilePath{"src", "runtime", "runtime2.go"}
var procPath = patch.FilePath{"src", "runtime", "proc.go"}

var jsonEncodingPath = patch.FilePath{"src", "encoding", "json", "encode.go"}

type InstrumentMode int

const (
	InstrumentMode_UseMark InstrumentMode = iota
	InstrumentMode_Force
	InstrumentMode_ForceAndIgnoreMark
)

type InstrumentRuntimeOptions struct {
	Mode                  InstrumentMode
	InstrumentVersionMark string
}

// only support go1.19 for now
func InstrumentRuntime(goroot string, goVersion *goinfo.GoVersion, opts InstrumentRuntimeOptions) error {
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

	instrumentMark := opts.InstrumentVersionMark
	if instrumentMark == "" {
		instrumentMark = "v0.0.1"
	}

	markFile := instrumentMarkPath.JoinPrefix(goroot)
	if opts.Mode == InstrumentMode_UseMark {
		markContent, statErr := os.ReadFile(markFile)
		if statErr != nil {
			if !os.IsNotExist(statErr) {
				return statErr
			}
		}
		if string(markContent) == instrumentMark {
			return nil
		}
	}

	err := instrumentRuntime2(goroot, goVersion.Major, goVersion.Minor)
	if err != nil {
		return fmt.Errorf("instrument runtime2: %w", err)
	}

	err = instrumentProc(goroot, goVersion)
	if err != nil {
		return fmt.Errorf("instrument proc: %w", err)
	}

	err = instrumentJsonEncoding(goroot, goVersion.Major, goVersion.Minor)
	if err != nil {
		return fmt.Errorf("instrument json encoding: %w", err)
	}

	// instrument xgo_trap.go
	xgoTrapFileStripped, err := patch.RemoveBuildIgnore(xgoTrapFile)
	if err != nil {
		return fmt.Errorf("remove build ignore: %w", err)
	}
	xgoTrapFileStripped = AppendGetFuncNameImpl(goVersion, xgoTrapFileStripped)

	err = os.WriteFile(xgoTrapFilePath.JoinPrefix(goroot), []byte(xgoTrapFileStripped), 0644)
	if err != nil {
		return err
	}

	if opts.Mode != InstrumentMode_ForceAndIgnoreMark {
		err := os.WriteFile(markFile, []byte(instrumentMark), 0644)
		if err != nil {
			return err
		}
	}
	return nil
}
