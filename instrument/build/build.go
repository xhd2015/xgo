package build

import (
	"os"
	"path/filepath"

	"github.com/xhd2015/xgo/support/cmd"
	"github.com/xhd2015/xgo/support/osinfo"
)

func BuildNativeBinary(goroot string, dir string, outputDir string, outputName string, arg string) error {
	// on Windows, rename failed with
	//  rename C:\Users\runneradmin\.xgo\go-instrument\go1.24.2_C_ho_wi_go_1._x6_339c415e\go1.24.2\bin\go.exe C:\Users\runneradmin\.xgo\go-instrument\go1.24.2_C_ho_wi_go_1._x6_339c415e\go1.24.2\bin\go.exe.bak: Access is denied.
	// const USE_RENAME = false

	origGo := filepath.Join(goroot, "bin", "go"+osinfo.EXE_SUFFIX)
	origFile := filepath.Join(outputDir, outputName+osinfo.EXE_SUFFIX)

	gorootEnv := EnvForNative(os.Environ(), goroot)
	gorootEnv = append(gorootEnv, "GO_BYPASS_XGO=true")
	return cmd.Dir(dir).
		NoInheritEnv().
		Env(gorootEnv).
		Run(origGo, "build", "-o", origFile, arg)
}
