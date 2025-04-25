package instrument_compiler

import (
	"github.com/xhd2015/xgo/instrument/build"
	"github.com/xhd2015/xgo/support/goinfo"
)

func BuildInstrument(origGoroot string, goroot string, goVersion *goinfo.GoVersion, xgoSrc string, forceReset bool, syncWithLink bool) error {
	err := patchCompiler(origGoroot, goroot, goVersion, xgoSrc, forceReset, syncWithLink)
	if err != nil {
		return err
	}

	return build.RebuildGoToolCompile(goroot)
}
