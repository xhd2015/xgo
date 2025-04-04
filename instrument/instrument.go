package instrument

import (
	"fmt"
	"strings"

	"github.com/xhd2015/xgo/instrument/constants"
	"github.com/xhd2015/xgo/instrument/edit"
	"github.com/xhd2015/xgo/instrument/instrument_func"
	"github.com/xhd2015/xgo/instrument/instrument_go"
	"github.com/xhd2015/xgo/instrument/instrument_runtime"
	"github.com/xhd2015/xgo/instrument/instrument_var"
	"github.com/xhd2015/xgo/instrument/instrument_xgo_runtime"
	"github.com/xhd2015/xgo/instrument/load"
	"github.com/xhd2015/xgo/instrument/overlay"
	"github.com/xhd2015/xgo/support/goinfo"
)

var ErrLinkFileNotFound = fmt.Errorf("xgo: link file not found")

// create an overlay: abs file -> content
type Overlay map[string]string

func LinkXgoRuntime(projectDir string, xgoRuntimeModuleDir string, overlayFS overlay.Overlay, mod string, modfile string, xgoVersion string, xgoRevision string, xgoNumber int, collectTestTrace bool, collectTestTraceDir string, overrideContent func(absFile overlay.AbsFile, content string)) (*edit.Packages, error) {
	opts := load.LoadOptions{
		Dir:     projectDir,
		Overlay: overlayFS,
		Mod:     mod,
		ModFile: modfile,
	}
	if xgoRuntimeModuleDir != "" {
		// xgo runtime is replaced in a separate module
		// so we need to load packages from the separate module
		opts = load.LoadOptions{
			Dir: xgoRuntimeModuleDir,
		}
	}
	packages, err := load.LoadPackages([]string{
		constants.RUNTIME_INTERNAL_RUNTIME_PKG,
		constants.RUNTIME_CORE_PKG,
		constants.RUNTIME_TRAP_FLAGS_PKG,
		constants.RUNTIME_TRACE_PKG,
		constants.RUNTIME_FUNCTAB_PKG,
	}, opts)
	if err != nil {
		// TODO: handle the case where error indicates the package is not found
		return nil, err
	}
	editPackages := edit.Edit(packages)
	var foundLink bool
	for _, pkg := range editPackages.Packages {
		importPath := pkg.LoadPackage.GoPackage.ImportPath
		if !strings.HasPrefix(importPath, constants.RUNTIME_MODULE) || importPath[len(constants.RUNTIME_MODULE)] != '/' {
			continue
		}
		n := len(constants.RUNTIME_MODULE) + 1
		suffixPkg := importPath[n:]
		if suffixPkg == constants.RUNTIME_FUNCTAB_PKG[n:] {
			// only for lookup
			continue
		}
		for _, file := range pkg.Files {
			loadFile := file.File
			content := loadFile.Content
			absFile := overlay.AbsFile(loadFile.AbsPath)
			var funcInfos []*edit.FuncInfo
			switch loadFile.Name {
			case constants.RUNTIME_LINK_FILE:
				if suffixPkg == constants.RUNTIME_INTERNAL_RUNTIME_PKG[n:] {
					foundLink = true
					overrideContent(absFile, instrument_xgo_runtime.GetLinkRuntimeCode())
				}
			case constants.VERSION_FILE:
				if suffixPkg == constants.RUNTIME_CORE_PKG[n:] {
					versionContent := instrument_xgo_runtime.ReplaceVersion(content, xgoVersion, xgoRevision, xgoNumber)
					overrideContent(absFile, versionContent)
				}
			case constants.FLAG_FILE:
				if suffixPkg == constants.RUNTIME_TRAP_FLAGS_PKG[n:] && collectTestTrace {
					flagsContent := instrument_xgo_runtime.InjectFlags(content, collectTestTrace, collectTestTraceDir)
					overrideContent(absFile, flagsContent)
				}
			case constants.TRACE_FILE:
				if suffixPkg == constants.RUNTIME_TRACE_PKG[n:] {
					edit := file.Edit
					funcInfos = instrument_func.InjectRuntimeTrap(edit, loadFile.Syntax, file.Index)
					if edit.HasEdit() {
						overrideContent(absFile, edit.Buffer().String())
					}
				}
			}
			file.TrapFuncs = funcInfos
		}
	}
	if !foundLink {
		return editPackages, ErrLinkFileNotFound
	}
	return editPackages, nil
}

func InstrumentVarTrap(packages *edit.Packages) error {
	instrument_var.Instrument(packages)
	return nil
}

func InstrumentFuncTrap(packages *edit.Packages) error {
	for _, pkg := range packages.Packages {
		for _, file := range pkg.Files {
			instrument_func.InjectRuntimeTrap(file.Edit, file.File.Syntax, file.Index)
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
