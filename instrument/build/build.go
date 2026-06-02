package build

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/xhd2015/xgo/support/cmd"
	"github.com/xhd2015/xgo/support/goinfo"
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

// NeedExternalLinker reports whether binaries should be built with
// -linkmode=external on the current host for the current target.
//
// On macOS 26+ (arm64), dyld requires all executables to have an LC_UUID load
// command. The Go internal linker added LC_UUID emission in Go 1.22.9 / 1.23.3
// / 1.24+. For older Go versions, we use -linkmode=external to invoke the
// system linker (clang/ld), which always emits LC_UUID.
//
// Empirically, go1.22.12 (which should include the 1.22.9 backport) still
// produces binaries without LC_UUID on some systems. To be safe, we apply
// external linking for all Go versions below 1.23.
//
// External linking only works when the target OS matches the host OS (both
// darwin). When cross-compiling to e.g. linux, the macOS system linker cannot
// produce the target binary, so this returns false.
//
// See: https://github.com/golang/go/issues/68678, #78012
func NeedExternalLinker(goVersion *goinfo.GoVersion) bool {
	if GetHostGOOS() != "darwin" || GetHostGOARCH() != "arm64" {
		return false
	}
	if GetTargetGOOS() != "darwin" {
		return false
	}
	if goVersion == nil {
		return false
	}
	return goVersion.Major == 1 && goVersion.Minor < 23
}

func ExternalLinkerFlags(goVersion *goinfo.GoVersion) []string {
	if !NeedExternalLinker(goVersion) {
		return nil
	}
	return []string{"-ldflags=-linkmode=external"}
}

func RebuildGoBinary(goroot string, goVersion *goinfo.GoVersion) error {
	flags := append([]string{"-a"}, ExternalLinkerFlags(goVersion)...)
	_, err := BuildGoBinray(goroot, flags, "go")
	return err
}

func RebuildGoToolCompile(goroot string, goVersion *goinfo.GoVersion) error {
	flags := append([]string{"-a"}, ExternalLinkerFlags(goVersion)...)
	_, err := BuildToolBinray(goroot, flags, "./cmd/compile", "compile")
	return err
}

func RebuildGoToolCover(goroot string, goVersion *goinfo.GoVersion) error {
	_, err := BuildToolBinray(goroot, ExternalLinkerFlags(goVersion), "./cmd/cover", "cover")
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
	return BuildGoBinray(goroot, []string{"-a", "-gcflags=all=-N -l"}, "go.debug")
}

func BuildGoToolCompileDebugBinary(goroot string) (string, error) {
	return BuildToolBinray(goroot, []string{"-a", "-gcflags=all=-N -l"}, "./cmd/compile", "compile.debug")
}
