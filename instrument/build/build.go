package build

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/xhd2015/xgo/support/cmd"
	"github.com/xhd2015/xgo/support/osinfo"
)

func BuildNativeBinary(goroot string, dir string, flags []string, outputDir string, outputName string, arg string) (string, error) {
	if arg == "" {
		return "", fmt.Errorf("requires arg")
	}
	// on Windows, rename failed with
	//  rename C:\Users\runneradmin\.xgo\go-instrument\go1.24.2_C_ho_wi_go_1._x6_339c415e\go1.24.2\bin\go.exe C:\Users\runneradmin\.xgo\go-instrument\go1.24.2_C_ho_wi_go_1._x6_339c415e\go1.24.2\bin\go.exe.bak: Access is denied.
	// const USE_RENAME = false

	origGo := filepath.Join(goroot, "bin", "go"+osinfo.EXE_SUFFIX)
	outputFile := filepath.Join(outputDir, outputName+osinfo.EXE_SUFFIX)

	runArgs := []string{"build", "-o", outputFile}
	runArgs = append(runArgs, flags...)
	runArgs = append(runArgs, arg)

	gorootEnv := EnvForNative(os.Environ(), goroot)
	gorootEnv = append(gorootEnv, "GO_BYPASS_XGO=true")
	err := cmd.Dir(dir).
		NoInheritEnv().
		Env(gorootEnv).
		Run(origGo, runArgs...)
	if err != nil {
		return "", err
	}
	return outputFile, nil
}

func RebuildGoBinary(goroot string) error {
	_, err := BuildGoBinray(goroot, nil, "go")
	return err
}

func RebuildGoToolCompile(goroot string) error {
	_, err := BuildToolBinray(goroot, nil, "./cmd/compile", "compile")
	return err
}

func RebuildGoToolCover(goroot string) error {
	_, err := BuildToolBinray(goroot, nil, "./cmd/cover", "cover")
	return err
}

func BuildGoBinray(goroot string, flags []string, outputName string) (string, error) {
	goroot, err := filepath.Abs(goroot)
	if err != nil {
		return "", err
	}
	outputFile, err := BuildNativeBinary(goroot, filepath.Join(goroot, "src"), flags, filepath.Join(goroot, "bin"), outputName, "./cmd/go")
	if err != nil {
		return "", err
	}
	return outputFile, nil
}

func BuildToolBinray(goroot string, flags []string, arg string, outputName string) (string, error) {
	goroot, err := filepath.Abs(goroot)
	if err != nil {
		return "", err
	}
	toolPath, err := GetToolPath(goroot)
	if err != nil {
		return "", err
	}
	outputFile, err := BuildNativeBinary(goroot, filepath.Join(goroot, "src"), flags, toolPath, outputName, arg)
	if err != nil {
		return "", err
	}
	return outputFile, nil
}

// debug
func BuildGoDebugBinary(goroot string) (string, error) {
	return BuildGoBinray(goroot, []string{"-gcflags=all=-N -l"}, "go.debug")
}

func BuildGoToolCompileDebugBinary(goroot string) (string, error) {
	return BuildToolBinray(goroot, []string{"-gcflags=all=-N -l"}, "./cmd/compile", "compile.debug")
}
