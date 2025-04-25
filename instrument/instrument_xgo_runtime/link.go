package instrument_xgo_runtime

import (
	"errors"
	"fmt"
	"go/token"
	"os"

	"github.com/xhd2015/xgo/instrument/config"
	"github.com/xhd2015/xgo/instrument/constants"
	"github.com/xhd2015/xgo/instrument/edit"
	"github.com/xhd2015/xgo/instrument/instrument_func"
	"github.com/xhd2015/xgo/instrument/instrument_runtime/template"
	"github.com/xhd2015/xgo/instrument/load"
	"github.com/xhd2015/xgo/instrument/overlay"
	"github.com/xhd2015/xgo/instrument/patch"
	"github.com/xhd2015/xgo/support/goinfo"
)

var ErrLinkFileNotFound = errors.New("xgo: link file not found")
var ErrLinkFileNotRequired = errors.New("xgo: link file not required")
var ErrRuntimeVersionDeprecatedV1_0_0 = errors.New("runtime version deprecated")

type LinkOptions struct {
	Fset                *token.FileSet
	Mod                 string
	Modfile             string
	XgoVersion          string
	XgoRevision         string
	XgoNumber           int
	CollectTestTrace    bool
	CollectTestTraceDir string

	ReadRuntimeGenFile func(path []string) ([]byte, error)
}

func LinkXgoRuntime(goroot string, projectDir string, xgoRuntimeModuleDir string, goVersion *goinfo.GoVersion, overlayFS overlay.Overlay, overrideContent func(absFile overlay.AbsFile, content string), linkOpts LinkOptions) (*edit.Packages, error) {
	fset := linkOpts.Fset
	mod := linkOpts.Mod
	modfile := linkOpts.Modfile
	xgoVersion := linkOpts.XgoVersion
	xgoRevision := linkOpts.XgoRevision
	xgoNumber := linkOpts.XgoNumber
	collectTestTrace := linkOpts.CollectTestTrace
	collectTestTraceDir := linkOpts.CollectTestTraceDir
	readRuntimeGenFile := linkOpts.ReadRuntimeGenFile

	var opts load.LoadOptions

	if xgoRuntimeModuleDir != "" {
		// xgo runtime is replaced in a separate module
		// so we need to load packages from the separate module
		opts = load.LoadOptions{
			Dir:  xgoRuntimeModuleDir,
			Fset: fset,
		}
	} else {
		opts = load.LoadOptions{
			Dir:     projectDir,
			Overlay: overlayFS,
			Mod:     mod,
			ModFile: modfile,
			Fset:    fset,
		}
	}
	packages, err := load.LoadPackages([]string{
		constants.RUNTIME_INTERNAL_RUNTIME_PKG,
		constants.RUNTIME_CORE_PKG,
		constants.RUNTIME_TRAP_FLAGS_PKG,
		constants.RUNTIME_FUNCTAB_PKG,
		constants.RUNTIME_LEGACY_CORE_INFO_PKG,
		constants.RUNTIME_MOCK_PKG,
		constants.RUNTIME_TRACE_PKG,
		constants.RUNTIME_TRAP_PKG,
	}, opts)
	if err != nil {
		// TODO: handle the case where error indicates the package is not found
		return nil, err
	}
	editPackages := edit.New(packages)
	var foundLink bool
	var runtimeLinkPkg *edit.Package
	var foundMock bool
	var foundTrace bool
	var foundTrap bool
	var foundFunctabPkg bool
	var foundLegacyCoreInfoPkg bool
	var traceFile *edit.File
	var funcTabPkg *edit.Package
	var runtimeCoreVersion string

	// find version first
	for _, pkg := range editPackages.Packages {
		goPkg := pkg.LoadPackage.GoPackage
		if goPkg.Incomplete {
			continue
		}
		if goPkg.ImportPath != constants.RUNTIME_CORE_PKG {
			continue
		}
		var versionFile *edit.File
		for _, file := range pkg.Files {
			if file.File.Name == constants.VERSION_FILE {
				versionFile = file
				break
			}
		}
		if versionFile == nil {
			break
		}
		content := versionFile.File.Content
		absFile := overlay.AbsFile(versionFile.File.AbsPath)
		var err error
		runtimeCoreVersion, err = ParseCoreVersion(content)
		if err != nil {
			return nil, err
		}
		if isDeprecatedCoreVersion(runtimeCoreVersion) {
			return nil, fmt.Errorf("%w: %s", ErrRuntimeVersionDeprecatedV1_0_0, runtimeCoreVersion)
		}
		versionContent := ReplaceActualXgoVersion(content, xgoVersion, xgoRevision, xgoNumber)
		versionContent = checkBypassVersionCheck(versionContent, runtimeCoreVersion)
		overrideContent(absFile, versionContent)
		break
	}

	for _, pkg := range editPackages.Packages {
		goPkg := pkg.LoadPackage.GoPackage
		if goPkg.Incomplete {
			continue
		}
		importPath := goPkg.ImportPath
		suffixPkg, ok := goinfo.PkgWithinModule(importPath, constants.RUNTIME_MODULE)
		if !ok {
			continue
		}
		n := len(constants.RUNTIME_MODULE) + 1
		switch suffixPkg {
		case constants.RUNTIME_MOCK_PKG[n:]:
			foundMock = true
		case constants.RUNTIME_TRACE_PKG[n:]:
			foundTrace = true
		case constants.RUNTIME_TRAP_PKG[n:]:
			foundTrap = true
		case constants.RUNTIME_FUNCTAB_PKG[n:]:
			foundFunctabPkg = true
			funcTabPkg = pkg
		case constants.RUNTIME_LEGACY_CORE_INFO_PKG[n:]:
			foundLegacyCoreInfoPkg = true
		}
		if suffixPkg == constants.RUNTIME_FUNCTAB_PKG[n:] ||
			suffixPkg == constants.RUNTIME_MOCK_PKG[n:] ||
			suffixPkg == constants.RUNTIME_TRAP_PKG[n:] ||
			suffixPkg == constants.RUNTIME_LEGACY_CORE_INFO_PKG[n:] {
			// only for lookup
			continue
		}
		if suffixPkg == constants.RUNTIME_INTERNAL_RUNTIME_PKG[n:] {
			runtimeLinkPkg = pkg
			var runtimeLinkFile *edit.File
			for _, file := range pkg.Files {
				switch file.File.Name {
				case constants.RUNTIME_LINK_FILE:
					foundLink = true
					runtimeLinkFile = file
				}
				if foundLink {
					break
				}
			}

			err := linkRuntimeTemplates(goroot, overlayFS, pkg.LoadPackage.GoPackage.Dir, goVersion, runtimeCoreVersion, runtimeLinkFile, readRuntimeGenFile, overrideContent)
			if err != nil {
				return nil, err
			}
			continue
		}
		for _, file := range pkg.Files {
			loadFile := file.File
			content := loadFile.Content
			absFile := overlay.AbsFile(loadFile.AbsPath)
			switch loadFile.Name {
			case constants.FLAG_FILE:
				if suffixPkg == constants.RUNTIME_TRAP_FLAGS_PKG[n:] && collectTestTrace {
					flagsContent := InjectFlags(content, collectTestTrace, collectTestTraceDir)
					overrideContent(absFile, flagsContent)
				}
			case constants.TRACE_FILE:
				if suffixPkg == constants.RUNTIME_TRACE_PKG[n:] {
					traceFile = file
				}
			}
		}
	}
	// found any usage of xgo public API, but does not found
	// link file, it means the runtime is not instrumented
	if !foundLink {
		if foundMock || foundTrace || foundTrap {
			// mock, trace and trap are public APIs of xgo/runtime
			return editPackages, ErrLinkFileNotFound
		}
		return editPackages, ErrLinkFileNotRequired
	}
	if runtimeCoreVersion == "1.1.0" && runtimeLinkPkg != nil {
		// remove the buggy var ptr trap behavior in runtime v1.1.0
		dir := runtimeLinkPkg.LoadPackage.GoPackage.Dir
		err := patchLegacy(dir, overrideContent)
		if err != nil {
			return nil, err
		}
	}
	if foundFunctabPkg && traceFile != nil {
		// trap trace.go
		edit := traceFile.Edit
		funcInfos, extraFuncs := instrument_func.TrapFuncs(edit, constants.RUNTIME_TRACE_PKG, traceFile.File.Syntax, traceFile.Index, instrument_func.Options{
			// trap all funcs inside trace.go,
			// in reality there is only one func: Trace
			InstrumentMode: config.InstrumentMode_All,
			// force in place edit, which uses overlay
			ForceInPlace: true,
		})
		if len(extraFuncs) > 0 {
			panic(fmt.Errorf("instrument %s.%s: unexpected extra compiler-assisted func: %d", constants.RUNTIME_TRACE_PKG, "Trace", len(extraFuncs)))
		}
		if edit.HasEdit() {
			overrideContent(overlay.AbsFile(traceFile.File.AbsPath), edit.Buffer().String())
		}
		traceFile.TrapFuncs = funcInfos
	}
	if foundLegacyCoreInfoPkg && foundFunctabPkg {
		addLegacyFunctabInit(funcTabPkg, overrideContent)
	}

	return editPackages, nil
}

// linkRuntimeTemplates links the runtime templates:
//
//	runtime_link_template.go -> runtime_link.go
//	xgo_trap_template.go -> GOROOT/src/runtime/xgo_trap.go
//
// why these two files cannot be found by list?
// because they have special build tag:
//
//	'//go:build ignore'
func linkRuntimeTemplates(goroot string, overlayFS overlay.Overlay, internalRuntimeDir string, goVersion *goinfo.GoVersion, coreVersion string, runtimeLinkFile *edit.File, readRuntimeGenFile func(path []string) ([]byte, error), overrideContent func(absFile overlay.AbsFile, content string)) error {
	// fmt.Fprintf(os.Stderr, "DEBUG linkRuntimeTemplates GOROOT: %s\n", goroot)
	// runtimeDir := filepath.Join(goroot, "src", "runtime")
	// runtimeNames, readErr := os.ReadDir(runtimeDir)
	// if readErr != nil {
	// 	fmt.Fprintf(os.Stderr, "DEBUG linkRuntimeTemplates GOROOT read error: %s\n", readErr)
	// } else {
	// 	for _, name := range runtimeNames {
	// 		fmt.Fprintf(os.Stderr, "DEBUG linkRuntimeTemplates runtime file: %s\n", name.Name())
	// 	}
	// }

	if runtimeLinkFile != nil {
		var runtimeLinkTemplateContent string
		if coreVersion != "1.1.0" {
			var err error
			runtimeLinkTemplateContent, err = readRuntimeFileFromDirOrGen(internalRuntimeDir, constants.RUNTIME_LINK_TEMPLATE_PATH, overlayFS, readRuntimeGenFile)
			if err != nil {
				return err
			}
		} else {
			runtimeLinkTemplateContent = legacyRuntimeLinkTemplate
		}
		runtimeLinkTemplateContent, err := patch.RemoveBuildIgnore(runtimeLinkTemplateContent)
		if err != nil {
			return err
		}
		// fmt.Fprintf(os.Stderr, "DEBUG linkRuntimeTemplates override: %s\ncontent: %s\n", runtimeLinkFile.File.AbsPath, runtimeLinkTemplateContent)
		// override: runtime_link_template.go -> runtime_link.go
		overrideContent(overlay.AbsFile(runtimeLinkFile.File.AbsPath), runtimeLinkTemplateContent)
	}

	var xgoTrapTemplateContent string
	if coreVersion != "1.1.1" {
		var err error
		xgoTrapTemplateContent, err = readRuntimeFileFromDirOrGen(internalRuntimeDir, constants.RUNTIME_XGO_TRAP_TEMPLATE_PATH, overlayFS, readRuntimeGenFile)
		if err != nil {
			return err
		}
	} else {
		// 1.1.1 replace with gen because we changed
		// signature of XgoRegister from func(fn *XgoFuncInfo) to func(fn interface{})
		content, err := readRuntimeGenFile(constants.RUNTIME_XGO_TRAP_TEMPLATE_PATH)
		if err != nil {
			return err
		}
		xgoTrapTemplateContent = string(content)
	}

	// override file under GOROOT
	// by always use the one bind with runtime ensures the trap file's
	// API always matches the one that's expected by xgo/runtime
	gorootTrapFile := overlay.AbsFile(constants.GetGoRuntimeXgoTrapFile(goroot))
	content, err := template.InstantiateXgoTrap(xgoTrapTemplateContent, goVersion)
	if err != nil {
		return err
	}

	// override: xgo_trap_template.go -> GOROOT/src/runtime/xgo_trap.go
	overlayFS.OverrideContent(gorootTrapFile, content)
	return nil
}

func readRuntimeFileFromDirOrGen(internalRuntimeDir string, path []string, overlayFS overlay.Overlay, readRuntimeGenFile func(path []string) ([]byte, error)) (string, error) {
	templateFile := hasFile(internalRuntimeDir, path[len(path)-1])
	_, templateContent, readErr := overlayFS.Read(overlay.AbsFile(templateFile))
	if readErr == nil {
		// fmt.Fprintf(os.Stderr, "DEBUG readRuntimeFileFromDirOrGen read from fs: %s\n", templateFile)
		return templateContent, nil
	}
	if !os.IsNotExist(readErr) {
		return "", readErr
	}

	// fallback to the one that xgo embeds(this only happens in when transition from v1.1.0 to v1.1.1)
	templateContentBytes, err := readRuntimeGenFile(path)
	if err != nil {
		return "", err
	}
	// fmt.Fprintf(os.Stderr, "DEBUG readRuntimeFileFromDirOrGen read from embedded: %s\n", path)
	return string(templateContentBytes), nil
}
