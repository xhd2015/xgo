package instrument

import (
	"fmt"
	"os"

	"github.com/xhd2015/xgo/support/edit/goedit"
	"github.com/xhd2015/xgo/support/goinfo"
	"github.com/xhd2015/xgo/support/instrument/edit"
	"github.com/xhd2015/xgo/support/instrument/instrument_func"
	"github.com/xhd2015/xgo/support/instrument/instrument_go"
	"github.com/xhd2015/xgo/support/instrument/instrument_runtime"
	"github.com/xhd2015/xgo/support/instrument/instrument_var"
	"github.com/xhd2015/xgo/support/instrument/instrument_xgo_runtime"
	"github.com/xhd2015/xgo/support/instrument/load"
	"github.com/xhd2015/xgo/support/instrument/overlay"
	"github.com/xhd2015/xgo/support/strutil"
)

var ErrLinkFileNotFound = fmt.Errorf("xgo: link file not found")

// create an overlay: abs file -> content
type Overlay map[string]string

func LinkXgoRuntime(projectDir string, xgoRuntimeModuleDir string, overlayFS overlay.Overlay, mod string, modfile string, xgoVersion string, xgoRevision string, xgoNumber int, collectTestTrace bool, collectTestTraceDir string) error {
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
		instrument_xgo_runtime.RUNTIME_INTERNAL_RUNTIME_PKG,
		instrument_xgo_runtime.RUNTIME_CORE_PKG,
		instrument_xgo_runtime.RUNTIME_TRAP_FLAGS_PKG,
		instrument_xgo_runtime.RUNTIME_TRACE_SIGNAL_PKG,
	}, opts)
	if err != nil {
		// TODO: handle the case where error indicates the package is not found
		return err
	}
	overrideContent := func(absFile overlay.AbsFile, content string) {
		if xgoRuntimeModuleDir == "" {
			overlayFS.OverrideContent(absFile, content)
			return
		}
		// we can directly replace the go files in the xgoRuntimeModuleDir
		// just write back the content to the file
		err := os.WriteFile(string(absFile), strutil.ToReadonlyBytes(content), 0644)
		if err != nil {
			panic(err)
		}
	}
	var foundLink bool
	for _, pkg := range packages.Packages {
		importPath := pkg.GoPackage.ImportPath
		var addTrap bool
		if importPath == instrument_xgo_runtime.RUNTIME_TRACE_SIGNAL_PKG {
			addTrap = true
		}
		for _, file := range pkg.Files {
			content := file.Content
			absFile := overlay.AbsFile(file.AbsPath)
			if addTrap {
				edit := goedit.New(packages.Fset, content)
				instrument_func.EditInjectRuntimeTrap(edit, file.Syntax)
				if edit.HasEdit() {
					overrideContent(absFile, edit.Buffer().String())
				}
				continue
			}
			switch file.Name {
			case instrument_xgo_runtime.RUNTIME_LINK_FILE:
				if importPath == instrument_xgo_runtime.RUNTIME_INTERNAL_RUNTIME_PKG {
					foundLink = true
					overrideContent(absFile, instrument_xgo_runtime.GetLinkRuntimeCode())
				}
			case instrument_xgo_runtime.VERSION_FILE:
				if importPath == instrument_xgo_runtime.RUNTIME_CORE_PKG {
					versionContent := instrument_xgo_runtime.ReplaceVersion(content, xgoVersion, xgoRevision, xgoNumber)
					overrideContent(absFile, versionContent)
				}
			case instrument_xgo_runtime.FLAG_FILE:
				if importPath == instrument_xgo_runtime.RUNTIME_TRAP_FLAGS_PKG && collectTestTrace {
					flagsContent := instrument_xgo_runtime.InjectFlags(content, collectTestTrace, collectTestTraceDir)
					overrideContent(absFile, flagsContent)
				}
			}
		}
	}
	if !foundLink {
		return ErrLinkFileNotFound
	}
	return nil
}

func InstrumentVarTrap(packages *edit.Packages) error {
	instrument_var.Instrument(packages)
	return nil
}

func InstrumentFuncTrap(packages *edit.Packages) error {
	for _, pkg := range packages.Packages {
		for _, file := range pkg.Files {
			instrument_func.EditInjectRuntimeTrap(file.Edit, file.File.Syntax)
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
