package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/xhd2015/xgo/instrument/compiler_extra"
	"github.com/xhd2015/xgo/instrument/config"
	"github.com/xhd2015/xgo/instrument/constants"
	"github.com/xhd2015/xgo/instrument/edit"
	"github.com/xhd2015/xgo/instrument/instrument_func"
	"github.com/xhd2015/xgo/instrument/instrument_go"
	"github.com/xhd2015/xgo/instrument/instrument_intf"
	"github.com/xhd2015/xgo/instrument/instrument_reg"
	"github.com/xhd2015/xgo/instrument/instrument_var"
	"github.com/xhd2015/xgo/instrument/instrument_xgo_runtime"
	"github.com/xhd2015/xgo/instrument/load"
	"github.com/xhd2015/xgo/instrument/overlay"
	"github.com/xhd2015/xgo/instrument/resolve"
	"github.com/xhd2015/xgo/support/cmd"
	"github.com/xhd2015/xgo/support/fileutil"
	"github.com/xhd2015/xgo/support/goinfo"
	"github.com/xhd2015/xgo/support/strutil"
)

// limit the instrument file size up to 1MB
// larger file may bloat in size, causing
// the go compiler to fail.
// see https://github.com/xhd2015/xgo/issues/303
// TODO: we may add unit test for this
// and check if later go version has fixed this
// known for: go1.19
const MAX_FILE_SIZE = 1 * 1024 * 1024

type instrumentResult struct {
	compilerExtra *compiler_extra.Packages
	// traceImportsAdded: import _ "…/runtime/trace" was added to the
	// packages of the command
	traceImportsAdded bool
}

// goroot is critical for stdlib
// includeAsMainModules: extra module paths treated as main for mock/trap (option B:
// reclassify packages already on the load graph; do not bulk-load module/...).
func instrumentUserCode(goroot string, projectDir string, projectRoot string, goVersion *goinfo.GoVersion, xgoSrc string, mod string, modfile string, mainModule string, includeAsMainModules []string, xgoRuntimeModuleDir string, mayHaveCover bool, overlayFS overlay.Overlay, overlayFile string, includeTest bool, rules []Rule, trapPkgs []string, trapAll string, collectTestTrace bool, collectTestTraceDir string, xgoRaceSafe bool, goFlag bool, triedUpgrade bool, trimpath bool, pkgPatterns []string, addTraceImports bool) (*instrumentResult, error) {
	logDebug("instrumentUserSpace: mod=%s, modfile=%s, xgoRuntimeModuleDir=%s, includeTest=%v, collectTestTrace=%v, includeAsMainModules=%v, overlayFile=%s", mod, modfile, xgoRuntimeModuleDir, includeTest, collectTestTrace, includeAsMainModules, overlayFile)
	if mod == "" {
		// check vendor dir
		vendorDir, err := getVendorDir(projectRoot)
		if err != nil {
			return nil, err
		}
		if vendorDir != "" {
			mod = "vendor"
		}
	}

	overrideXgoContent := func(absFile overlay.AbsFile, content string) {
		// a file instrumented from an in-memory content (the pristine
		// trace.go, see importResult.pristineTrace) stays in the overlay:
		// other runs in the project may be compiling the file on disk
		if xgoRuntimeModuleDir == "" || overlayFS.Get(absFile) != nil {
			overlayFS.OverrideContent(absFile, content)
			return
		}
		// we can directly replace the other go files in the
		// xgoRuntimeModuleDir: their content is the same for the same
		// runtime, write it only when it differs, and atomically
		err := overlay.WriteFileIfChanged(string(absFile), strutil.ToReadonlyBytes(content))
		if err != nil {
			panic(err)
		}
	}
	includeMain, loadPkgs, err := getLoadPackages(rules)
	if err != nil {
		return nil, err
	}
	logDebug("loadPkgs: includeMain=%v loadPkgs=%v", includeMain, loadPkgs)
	var loadArgs []string
	// targeted: list only what the command compiles, i.e. the packages
	// of its patterns, their dependencies and those of their tests
	// (`go list -deps -test`), rather than the whole main module
	targeted := includeMain && len(pkgPatterns) > 0
	// packages whose tests the command builds; nil means unknown
	var targetPkgs map[string]bool
	if targeted {
		loadArgs = append(loadArgs, pkgPatterns...)
	} else if includeMain {
		loadArgs = append(loadArgs, mainModule+"/...")
		loadArgs = append(loadArgs, "main")
	}
	loadArgs = append(loadArgs, loadPkgs...)
	loadArgs = append(loadArgs, trapPkgs...)

	fset := token.NewFileSet()
	pkgs := &edit.Packages{
		Fset: fset,
		LoadOptions: load.LoadOptions{
			Dir:             projectDir,
			Mod:             mod,
			Overlay:         overlayFS,
			OverlayFile:     overlayFile,
			IncludeTest:     includeTest,
			ModFile:         modfile,
			MaxFileSize:     MAX_FILE_SIZE,
			FilterErrorFile: true,
			Goroot:          goroot,
			GoVersion:       goVersion,
			Fset:            fset,
			// always load deps, should be fast enough
			Deps: true,
			// Most dependencies are not instrumented unless a mock refers
			// to them, and traverse loads those on demand: only parse what
			// is instrumented regardless.
			Test: targeted,
			OnList: func(list []*goinfo.Package) {
				if targeted {
					targetPkgs = matchedPackages(list, pkgPatterns)
				}
			},
			ParseFilter: func(p *goinfo.Package) bool {
				if !p.DepOnly || pkgWithinAnyModule(p.ImportPath, mainModule, includeAsMainModules) {
					return true
				}
				return p.Standard && config.GetPkgConfig(p.ImportPath) != nil
			},
			// test files of other packages are not compiled
			TestFilter: func(p *goinfo.Package) bool {
				if targetPkgs != nil {
					return targetPkgs[p.ImportPath]
				}
				return !p.DepOnly
			},
		},
	}
	// List the packages to instrument while the xgo runtime is linked
	// below: both mostly wait for a go list. If the runtime turns out not
	// to need instrumenting, the listing is wasted, which is rare.
	type listResult struct {
		pkgs []*goinfo.Package
		err  error
	}
	listDone := make(chan listResult, 1)
	listOpts := load.ListOptions(pkgs.LoadOptions)
	go func() {
		list, err := goinfo.ListPackages(loadArgs, listOpts)
		listDone <- listResult{list, err}
	}()
	pkgs.LoadOptions.Listed = func() ([]*goinfo.Package, error) {
		r := <-listDone
		return r.pkgs, r.err
	}

	xgoPkgs, err := instrument_xgo_runtime.LinkXgoRuntime(goroot, projectDir, xgoRuntimeModuleDir, goVersion, overlayFS, overrideXgoContent, instrument_xgo_runtime.LinkOptions{
		Fset:                fset,
		Mod:                 mod,
		Modfile:             modfile,
		XgoVersion:          VERSION,
		XgoRevision:         REVISION,
		XgoNumber:           NUMBER,
		CollectTestTrace:    collectTestTrace,
		CollectTestTraceDir: collectTestTraceDir,
		XgoRaceSafe:         xgoRaceSafe,
		OverlayFile:         overlayFile,
		ReadRuntimeGenFile: func(path []string) ([]byte, error) {
			return readRuntimeGenFile(xgoSrc, path)
		},
	})
	if err != nil {
		if err == instrument_xgo_runtime.ErrLinkFileNotRequired {
			return nil, nil
		}
		if err == instrument_xgo_runtime.ErrLinkFileNotFound {
			if !goFlag {
				fmt.Fprintf(os.Stderr, `WARNING: xgo: skip runtime instrumentation, upgrade:
  go get %s@latest
  import _ %q
	`, constants.RUNTIME_INTERNAL_TRAP_PKG, constants.RUNTIME_INTERNAL_TRAP_PKG)
			}
			return nil, nil
		}
		if errors.Is(err, instrument_xgo_runtime.ErrRuntimeVersionDeprecatedV1_0_0) {
			if !goFlag {
				if triedUpgrade {
					return nil, fmt.Errorf("xgo v%s auto upgrade failed, you can fix this by:\n  go get %s@latest", VERSION, constants.RUNTIME_MODULE)
				}
				return nil, fmt.Errorf("xgo v%s cannot work with deprecated xgo/runtime: %w, upgrade with:\n  go get %s@latest", VERSION, err, constants.RUNTIME_MODULE)
			}
			return nil, nil
		}
		return nil, err
	}
	funcTabPkg := xgoPkgs.PackageByPath[constants.RUNTIME_FUNCTAB_PKG]
	if funcTabPkg == nil || !hasFunc(funcTabPkg, constants.RUNTIME_FUNCTAB_REGISTER) {
		logDebug("skip functab registering")
		return nil, nil
	}

	logDebug("start load: %v", loadArgs)
	loadPackages, err := load.LoadPackages(loadArgs, pkgs.LoadOptions)
	if err != nil {
		return nil, err
	}
	pkgs.Add(loadPackages)
	// remove deps flag
	pkgs.LoadOptions.Deps = false
	// later loads list single packages: keep the targets of the initial load
	pkgs.LoadOptions.Test = false
	pkgs.LoadOptions.OnList = nil
	pkgs.LoadOptions.Listed = nil
	if targeted {
		logDebug("instrument: %d target packages", len(targetPkgs))
	}

	// merge with xgo pkgs with override
	// because xgoPkgs may have instrumented files
	// like trace.go
	pkgs.Merge(xgoPkgs, true)

	var initCnt int
	var xgoCnt int
	var allowCnt int
	var depOnlyCnt int
	var mainCnt int
	for _, pkg := range pkgs.Packages {
		pkgPath := pkg.LoadPackage.GoPackage.ImportPath
		isMain := pkgWithinAnyModule(pkgPath, mainModule, includeAsMainModules)
		if isMain {
			pkg.Main = true
			mainCnt++
		}
		if !pkg.LoadPackage.GoPackage.DepOnly {
			pkg.Initial = true
			initCnt++
		} else {
			depOnlyCnt++
		}
		isXgo, allow := config.CheckInstrument(pkgPath)
		if isXgo {
			pkg.Xgo = true
			xgoCnt++
		}
		if allow {
			pkg.AllowInstrument = true
			allowCnt++
		}
	}
	logDebug("instrument: main pkgs=%d, init pkgs=%d, depOnly pkgs=%d, xgo pkgs=%d, allow pkgs=%d", mainCnt, initCnt, depOnlyCnt, xgoCnt, allowCnt)

	// insert func trap
	// disable instrumenting xgo/runtime, except xgo/runtime/test
	reg := resolve.NewPackagesRegistry(pkgs)
	// trap var for packages in main module
	mainPkgs := pkgs.Filter(func(pkg *edit.Package) bool {
		return pkg.Main && pkg.AllowInstrument
	})
	for _, pkg := range mainPkgs {
		resolve.CollectDecls(pkg)
	}

	// List every import of the files traverse visits that is not listed
	// yet (test-only imports are not in the -deps list) in one go, instead
	// of one `go list` each when traverse reaches them. It also names every
	// import before traverse starts: files are traversed in parallel, and
	// the name of an import must not depend on whether another file listed
	// it first.
	var fileImports []string
	for _, pkg := range mainPkgs {
		for _, file := range pkg.Files {
			for _, imp := range file.File.Syntax.Imports {
				impPath, err := strconv.Unquote(imp.Path.Value)
				if err == nil && impPath != "C" {
					fileImports = append(fileImports, impPath)
				}
			}
		}
	}
	if err := pkgs.ListPackages(fileImports); err != nil {
		// only a prefetch: traverse loads whatever it reaches on its own
		logDebug("list imports: %v", err)
	}

	traverseBegin := time.Now()
	logDebug("traverse: len(mainPkgs)=%d", len(mainPkgs))

	var recorder resolve.Recorder
	err = resolve.Traverse(reg, mainPkgs, &recorder)
	if err != nil {
		return nil, err
	}
	logDebug("traverse: cost=%v", time.Since(traverseBegin))

	// Option B: reclassify packages loaded during traverse (deps pulled by
	// imports/mock-refs) as main when they fall under include-as modules.
	// Initial load often does not yet include those packages.
	if len(includeAsMainModules) > 0 {
		var extraMain int
		for _, pkg := range pkgs.Packages {
			if pkg.Main {
				continue
			}
			pkgPath := pkg.LoadPackage.GoPackage.ImportPath
			if pkgWithinAnyModule(pkgPath, "", includeAsMainModules) {
				pkg.Main = true
				extraMain++
				logDebug("instrument: post-traverse include-as-main pkg=%s", pkgPath)
			}
		}
		if extraMain > 0 {
			logDebug("instrument: post-traverse extra main pkgs=%d", extraMain)
			// Collect decls on newly main packages so var trap / resolve stay consistent.
			newMainPkgs := pkgs.Filter(func(pkg *edit.Package) bool {
				return pkg.Main && pkg.AllowInstrument
			})
			for _, pkg := range newMainPkgs {
				resolve.CollectDecls(pkg)
			}
			mainPkgs = newMainPkgs
		}
	}

	logDebug("start instrumentVarTrap: len(mainPkgs)=%d", len(mainPkgs))
	err = instrument_var.TrapVariables(pkgs, &recorder)
	if err != nil {
		return nil, err
	}

	// ""->default
	needTrapAll := trapAll == "true" || (trapAll == "" && recorder.HasTrapInterceptorRef)
	if needTrapAll {
		// exported functions of every dependency are instrumented
		pkgs.ParseDeferred()
	}

	logDebug("start instrumentFuncTrap: len(packages)=%d, needTrapAll=%v", len(pkgs.Packages), needTrapAll)
	// packages are instrumented in parallel: each one edits only its own
	// files; the recorder is not safe for concurrent use, look it up first
	pkgRecorders := make([]*resolve.PkgRecorder, len(pkgs.Packages))
	for i, pkg := range pkgs.Packages {
		if pkg.AllowInstrument {
			pkgRecorders[i] = recorder.GetOrInit(pkg.LoadPackage.GoPackage.ImportPath)
		}
	}
	pkgExtras := make([]*compiler_extra.Package, len(pkgs.Packages))
	forEachParallel(len(pkgs.Packages), func(i int) {
		pkg := pkgs.Packages[i]
		if !pkg.AllowInstrument {
			return
		}
		pkgPath := pkg.LoadPackage.GoPackage.ImportPath
		cfg := config.GetPkgConfig(pkgPath)
		stdlib := pkg.LoadPackage.GoPackage.Standard
		main := pkg.Main
		initial := pkg.Initial
		if stdlib && cfg != nil {
			// the predefined std packages are listed only to be loaded:
			// keep them on their whitelist instead of trapping every
			// function, which made calls like time.Time.Sub allocate
			initial = false
		}
		pkgRecorder := pkgRecorders[i]

		var hasVarTrap bool
		if pkgRecorder != nil && pkgRecorder.NumVars > 0 {
			hasVarTrap = true
		}

		mode := config.CheckInstrumentMode(stdlib, main, initial, needTrapAll)

		var extraFiles []*compiler_extra.File
		pkgExtraQuota := config.MAX_EXTRA_FUNCS_PER_PKG
		for _, file := range pkg.Files {
			if !pkg.Main && strings.HasSuffix(file.File.Name, "_test.go") {
				// skip test files outside main package
				continue
			}
			funcs, extraFuncs := instrument_func.TrapFuncs(file.Edit, pkgPath, file.File.Syntax, file.Index, instrument_func.Options{
				PkgRecorder:    pkgRecorder,
				PkgConfig:      cfg,
				Main:           main,
				InstrumentMode: mode,
				PkgExtraQuota:  &pkgExtraQuota,
			})
			file.TrapFuncs = append(file.TrapFuncs, funcs...)

			// interface types
			var extraInterfaces []*compiler_extra.Interface
			if mode != config.InstrumentMode_None {
				intfTypes := instrument_intf.CollectInterfaces(file)
				if main {
					file.InterfaceTypes = append(file.InterfaceTypes, intfTypes...)
				} else {
					for _, intf := range intfTypes {
						if mode == config.InstrumentMode_All {
							// pass
						} else if mode == config.InstrumentMode_Exported {
							if !token.IsExported(intf.Name) {
								continue
							}
						} else {
							if config.IS_DEV {
								panic(fmt.Sprintf("unhandled instrument mode: %d", mode))
							}
						}
						extraInterfaces = append(extraInterfaces, &compiler_extra.Interface{
							Name: intf.Name,
						})
					}
				}
			}

			if len(extraFuncs) > 0 || len(extraInterfaces) > 0 {
				extraFiles = append(extraFiles, &compiler_extra.File{
					Name:       file.File.Name,
					AbsFile:    file.File.AbsPath,
					Funcs:      extraFuncs,
					Interfaces: extraInterfaces,
				})
			}
		}
		if len(extraFiles) > 0 {
			md5sum := md5sumTraps(extraFiles, hasVarTrap)
			pkgExtras[i] = &compiler_extra.Package{
				Path:       pkg.LoadPackage.GoPackage.ImportPath,
				Files:      extraFiles,
				HasVarTrap: hasVarTrap,
				TrapMD5Sum: md5sum,
			}
		}
	})
	var extraPkgs []*compiler_extra.Package
	for _, extra := range pkgExtras {
		if extra != nil {
			extraPkgs = append(extraPkgs, extra)
		}
	}
	logDebug("extraPkgs: %d", len(extraPkgs))

	logDebug("generate functab register")
	registerFuncTab(pkgs, trimpath)

	var traceImportsAdded bool
	if addTraceImports && targetPkgs != nil {
		addTraceImportEdits(pkgs, targetPkgs)
		traceImportsAdded = true
	}

	logDebug("collect edits")
	updatedFiles, err := applyInstrumentWithEditNotes(overlayFS, pkgs, mayHaveCover, overrideXgoContent)
	if err != nil {
		return nil, err
	}
	var compilerExtra *compiler_extra.Packages
	if len(extraPkgs) > 0 {
		compilerExtra = &compiler_extra.Packages{
			Packages: extraPkgs,
		}
	}
	logDebug("finish instruments, updated %d files", updatedFiles)
	return &instrumentResult{
		compilerExtra:     compilerExtra,
		traceImportsAdded: traceImportsAdded,
	}, nil
}

// addTraceImportEdits adds import _ "…/runtime/trace" to the packages of
// the command, like addBlankImports does with copies of their files, but
// as an edit of the files already loaded: one Go file, or every test file
// of packages that only have tests. Inserted right after the package
// name, it keeps line numbers.
func addTraceImportEdits(pkgs *edit.Packages, targets map[string]bool) {
	for _, pkg := range pkgs.Packages {
		goPkg := pkg.LoadPackage.GoPackage
		if !targets[goPkg.ImportPath] || goPkg.Standard {
			continue
		}
		var files []string
		if len(goPkg.GoFiles) > 0 {
			if hasString(goPkg.Imports, constants.RUNTIME_TRACE_PKG) {
				continue
			}
			files = goPkg.GoFiles[:1]
		} else {
			files = append(append(files, goPkg.TestGoFiles...), goPkg.XTestGoFiles...)
		}
		for _, file := range pkg.Files {
			if file.File.Syntax == nil || !hasString(files, file.File.Name) {
				continue
			}
			file.Edit.Insert(file.File.Syntax.Name.End(), fmt.Sprintf(";import _ %q", constants.RUNTIME_TRACE_PKG))
		}
	}
}

func md5sumTraps(files []*compiler_extra.File, hasVarTrap bool) string {
	h := md5.New()
	for _, file := range files {
		// assume files are stable
		io.WriteString(h, file.Name)
		for _, fn := range file.Funcs {
			io.WriteString(h, fn.IdentityName)
		}
		for _, intf := range file.Interfaces {
			io.WriteString(h, intf.Name)
		}
	}
	var v string = "v:0"
	if hasVarTrap {
		v = "v:1"
	}
	io.WriteString(h, v)
	md5 := h.Sum(nil)
	return hex.EncodeToString(md5[:])
}

func registerFuncTab(packages *edit.Packages, trimpath bool) {
	fset := packages.Fset
	// each file registers its own functions, in parallel
	forEachParallel(len(packages.Packages), func(i int) {
		pkg := packages.Packages[i]
		pkgPath := pkg.LoadPackage.GoPackage.ImportPath
		stdlib := pkg.LoadPackage.GoPackage.Standard
		for _, file := range pkg.Files {
			instrument_reg.RegisterFuncTab(fset, file, pkgPath, stdlib, trimpath)
		}
	})
}

func applyInstrumentWithEditNotes(overlayFS overlay.Overlay, packages *edit.Packages, mayHaveCover bool, xgoOverrideContent func(absFile overlay.AbsFile, content string)) (int, error) {
	type editedFile struct {
		pkg  *edit.Package
		file *edit.File
	}
	var files []editedFile
	for _, pkg := range packages.Packages {
		for _, file := range pkg.Files {
			if file.Edit.HasEdit() {
				files = append(files, editedFile{pkg: pkg, file: file})
			}
		}
	}
	// the contents are built in parallel, and registered in order
	contents := make([]string, len(files))
	errs := make([]error, len(files))
	forEachParallel(len(files), func(i int) {
		file := files[i].file
		// Bytes returns a new slice, used nowhere else
		content := strutil.ToReadonlyString(file.Edit.Buffer().Bytes())
		if mayHaveCover {
			var err error
			content, err = instrument_go.AddEditsNotes(file.Edit, file.File.AbsPath, strutil.ToReadonlyString(file.File.Content), content)
			if err != nil {
				errs[i] = fmt.Errorf("failed to add edits: %s %w", file.File.AbsPath, err)
				return
			}
		}
		contents[i] = content
	})
	for i, f := range files {
		if errs[i] != nil {
			return 0, errs[i]
		}
		absFile := overlay.AbsFile(f.file.File.AbsPath)
		if f.pkg.Xgo && xgoOverrideContent != nil {
			xgoOverrideContent(absFile, contents[i])
		} else {
			overlayFS.OverrideContent(absFile, contents[i])
		}
	}
	return len(files), nil
}

func hasFunc(pkg *edit.Package, fn string) bool {
	for _, file := range pkg.Files {
		for _, decl := range file.File.Syntax.Decls {
			fnDecl, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			if fnDecl.Name != nil && fnDecl.Name.Name == fn {
				return true
			}
		}
	}
	return false
}

func getLoadPackages(rules []Rule) (includeMain bool, packages []string, err error) {
	var mainExcludueFunc bool
	var mainExcludeVar bool
	for _, rule := range rules {
		if rule.MainModule != nil && *rule.MainModule {
			var kinds []string
			if rule.Kind != nil {
				kinds = splitCommaList(*rule.Kind)
			}
			if rule.Action == "exclude" {
				if len(kinds) == 0 {
					mainExcludueFunc = true
					mainExcludeVar = true
				} else {
					for _, kind := range kinds {
						if kind == "func" {
							mainExcludueFunc = true
						} else if kind == "var" {
							mainExcludeVar = true
						}
					}
				}
			}
			continue
		}
		// must be explicit include for
		// third party packages
		if rule.Action != "include" {
			continue
		}
		var rulePkgs []string
		if rule.Pkg != nil {
			rulePkgs = splitCommaList(*rule.Pkg)
		}
		if len(rulePkgs) > 0 {
			for _, pkg := range rulePkgs {
				n := len(pkg)
				starIdx := strings.Index(pkg, "*")
				if starIdx == -1 {
					packages = append(packages, pkg)
					continue
				}
				if starIdx == n-1 {
					// last one
					return false, nil, fmt.Errorf("only support pkg/**: %s", pkg)
				}
				if starIdx < n-2 {
					return false, nil, fmt.Errorf("* in the middle not supported, only support pkg/**: %s", pkg)
				}
				if starIdx != n-2 {
					panic("unexpected *")
				}
				if pkg[n-1] != '*' {
					return false, nil, fmt.Errorf("only support pkg/**: %s", pkg)
				}
				if n-3 < 0 || pkg[n-3] != '/' {
					return false, nil, fmt.Errorf("only support pkg/**: %s", pkg)
				}
				// replace ** with ...
				pkg = pkg[:n-2] + "..."
				packages = append(packages, pkg)
			}
		}
	}
	return !(mainExcludueFunc && mainExcludeVar), packages, nil
}

func splitCommaList(s string) []string {
	if s == "" {
		return nil
	}
	list := strings.Split(s, ",")
	trimmedList := make([]string, 0, len(list))
	for _, item := range list {
		item := strings.TrimSpace(item)
		if item == "" {
			continue
		}
		trimmedList = append(trimmedList, item)
	}
	return trimmedList
}

// pkgWithinAnyModule reports whether pkgPath is under realMain or any extra
// module path (mock-rule include-as-main-module, additive).
func pkgWithinAnyModule(pkgPath string, realMain string, extras []string) bool {
	if realMain != "" {
		if _, ok := goinfo.PkgWithinModule(pkgPath, realMain); ok {
			return true
		}
	}
	for _, m := range extras {
		if m == "" || m == realMain {
			continue
		}
		if _, ok := goinfo.PkgWithinModule(pkgPath, m); ok {
			return true
		}
	}
	return false
}

func setupLocalXgoGenDir(xgoGenDir string) error {
	// to avoid `go mod vendor` detecting this gen dir
	xgoGenGoMod := filepath.Join(xgoGenDir, "go.mod")
	stat, statErr := os.Stat(xgoGenGoMod)
	if statErr != nil {
		if !os.IsNotExist(statErr) {
			return statErr
		}
	} else {
		if !stat.IsDir() {
			return nil
		}
		err := os.RemoveAll(xgoGenGoMod)
		if err != nil {
			return err
		}
	}

	err := os.MkdirAll(xgoGenDir, 0755)
	if err != nil {
		return err
	}

	err = os.WriteFile(xgoGenGoMod, []byte(strings.Join([]string{
		"// a dummy go.mod generated by xgo to avoid affecting `go mod tidy` and `go mod vendor`",
		"module github.com/xhd2015/xgo/gen",
		"",
		"go 1.18",
		"",
	}, "\n")), 0644)
	if err != nil {
		return err
	}
	return nil
}

func getLocalXgoGenDir(projectRootDir string) (string, error) {
	xgoDir, err := getLocalXgoDir(projectRootDir)
	if err != nil {
		return "", err
	}
	xgoGenDir := filepath.Join(xgoDir, "gen")
	err = setupLocalXgoGenDir(xgoGenDir)
	if err != nil {
		return "", err
	}
	return xgoGenDir, nil
}

func getLocalXgoDir(projectDir string) (string, error) {
	xgoDir := filepath.Join(projectDir, ".xgo")
	stat, statErr := os.Stat(xgoDir)
	if statErr != nil {
		if !os.IsNotExist(statErr) {
			return "", statErr
		}
	} else {
		if !stat.IsDir() {
			return "", fmt.Errorf("%s is not a directory, check or delete it so xgo can generate it as directory", humanReadablePath(xgoDir))
		}
		return xgoDir, nil
	}
	err := os.MkdirAll(xgoDir, 0755)
	if err != nil {
		return "", err
	}

	// show toplevel dir without warning .git is not found
	topLevel, _ := cmd.Dir(projectDir).Stderr(io.Discard).Output("git", "rev-parse", "--show-toplevel")
	if topLevel != "" {
		// add .xgo/gen to .gitignore
		err := fileutil.UpdateFile(filepath.Join(topLevel, ".gitignore"), func(content []byte) (bool, []byte, error) {
			newContent, updated := gitignoreAdd(content, XGO_GEN_IGNORE_PATTERN)
			if !updated {
				return false, content, nil
			}
			return true, newContent, nil
		})
		if err != nil {
			return "", err
		}
	}

	return xgoDir, nil
}

const XGO_GEN_IGNORE_PATTERN = "**/.xgo/gen"

func gitignoreAdd(content []byte, pattern string) ([]byte, bool) {
	idx := bytes.Index(content, []byte(pattern))
	if idx >= 0 && isDirectiveEnd(content, idx+len(pattern)) {
		return content, false
	}
	var prefix string
	if len(content) > 0 && !bytes.HasSuffix(content, []byte("\n")) {
		prefix = "\n"
	}
	return append(content, []byte(prefix+pattern+"\n")...), true
}

func isDirectiveEnd(content []byte, i int) bool {
	n := len(content)
	for ; i < n; i++ {
		if content[i] == ' ' || content[i] == '\t' {
			continue
		}
		break
	}
	if i >= n {
		return true
	}
	if bytes.HasSuffix(content[i:], []byte("#")) {
		return true
	}
	if bytes.HasSuffix(content[i:], []byte("\n")) {
		return true
	}
	if bytes.HasSuffix(content[i:], []byte("\r\n")) {
		return true
	}
	return false
}

func humanReadablePath(path string) string {
	wd, err := os.Getwd()
	if err != nil {
		return path
	}
	rel, err := filepath.Rel(wd, path)
	if err != nil {
		return path
	}
	if strings.Count(rel, ".."+string(filepath.Separator)) > 2 {
		return path
	}
	return rel
}

// writeCompilerExtra writes the compiler extra mapping and its sums into
// content directories under overlayDir, so that concurrent runs needing
// different ones do not overwrite each other's, and returns their paths.
func writeCompilerExtra(compilerExtra *compiler_extra.Packages, overlayDir string) (string, string, error) {
	mapping := compilerExtra.BuildMapping()
	extraData, err := json.Marshal(mapping)
	if err != nil {
		return "", "", err
	}
	compilerExtraFile, err := overlay.WriteContentFile(overlayDir, "compiler_extra.json", extraData)
	if err != nil {
		return "", "", fmt.Errorf("write compiler extra file: %w", err)
	}

	// pkg->sum
	md5sumMapping := compilerExtra.BuildMD5SumMapping()
	sumData, err := json.Marshal(md5sumMapping)
	if err != nil {
		return "", "", fmt.Errorf("write compiler extra sum file: %w", err)
	}
	compilerExtraSumFile, err := overlay.WriteContentFile(overlayDir, "compiler_extra.sum.json", sumData)
	if err != nil {
		return "", "", fmt.Errorf("write compiler extra sum file: %w", err)
	}
	return compilerExtraFile, compilerExtraSumFile, nil
}

// matchedPackages returns the packages that patterns select, according to
// go list's Match.
func matchedPackages(list []*goinfo.Package, patterns []string) map[string]bool {
	targets := make(map[string]bool)
	for _, pattern := range patterns {
		pattern = cleanPattern(pattern)
		var matched bool
		for _, p := range list {
			for _, m := range p.Match {
				if m == pattern {
					targets[p.ImportPath] = true
					matched = true
				}
			}
		}
		if !matched {
			// written differently than go list reports it (or matching
			// nothing): instrument the whole main module rather than
			// miss the packages it names
			return nil
		}
	}
	return targets
}

// cleanPattern puts a package pattern in the canonical form go list
// reports in Match, like cmd/go does: "./a/" and "./a//b" become "./a"
// and "./a/b".
func cleanPattern(pattern string) string {
	if strings.Contains(pattern, "@") || strings.Contains(pattern, ":") {
		// pkg@version, or windows paths: left to the fallback
		return pattern
	}
	pattern = filepath.ToSlash(pattern)
	if strings.HasPrefix(pattern, "./") {
		pattern = "./" + path.Clean(pattern)
		if pattern == "./." {
			pattern = "."
		}
		return pattern
	}
	return path.Clean(pattern)
}

func hasString(list []string, s string) bool {
	for _, e := range list {
		if e == s {
			return true
		}
	}
	return false
}
