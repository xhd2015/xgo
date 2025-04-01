package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/xhd2015/xgo/support/fileutil"
	"github.com/xhd2015/xgo/support/git"
	"github.com/xhd2015/xgo/support/instrument"
	"github.com/xhd2015/xgo/support/instrument/constants"
	"github.com/xhd2015/xgo/support/instrument/edit"
	"github.com/xhd2015/xgo/support/instrument/instrument_func"
	"github.com/xhd2015/xgo/support/instrument/instrument_go"
	"github.com/xhd2015/xgo/support/instrument/load"
	"github.com/xhd2015/xgo/support/instrument/overlay"
	"github.com/xhd2015/xgo/support/instrument/patch"
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

func instrumentUserSpace(projectDir string, projectRoot string, mod string, modfile string, mainModule string, xgoRuntimeModuleDir string, mayHaveCover bool, overlayFS overlay.Overlay, includeTest bool, rules []Rule, trapPkgs []string, collectTestTrace bool, collectTestTraceDir string) error {
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
	xgoPkgs, err := instrument.LinkXgoRuntime(projectDir, xgoRuntimeModuleDir, overlayFS, mod, modfile, VERSION, REVISION, NUMBER, collectTestTrace, collectTestTraceDir, overrideXgoContent)
	if err != nil {
		if err != instrument.ErrLinkFileNotFound {
			return err
		}
		fmt.Fprintf(os.Stderr, `WARNING: xgo: skip runtime instrumentation, upgrade:
  go get %s@latest
  import _ %q
`, constants.RUNTIME_TRAP_PKG, constants.RUNTIME_TRAP_PKG)
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
	}
	loadArgs = append(loadArgs, loadPkgs...)
	loadArgs = append(loadArgs, trapPkgs...)

	logDebug("start load: %v", loadArgs)
	loadPackages, err := load.LoadPackages(loadArgs, load.LoadOptions{
		Dir:             projectDir,
		Mod:             mod,
		Overlay:         overlayFS,
		IncludeTest:     includeTest,
		ModFile:         modfile,
		MaxFileSize:     MAX_FILE_SIZE,
		FilterErrorFile: true,
	})
	if err != nil {
		return err
	}

	// insert func trap
	packages := edit.Edit(loadPackages)
	logDebug("start instrumentFuncTrap: len(packages)=%d", len(packages.Packages))
	for _, pkg := range packages.Packages {
		for _, file := range pkg.Files {
			funcs := instrument_func.EditInjectRuntimeTrap(file.Edit, file.File.Syntax)
			file.TrapFuncs = append(file.TrapFuncs, funcs...)
		}
	}

	// trap var for packages in main module
	varPkgs := packages.Filter(func(pkg *edit.Package) bool {
		return pkgWithinModule(pkg.LoadPackage.GoPackage.ImportPath, mainModule)
	})
	logDebug("start instrumentVarTrap: len(varPkgs)=%d", len(varPkgs.Packages))
	err = instrument.InstrumentVarTrap(varPkgs)
	if err != nil {
		return err
	}
	funcTabPkg := xgoPkgs.PackageByPath[constants.RUNTIME_FUNCTAB_PKG]
	if funcTabPkg == nil || !hasFunc(funcTabPkg, constants.RUNTIME_REGISTER_FUNC_TAB) {
		logDebug("skip functab registering")
	} else {
		logDebug("generate functab register")
		collectFuncInfos(xgoPkgs)
		collectFuncInfos(packages)
	}

	logDebug("collect edits")
	updatedFiles, err := addEditNotes(overlayFS, packages, mayHaveCover, nil)
	if err != nil {
		return err
	}
	_, err = addEditNotes(overlayFS, xgoPkgs, mayHaveCover, overrideXgoContent)
	if err != nil {
		return err
	}
	logDebug("finish instruments, updated %d files", updatedFiles)
	return nil
}

func collectFuncInfos(packages *edit.Packages) {
	fset := packages.Fset
	for _, pkg := range packages.Packages {
		for _, file := range pkg.Files {
			const FUNC_INFO = "FuncInfo"
			FUNC_TAB := constants.RUNTIME_PKG_NAME_FUNCTAB
			REGISTER := constants.RUNTIME_REGISTER_FUNC_TAB

			// TODO: add fn and var ptr
			lines := make([]string, 0, len(file.TrapFuncs)+len(file.TrapVars))
			makeLiteral := func(kind string, name string, identityName string, pos token.Pos, extra []string) string {
				var suffix string
				if len(extra) > 0 {
					suffix = "," + strings.Join(extra, ",")
				}
				lineNum := fset.Position(pos).Line
				return fmt.Sprintf("&%s.%s{Kind:%s.%s,Pkg:pkgPath,Name:%q,IdentityName:%q,File:absFile,Line:%d%s}", FUNC_TAB, FUNC_INFO, FUNC_TAB, kind,
					name,
					identityName,
					lineNum,
					suffix,
				)
			}
			for _, varInfo := range file.TrapVars {
				extra := []string{
					"Var:" + "&" + varInfo.Name,
					fmt.Sprintf("ResNames:[]string{%q}", varInfo.Name),
				}

				literal := makeLiteral("Kind_Var", varInfo.Name, varInfo.Name, varInfo.Decl.Decl.Pos(), extra)
				lines = append(lines, fmt.Sprintf("%s.%s(%s)", FUNC_TAB, REGISTER, literal))
			}
			for _, funcInfo := range file.TrapFuncs {
				var extra []string
				identityName := getIdentityName(fset, file.File.Content, funcInfo.FuncDecl)
				if !isGeneric(funcInfo.FuncDecl) {
					extra = append(extra, "Func:"+identityName)
				} else {
					extra = append(extra, "Generic:true")
				}
				if funcInfo.Receiver != nil {
					extra = append(extra, fmt.Sprintf("RecvName:%q", funcInfo.Receiver.Name))
				}
				if len(funcInfo.Params) > 0 {
					extra = append(extra, fmt.Sprintf("ArgNames:[]string{%s}", strings.Join(quoteNames(funcInfo.Params.Names()), ",")))
				}
				if len(funcInfo.Results) > 0 {
					extra = append(extra, fmt.Sprintf("ResNames:[]string{%s}", strings.Join(quoteNames(funcInfo.Results.Names()), ",")))
				}

				literal := makeLiteral("Kind_Func", funcInfo.FuncDecl.Name.Name, identityName, funcInfo.FuncDecl.Pos(), extra)
				lines = append(lines, fmt.Sprintf("%s.%s(%s)", FUNC_TAB, REGISTER, literal))
			}

			if len(lines) == 0 {
				continue
			}
			fileEdit := file.Edit
			fileSyntax := file.File.Syntax

			patch.AddImport(fileEdit, fileSyntax, FUNC_TAB, constants.RUNTIME_FUNCTAB_PKG)

			pkgPath := pkg.LoadPackage.GoPackage.ImportPath
			absFile := file.File.AbsPath
			declLines := make([]string, 0, len(lines)+2)
			declLines = append(declLines,
				fmt.Sprintf("pkgPath:=%q", pkgPath),
				fmt.Sprintf("absFile:=%q", absFile),
			)
			declLines = append(declLines, lines...)

			code := strings.Join(declLines, ";")
			pos := patch.GetFuncInsertPosition(fileSyntax)
			patch.AddCode(fileEdit, pos, "func init(){"+code+";}")
		}
	}
}

func addEditNotes(overlayFS overlay.Overlay, packages *edit.Packages, mayHaveCover bool, overrideContent func(absFile overlay.AbsFile, content string)) (int, error) {
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
			if overrideContent != nil {
				overrideContent(absFile, content)
			} else {
				overlayFS.OverrideContent(absFile, content)
			}
			updatedFiles++
		}
	}
	return updatedFiles, nil
}

func getIdentityName(fset *token.FileSet, code string, funcDecl *ast.FuncDecl) string {
	if funcDecl.Recv == nil || len(funcDecl.Recv.List) == 0 || funcDecl.Recv.List[0].Type == nil {
		return funcDecl.Name.Name
	}
	recv := funcDecl.Recv.List[0]
	recvType := recv.Type

	start := fset.Position(recvType.Pos()).Offset
	end := fset.Position(recvType.End()).Offset
	recvName := code[start:end]

	if strings.HasPrefix(recvName, "*") {
		recvName = "(" + recvName + ")"
	}
	return fmt.Sprintf("%s.%s", recvName, funcDecl.Name.Name)
}

func quoteNames(names []string) []string {
	quotedNames := make([]string, len(names))
	for i, name := range names {
		quotedNames[i] = strconv.Quote(name)
	}
	return quotedNames
}

func isGeneric(funcDecl *ast.FuncDecl) bool {
	if funcDecl.Type.TypeParams == nil || len(funcDecl.Type.TypeParams.List) == 0 {
		return false
	}
	return true
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

func pkgWithinModule(pkgPath string, mainModule string) bool {
	if !strings.HasPrefix(pkgPath, mainModule) {
		return false
	}
	if len(pkgPath) == len(mainModule) {
		return true
	}
	if pkgPath[len(mainModule)] != '/' {
		return false
	}
	return true
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
	trimedList := make([]string, 0, len(list))
	for _, item := range list {
		item := strings.TrimSpace(item)
		if item == "" {
			continue
		}
		trimedList = append(trimedList, item)
	}
	return trimedList
}

func getLocalXgoGenDir(projectDir string) (string, error) {
	xgoDir, err := getLocalXgoDir(projectDir)
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
			return "", fmt.Errorf("%s is not a directory", humanReadablePath(xgoDir))
		}
		return xgoDir, nil
	}
	err := os.MkdirAll(xgoDir, 0755)
	if err != nil {
		return "", err
	}
	topLevel, _ := git.ShowTopLevel(projectDir)
	if topLevel != "" {
		// add .xgo/gen to .gitignore
		err := fileutil.UpdateFile(filepath.Join(topLevel, ".gitignore"), func(content []byte) (bool, []byte, error) {
			if bytes.Contains(content, []byte("**/.xgo/gen\n")) {
				return false, content, nil
			}
			var prefix string
			if len(content) > 0 && !bytes.HasSuffix(content, []byte("\n")) {
				prefix = "\n"
			}
			return true, append(content, []byte(prefix+"**/.xgo/gen\n")...), nil
		})
		if err != nil {
			return "", err
		}
	}

	return xgoDir, nil
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
