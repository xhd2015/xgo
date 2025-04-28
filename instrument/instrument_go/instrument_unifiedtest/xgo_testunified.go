// the file will be copied to GOROOT/src/cmd/go/internal/test

//go:build ignore

package test

import (
	"bytes"
	"cmd/go/internal/load"
	"cmd/internal/test2json"
	"context"
	"fmt"
	"go/build"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// search for formatTestmain and testmainTmpl

// import _test "...."
// _test.X

// check builderTest for per-package build
var _ = builderTest
var _ = load.TestPackagesFor

func xgoUnifyTestPackages(ctx context.Context, pkgs []*load.Package) []*load.Package {
	if os.Getenv("XGO_UNIFIEDTEST") != "true" {
		return pkgs
	}

	const STUB_GO = "stub.go"
	unified := &load.Package{
		PackagePublic: load.PackagePublic{
			Dir:         "", // filled later
			ImportPath:  xgoUnifiedPackage,
			Name:        xgoUnifiedPkgName,
			TestGoFiles: []string{STUB_GO},
		},
		Internal: load.PackageInternal{
			Build: &build.Package{},
		},
	}

	neededImports := make(map[string]bool)
	var neededImportList []string

	var edits []func()

	// load test funcs for each package
	var extra []*load.Package
	for _, pkg := range pkgs {
		pkg := pkg
		if len(pkg.XTestGoFiles) > 0 {
			extra = append(extra, pkg)
			continue
		}
		t, err := load.XgoLoadTestFuncs(pkg)
		if err != nil {
			log.Fatalf("failed to load test funcs for package %s: %s", pkg.ImportPath, err)
		}
		// skip test main
		if t.TestMain != nil {
			extra = append(extra, pkg)
			continue
		}
		// don't do clone, there are other imports
		// reference this package, we cannot modify
		// all of them
		//
		//   clonePkg = new(pkg)
		//   *clonePkg = *pkg
		clonePkg := pkg

		edits = append(edits, func() {
			// record imports
			for _, imp := range pkg.TestImports {
				if !neededImports[imp] {
					neededImports[imp] = true
					neededImportList = append(neededImportList, imp)
				}
			}

			// clone files
			mergedFiles := make([]string, 0, len(pkg.GoFiles)+len(pkg.TestGoFiles))
			mergedFiles = append(mergedFiles, pkg.GoFiles...)
			mergedFiles = append(mergedFiles, pkg.TestGoFiles...)

			mergedImports := make([]string, 0, len(pkg.Imports)+len(pkg.TestImports))
			mergedImports = append(mergedImports, pkg.Imports...)
			mergedImports = append(mergedImports, pkg.TestImports...)

			// clonedInternalImports := make([]*load.Package, len(pkg.Internal.Imports))
			// copy(clonedInternalImports, pkg.Internal.Imports)

			clonePkg.GoFiles = mergedFiles
			clonePkg.Imports = mergedImports
			// clonePkg.Internal.Imports = clonedInternalImports
		})

		var pkgTests []*xgoPkgTestInfo
		for _, test := range t.Tests {
			pkgTests = append(pkgTests, &xgoPkgTestInfo{
				Name: test.Name,
				Ref:  test.Name,
			})
		}
		unified.PackagePublic.TestImports = append(unified.PackagePublic.TestImports, pkg.ImportPath)
		unified.Internal.Imports = append(unified.Internal.Imports, clonePkg)
		xgoTestPackages = append(xgoTestPackages, &xgoTestPackage{
			LoadPackage: clonePkg,
			Tests:       pkgTests,
		})
	}

	if len(xgoTestPackages) <= 1 {
		return pkgs
	}

	// apply edits
	for _, edit := range edits {
		edit()
	}

	// prepare stub.go
	tmpDir, err := os.MkdirTemp("", xgoUnifiedPkgName)
	if err != nil {
		log.Fatalf("failed to create temp dir: %s", err)
	}
	unified.PackagePublic.Dir = tmpDir
	xgoTmpDir = tmpDir
	err = os.WriteFile(filepath.Join(tmpDir, STUB_GO), []byte("package "+xgoUnifiedPkgName), 0644)
	if err != nil {
		log.Fatalf("failed to write stub file: %s", err)
	}

	// load required imports, and add to internal imports
	// to avoid the `could not import XXXX (open : no such file or directory)` issue
	// see issue https://github.com/xhd2015/xgo/issues/311#issuecomment-2833452517
	loadedPkgs := xgoResolveNeededImports(ctx, neededImportList, pkgs)
	for _, testPkg := range xgoTestPackages {
		tp := testPkg.LoadPackage
		for _, imp := range tp.TestImports {
			impPkg := loadedPkgs[imp]
			if impPkg == nil {
				panic(fmt.Errorf("failed to resolve import %s for package %s", imp, tp.ImportPath))
			}
			tp.Internal.Imports = append(tp.Internal.Imports, impPkg)
		}
	}

	// uniq names
	xgoTestInfoMapping = make(map[string]xgoTestInfo)
	dupName := make(map[string]int)
	for _, testPkg := range xgoTestPackages {
		loadPkg := testPkg.LoadPackage
		pkgPath := loadPkg.ImportPath
		pkgDir := loadPkg.Dir
		for _, test := range testPkg.Tests {
			testName := test.Name

			j := dupName[testName]
			dupName[testName] = j + 1

			uniqName := testName
			if j > 0 {
				// this changes test.Name, result in test.Name != test.Ref
				uniqName = fmt.Sprintf("%s_%d", testName, j)
				test.Name = uniqName
			}

			xgoTestInfoMapping[uniqName] = xgoTestInfo{
				Package: pkgPath,
				Dir:     pkgDir,
				Ref:     test.Ref,
			}
		}
	}
	if testJSON {
		test2json.XgoGetTestName = xgoGetTestName
	}

	// link
	load.XgoAfterGenerateTestMain = xgoAfterGenerateTestMain

	xgoUnified = unified
	merged := make([]*load.Package, 0, len(extra)+1)
	merged = append(merged, unified)
	merged = append(merged, extra...)
	return merged
}

func xgoCleanup() {
	if xgoTmpDir != "" {
		os.RemoveAll(xgoTmpDir)
	}
}

func xgoAfterGenerateTestMain(p *load.Package, pmain *load.Package, code []byte) []byte {
	if p != xgoUnified {
		return code
	}

	var testDefs []string
	var nameDirPairs []string
	var testImports []string
	for i, pkg := range xgoTestPackages {
		pkgAlias := xgoGetPkgAlias(i)
		if len(pkg.Tests) > 0 {
			testImports = append(testImports, fmt.Sprintf("import %s %q", pkgAlias, pkg.LoadPackage.ImportPath))
		}
		// var tests = []testing.InternalTest{}
		// replace all testinfo.GetTestName(test) with pkgAlias
		for _, test := range pkg.Tests {
			testInfo, ok := xgoTestInfoMapping[test.Name]
			if ok {
				if testInfo.Dir != "" {
					nameDirPairs = append(nameDirPairs, fmt.Sprintf("%q: %q", test.Name, testInfo.Dir))
				}
			}
			testDefs = append(testDefs, fmt.Sprintf("{Name: %q, F: %s.%s}", test.Name, pkgAlias, test.Ref))
		}

		pmain.Internal.Imports = append(pmain.Internal.Imports, pkg.LoadPackage)
	}

	// check instrument/instrument_runtime/instrument_testing/testing.go
	if len(nameDirPairs) > 0 {
		mapLiteral := "map[string]string{" + strings.Join(nameDirPairs, ",") + "}"
		funcMain := "func main() {"
		code = bytes.ReplaceAll(code, []byte(funcMain), []byte(funcMain+"testing.XgoTestNameToDir = "+mapLiteral+";"))
	}
	if len(testImports) > 0 {
		code = bytes.Replace(code, []byte("package main"), []byte("package main;"+strings.Join(testImports, ";")), 1)
	}
	if len(testDefs) > 0 {
		testDefLine := "var tests = []testing.InternalTest{"
		code = bytes.Replace(code, []byte(testDefLine), []byte(testDefLine+strings.Join(testDefs, ",")+","), 1)
	}

	// DEBUG
	// os.WriteFile(filepath.Join("/tmp/__xgo_debug_main.go"), code, 0644)

	return code
}

func xgoResolveNeededImports(ctx context.Context, needPkgs []string, loadedPkgs []*load.Package) map[string]*load.Package {
	pkgsMapping := make(map[string]*load.Package, len(needPkgs))
	var missing []string
	for _, needPkg := range needPkgs {
		var found *load.Package
		for _, pkg := range loadedPkgs {
			if pkg.ImportPath == needPkg {
				found = pkg
				break
			}
		}
		if found != nil {
			pkgsMapping[needPkg] = found
		} else {
			missing = append(missing, needPkg)
		}
	}
	if len(missing) > 0 {
		loadedMissingPkgs := load.PackagesAndErrors(ctx, load.PackageOpts{}, missing)
		for _, pkg := range loadedMissingPkgs {
			pkgsMapping[pkg.ImportPath] = pkg
		}
	}
	return pkgsMapping
}

func xgoGetCmdDir(pkg *load.Package) string {
	if pkg != xgoUnified {
		return ""
	}
	if len(xgoUnified.Internal.Imports) == 0 {
		return ""
	}
	return xgoUnified.Internal.Imports[0].Dir
}
