package main

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

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

// goroot is critical for stdlib
func instrumentUserCode(goroot string, projectDir string, projectRoot string, goVersion *goinfo.GoVersion, xgoSrc string, mod string, modfile string, mainModule string, xgoRuntimeModuleDir string, mayHaveCover bool, overlayFS overlay.Overlay, includeTest bool, rules []Rule, trapPkgs []string, trapAll string, collectTestTrace bool, collectTestTraceDir string, goFlag bool, triedUpgrade bool) error {
	logDebug("instrumentUserSpace: mod=%s, modfile=%s, xgoRuntimeModuleDir=%s, includeTest=%v, collectTestTrace=%v", mod, modfile, xgoRuntimeModuleDir, includeTest, collectTestTrace)
	if mod == "" {
		// check vendor dir
		vendorDir, err := getVendorDir(projectRoot)
		if err != nil {
			return err
		}
		if vendorDir != "" {
			mod = "vendor"
		}
	}

	overrideXgoContent := func(absFile overlay.AbsFile, content string) {
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
	fset := token.NewFileSet()
	xgoPkgs, err := instrument_xgo_runtime.LinkXgoRuntime(goroot, projectDir, xgoRuntimeModuleDir, goVersion, overlayFS, overrideXgoContent, instrument_xgo_runtime.LinkOptions{
		Fset:                fset,
		Mod:                 mod,
		Modfile:             modfile,
		XgoVersion:          VERSION,
		XgoRevision:         REVISION,
		XgoNumber:           NUMBER,
		CollectTestTrace:    collectTestTrace,
		CollectTestTraceDir: collectTestTraceDir,
		ReadRuntimeGenFile: func(path []string) ([]byte, error) {
			return readRuntimeGenFile(xgoSrc, path)
		},
	})
	if err != nil {
		if err == instrument_xgo_runtime.ErrLinkFileNotRequired {
			return nil
		}
		if err == instrument_xgo_runtime.ErrLinkFileNotFound {
			if !goFlag {
				fmt.Fprintf(os.Stderr, `WARNING: xgo: skip runtime instrumentation, upgrade:
  go get %s@latest
  import _ %q
	`, constants.RUNTIME_INTERNAL_TRAP_PKG, constants.RUNTIME_INTERNAL_TRAP_PKG)
			}
			return nil
		}
		if errors.Is(err, instrument_xgo_runtime.ErrRuntimeVersionDeprecatedV1_0_0) {
			if !goFlag {
				if triedUpgrade {
					return fmt.Errorf("xgo v%s auto upgrade failed, you can fix this by:\n  go get %s@latest", VERSION, constants.RUNTIME_MODULE)
				}
				return fmt.Errorf("xgo v%s cannot work with deprecated xgo/runtime: %w, upgrade with:\n  go get %s@latest", VERSION, err, constants.RUNTIME_MODULE)
			}
			return nil
		}
		return err
	}
	funcTabPkg := xgoPkgs.PackageByPath[constants.RUNTIME_FUNCTAB_PKG]
	if funcTabPkg == nil || !hasFunc(funcTabPkg, constants.RUNTIME_FUNCTAB_REGISTER) {
		logDebug("skip functab registering")
		return nil
	}

	includeMain, loadPkgs, err := getLoadPackages(rules)
	if err != nil {
		return err
	}
	logDebug("loadPkgs: includeMain=%v loadPkgs=%v", includeMain, loadPkgs)
	var loadArgs []string
	if includeMain {
		loadArgs = append(loadArgs, mainModule+"/...")
		loadArgs = append(loadArgs, "main")
	}
	loadArgs = append(loadArgs, loadPkgs...)
	loadArgs = append(loadArgs, trapPkgs...)

	pkgs := &edit.Packages{
		Fset: fset,
		LoadOptions: load.LoadOptions{
			Dir:             projectDir,
			Mod:             mod,
			Overlay:         overlayFS,
			IncludeTest:     includeTest,
			ModFile:         modfile,
			MaxFileSize:     MAX_FILE_SIZE,
			FilterErrorFile: true,
			Goroot:          goroot,
			Fset:            fset,
			// always load deps, should be fast enough
			Deps: true,
		},
	}
	logDebug("start load: %v", loadArgs)
	loadPackages, err := load.LoadPackages(loadArgs, pkgs.LoadOptions)
	if err != nil {
		return err
	}
	pkgs.Add(loadPackages)
	// remove deps flag
	pkgs.LoadOptions.Deps = false

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
		_, isMain := goinfo.PkgWithinModule(pkgPath, mainModule)
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

	traverseBegin := time.Now()
	logDebug("traverse: len(mainPkgs)=%d", len(mainPkgs))

	var recorder resolve.Recorder
	err = resolve.Traverse(reg, mainPkgs, &recorder)
	if err != nil {
		return err
	}
	logDebug("traverse: cost=%v", time.Since(traverseBegin))

	logDebug("start instrumentVarTrap: len(mainPkgs)=%d", len(mainPkgs))
	err = instrument_var.TrapVariables(pkgs, &recorder)
	if err != nil {
		return err
	}

	// ""->default
	needTrapAll := trapAll == "true" || (trapAll == "" && recorder.HasTrapInterceptorRef)

	logDebug("start instrumentFuncTrap: len(packages)=%d, needTrapAll=%v", len(pkgs.Packages), needTrapAll)
	for _, pkg := range pkgs.Packages {
		if !pkg.AllowInstrument {
			continue
		}
		pkgPath := pkg.LoadPackage.GoPackage.ImportPath
		cfg := config.GetPkgConfig(pkgPath)
		var defaultAllow bool
		if !pkg.LoadPackage.GoPackage.Standard {
			if needTrapAll || pkg.Initial {
				defaultAllow = true
			}
		}
		for _, file := range pkg.Files {
			if !pkg.Main && strings.HasSuffix(file.File.Name, "_test.go") {
				// skip test files outside main package
				continue
			}
			funcs := instrument_func.TrapFuncs(file.Edit, pkgPath, file.File.Syntax, file.Index, instrument_func.Options{
				PkgRecorder:    recorder.Pkgs[pkgPath],
				PkgConfig:      cfg,
				DefaultDisable: !defaultAllow,
			})
			file.TrapFuncs = append(file.TrapFuncs, funcs...)

			// interface types
			intfTypes := instrument_intf.CollectInterfaces(file)
			file.InterfaceTypes = append(file.InterfaceTypes, intfTypes...)
		}
	}

	logDebug("generate functab register")
	registerFuncTab(pkgs)

	logDebug("collect edits")
	updatedFiles, err := applyInstrumentWithEditNotes(overlayFS, pkgs, mayHaveCover, overrideXgoContent)
	if err != nil {
		return err
	}
	logDebug("finish instruments, updated %d files", updatedFiles)
	return nil
}

func registerFuncTab(packages *edit.Packages) {
	fset := packages.Fset
	for _, pkg := range packages.Packages {
		pkgPath := pkg.LoadPackage.GoPackage.ImportPath
		stdlib := pkg.LoadPackage.GoPackage.Standard
		for _, file := range pkg.Files {
			instrument_reg.RegisterFuncTab(fset, file, pkgPath, stdlib)
		}
	}
}

func applyInstrumentWithEditNotes(overlayFS overlay.Overlay, packages *edit.Packages, mayHaveCover bool, xgoOverrideContent func(absFile overlay.AbsFile, content string)) (int, error) {
	updatedFiles := 0
	for _, pkg := range packages.Packages {
		for _, file := range pkg.Files {
			if !file.Edit.HasEdit() {
				continue
			}
			content := string(file.Edit.Buffer().Bytes())
			if mayHaveCover {
				var err error
				content, err = instrument_go.AddEditsNotes(file.Edit, file.File.AbsPath, file.File.Content, content)
				if err != nil {
					return 0, fmt.Errorf("failed to add edits: %s %w", file.File.AbsPath, err)
				}
			}
			absFile := overlay.AbsFile(file.File.AbsPath)
			if pkg.Xgo && xgoOverrideContent != nil {
				xgoOverrideContent(absFile, content)
			} else {
				overlayFS.OverrideContent(absFile, content)
			}
			updatedFiles++
		}
	}
	return updatedFiles, nil
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

func getLocalXgoGenDir(projectRootDir string) (string, error) {
	xgoDir, err := getLocalXgoDir(projectRootDir)
	if err != nil {
		return "", err
	}
	xgoGenDir := filepath.Join(xgoDir, "gen")

	// to avoid `go mod vendor` detecting this gen dir
	xgoGenGoMod := filepath.Join(xgoGenDir, "go.mod")
	stat, statErr := os.Stat(xgoGenGoMod)
	if statErr != nil {
		if !os.IsNotExist(statErr) {
			return "", statErr
		}
	} else {
		if !stat.IsDir() {
			return xgoGenDir, nil
		}
		err = os.RemoveAll(xgoGenGoMod)
		if err != nil {
			return "", err
		}
	}

	err = os.MkdirAll(xgoGenDir, 0755)
	if err != nil {
		return "", err
	}

	err = os.WriteFile(xgoGenGoMod, []byte("// empty go.mod generated by xgo to avoid affecting `go mod tidy` and `go mod vendor`\n"), 0644)
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
