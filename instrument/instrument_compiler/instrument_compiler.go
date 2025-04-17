package instrument_compiler

import (
	"path/filepath"

	"github.com/xhd2015/xgo/instrument/build"
	"github.com/xhd2015/xgo/support/goinfo"
)

func BuildInstrument(origGoroot string, goroot string, goVersion *goinfo.GoVersion, xgoSrc string, forceReset bool, syncWithLink bool) error {
	err := patchCompiler(origGoroot, goroot, goVersion, xgoSrc, forceReset, false)
	if err != nil {
		return err
	}

	toolPath, err := build.GetToolPath(goroot)
	if err != nil {
		return err
	}

	// build go command
	return build.BuildNativeBinary(goroot, filepath.Join(goroot, "src"), toolPath, "compile", "./cmd/compile")
}
