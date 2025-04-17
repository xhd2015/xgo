package instrument_compiler

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/xhd2015/xgo/instrument/build"
	"github.com/xhd2015/xgo/instrument/patch"
	"github.com/xhd2015/xgo/support/goinfo"
)

var CompilerRuntimeDefFile = patch.FilePath{"src", "cmd", "compile", "internal", "typecheck", "_builtin", "runtime.go"}
var CompilerRuntimeDefFile18 = patch.FilePath{"src", "cmd", "compile", "internal", "typecheck", "builtin", "runtime.go"}
var CompilerRuntimeDefFile16 = patch.FilePath{"src", "cmd", "compile", "internal", "gc", "builtin", "runtime.go"}

// cd src/cmd/compile/internal/typecheck && go run mkbuiltin.go
func MkBuiltin(origGoroot string, outputGoroot string, goVersion *goinfo.GoVersion, extraDefs string) error {
	err := PrepareRuntimeDefs(outputGoroot, goVersion, extraDefs)
	if err != nil {
		return err
	}

	// run mkbuiltin
	cmd := exec.Command(filepath.Join(origGoroot, "bin", "go"), "run", "mkbuiltin.go")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Dir = GetMkBuiltinDir(outputGoroot, goVersion)
	cmd.Env = build.EnvForNative(os.Environ(), origGoroot)

	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func GetMkBuiltinDir(goroot string, goVersion *goinfo.GoVersion) string {
	var paths []string
	if goVersion.Major > 1 || (goVersion.Major == 1 && goVersion.Minor > 16) {
		paths = []string{goroot, "src", "cmd", "compile", "internal", "typecheck"}
	} else {
		paths = []string{goroot, "src", "cmd", "compile", "internal", "gc"}
	}
	return filepath.Join(paths...)
}

func PrepareRuntimeDefs(goRoot string, goVersion *goinfo.GoVersion, extraDef string) error {
	runtimeDefFiles := []string(CompilerRuntimeDefFile)
	if goVersion.Major == 1 && goVersion.Minor <= 19 {
		if goVersion.Minor > 16 {
			// in go1.19 and below, builtin has no _ prefix
			runtimeDefFiles = []string(CompilerRuntimeDefFile18)
		} else {
			runtimeDefFiles = []string(CompilerRuntimeDefFile16)
		}
	}
	runtimeDefFile := filepath.Join(runtimeDefFiles...)
	fullFile := filepath.Join(goRoot, runtimeDefFile)

	return patch.EditFile(fullFile, func(content string) (string, error) {
		content = patch.UpdateContent(content,
			`/*<begin extra_runtime_func>*/`, `/*<end extra_runtime_func>*/`,
			[]string{
				`var x86HasFMA bool`,
				`var armHasVFPv4 bool`,
				`var arm64HasATOMICS bool`,
			},
			2,
			patch.UpdatePosition_After,
			extraDef+"\n",
		)
		return content, nil
	})
}
