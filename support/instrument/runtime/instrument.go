package runtime

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/xhd2015/xgo/support/edit"
	"github.com/xhd2015/xgo/support/goinfo"
)

//go:embed xgo_trap.go.txt
var xgoTrapFile string

const instrumentMark = "v0.0.1"

// only support go1.19 for now
func InstrumentRuntime(goroot string) error {
	runtimeDir := filepath.Join(goroot, "src", "runtime")
	markFile := filepath.Join(runtimeDir, "xgo_trap_instrument_mark.txt")
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
	if goVersion.Major != 1 || goVersion.Minor != 19 {
		return fmt.Errorf("unsupported go version: %s, available: go1.19", goVersionStr)
	}

	runtime2File := filepath.Join(runtimeDir, "runtime2.go")
	runtime2Bytes, err := os.ReadFile(runtime2File)
	if err != nil {
		return err
	}
	const typegDef = "type g struct {"
	const typegDefEnd = "}\n"
	const insertDef = "__xgo_g *__xgo_g"
	idx := bytes.Index(runtime2Bytes, []byte(typegDef))
	if idx == -1 {
		return fmt.Errorf("%s not found", typegDef)
	}
	structEnd := bytes.Index(runtime2Bytes[idx+len(typegDef):], []byte(typegDefEnd))
	if structEnd == -1 {
		return fmt.Errorf("%s missing %s", typegDef, typegDefEnd)
	}
	structEnd += idx + len(typegDef)
	if !bytes.Contains(runtime2Bytes[idx:structEnd], []byte(insertDef)) {
		// allow re-entrant instrument

		runtime2Edit := edit.NewBuffer(runtime2Bytes)
		runtime2Edit.Replace(structEnd, structEnd+len(typegDefEnd), "    __xgo_g *__xgo_g\n"+typegDefEnd)

		err = os.WriteFile(runtime2File, runtime2Edit.Bytes(), 0644)
		if err != nil {
			return err
		}
	}

	trapFile := filepath.Join(runtimeDir, "xgo_trap.go")
	err = os.WriteFile(trapFile, []byte(xgoTrapFile), 0644)
	if err != nil {
		return err
	}

	return os.WriteFile(markFile, []byte(instrumentMark), 0644)
}
