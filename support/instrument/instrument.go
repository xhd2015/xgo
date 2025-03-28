package instrument

import (
	"fmt"
	"path/filepath"

	"github.com/xhd2015/xgo/support/goinfo"
	"github.com/xhd2015/xgo/support/instrument/edit"
	"github.com/xhd2015/xgo/support/instrument/instrument_func"
	"github.com/xhd2015/xgo/support/instrument/instrument_var"
	"github.com/xhd2015/xgo/support/instrument/instrument_xgo_runtime"
	"github.com/xhd2015/xgo/support/instrument/overlay"
	"github.com/xhd2015/xgo/support/instrument/runtime"
)

var ErrLinkFileNotFound = fmt.Errorf("xgo: link file not found")

// create an overlay: abs file -> content
type Overlay map[string]string

func LinkXgoRuntime(projectDir string, overlayFS overlay.Overlay, modfile string, xgoVersion string, xgoRevision string, xgoNumber int, collectTestTrace bool, collectTestTraceDir string) error {
	packages, err := goinfo.ListPackages([]string{
		instrument_xgo_runtime.RUNTIME_INTERNAL_RUNTIME_PKG,
		instrument_xgo_runtime.RUNTIME_CORE_PKG,
		instrument_xgo_runtime.RUNTIME_TRAP_FLAGS_PKG,
	}, goinfo.LoadPackageOptions{
		Dir:     projectDir,
		ModFile: modfile,
	})
	if err != nil {
		// TODO: handle the case where error indicates the package is not found
		return err
	}
	var foundLink bool
	for _, pkg := range packages {
		for _, file := range pkg.GoFiles {
			switch file {
			case instrument_xgo_runtime.RUNTIME_LINK_FILE:
				if pkg.ImportPath == instrument_xgo_runtime.RUNTIME_INTERNAL_RUNTIME_PKG {
					foundLink = true
					absFile := overlay.AbsFile(filepath.Join(pkg.Dir, file))
					overlayFS.OverrideContent(absFile, instrument_xgo_runtime.GetLinkRuntimeCode())
				}
			case instrument_xgo_runtime.VERSION_FILE:
				if pkg.ImportPath == instrument_xgo_runtime.RUNTIME_CORE_PKG {
					absFile := overlay.AbsFile(filepath.Join(pkg.Dir, file))
					_, content, err := overlayFS.Read(absFile)
					if err != nil {
						return err
					}
					versionContent := instrument_xgo_runtime.ReplaceVersion(content, xgoVersion, xgoRevision, xgoNumber)
					overlayFS.OverrideContent(absFile, versionContent)
				}
			case instrument_xgo_runtime.FLAG_FILE:
				if pkg.ImportPath == instrument_xgo_runtime.RUNTIME_TRAP_FLAGS_PKG && collectTestTrace {
					absFile := overlay.AbsFile(filepath.Join(pkg.Dir, file))
					_, content, err := overlayFS.Read(absFile)
					if err != nil {
						return err
					}
					flagsContent := instrument_xgo_runtime.InjectFlags(content, collectTestTrace, collectTestTraceDir)
					overlayFS.OverrideContent(absFile, flagsContent)
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

func InstrumentRuntime(goroot string, opts runtime.InstrumentRuntimeOptions) error {
	return runtime.InstrumentRuntime(goroot, opts)
}
