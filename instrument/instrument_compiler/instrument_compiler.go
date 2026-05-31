package instrument_compiler

import (
	"github.com/xhd2015/xgo/instrument/build"
	"github.com/xhd2015/xgo/support/goinfo"
)

func BuildInstrument(origGoroot string, goroot string, goVersion *goinfo.GoVersion, xgoSrc string, forceReset bool, syncWithLink bool, skipRebuildCompiler bool) error {
	err := patchCompiler(origGoroot, goroot, goVersion, xgoSrc, forceReset, syncWithLink)
	if err != nil {
		return err
	}

	if skipRebuildCompiler {
		return nil
	}
	return build.RebuildGoToolCompile(goroot)
}
