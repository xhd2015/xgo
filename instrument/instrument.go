package instrument

import (
	"github.com/xhd2015/xgo/instrument/edit"
	"github.com/xhd2015/xgo/instrument/instrument_func"
	"github.com/xhd2015/xgo/instrument/instrument_go"
	"github.com/xhd2015/xgo/instrument/instrument_runtime"
	"github.com/xhd2015/xgo/instrument/instrument_var"
	"github.com/xhd2015/xgo/instrument/instrument_xgo_runtime"
	"github.com/xhd2015/xgo/instrument/overlay"
	"github.com/xhd2015/xgo/support/goinfo"
)

func LinkXgoRuntime(projectDir string, xgoRuntimeModuleDir string, overlayFS overlay.Overlay, mod string, modfile string, xgoVersion string, xgoRevision string, xgoNumber int, collectTestTrace bool, collectTestTraceDir string, overrideContent func(absFile overlay.AbsFile, content string)) (*edit.Packages, error) {
	return instrument_xgo_runtime.LinkXgoRuntime(projectDir, xgoRuntimeModuleDir, overlayFS, mod, modfile, xgoVersion, xgoRevision, xgoNumber, collectTestTrace, collectTestTraceDir, overrideContent)
}

func InstrumentVarTrap(packages *edit.Packages) error {
	instrument_var.Instrument(packages)
	return nil
}

func InstrumentFuncTrap(packages *edit.Packages) error {
	for _, pkg := range packages.Packages {
		pkgPath := pkg.LoadPackage.GoPackage.ImportPath
		for _, file := range pkg.Files {
			instrument_func.InjectRuntimeTrap(file.Edit, pkgPath, file.File.Syntax, file.Index)
		}
	}
	return nil
}

func InstrumentGo(goroot string, goVersion *goinfo.GoVersion) error {
	return instrument_go.InstrumentGo(goroot, goVersion)
}

func InstrumentRuntime(goroot string, goVersion *goinfo.GoVersion, opts instrument_runtime.InstrumentRuntimeOptions) error {
	return instrument_runtime.InstrumentRuntime(goroot, goVersion, opts)
}
