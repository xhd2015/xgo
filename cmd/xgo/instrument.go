package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/xgo/support/fileutil"
	"github.com/xhd2015/xgo/support/git"
	"github.com/xhd2015/xgo/support/instrument"
	"github.com/xhd2015/xgo/support/instrument/edit"
	"github.com/xhd2015/xgo/support/instrument/instrument_xgo_runtime"
	"github.com/xhd2015/xgo/support/instrument/load"
	"github.com/xhd2015/xgo/support/instrument/overlay"
)

func instrumentUserSpace(projectDir string, projectRoot string, mod string, modfile string, mainModule string, overlayFS overlay.Overlay, includeTest bool, rules []Rule, trapPkgs []string, collectTestTrace bool, collectTestTraceDir string) error {
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
	err := instrument.LinkXgoRuntime(projectDir, overlayFS, modfile, VERSION, REVISION, NUMBER, collectTestTrace, collectTestTraceDir)
	if err != nil {
		if err != instrument.ErrLinkFileNotFound {
			return err
		}
		fmt.Fprintf(os.Stderr, `WARNING: xgo: skipping runtime instrumentation, requires upgrade %s
  go get %s@latest
  import _ %q`, instrument_xgo_runtime.RUNTIME_TRAP_PKG, instrument_xgo_runtime.RUNTIME_TRAP_PKG, instrument_xgo_runtime.RUNTIME_TRAP_PKG)
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
	loadArgs = append(loadArgs,
		instrument_xgo_runtime.RUNTIME_TRACE_SIGNAL_PKG,
	)

	logDebug("start load: %v", loadArgs)
	loadPackages, err := load.LoadPackages(loadArgs, load.LoadOptions{
		Dir:         projectDir,
		Mod:         mod,
		Overlay:     overlayFS,
		IncludeTest: includeTest,
		ModFile:     modfile,
	})
	if err != nil {
		return err
	}
	packages := edit.Edit(loadPackages)

	logDebug("start instrumentFuncTrap: len(packages)=%d", len(packages.Packages))
	err = instrument.InstrumentFuncTrap(packages)
	if err != nil {
		return err
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
	logDebug("collect edits")
	updatedFiles := 0
	for _, pkg := range packages.Packages {
		for _, file := range pkg.Files {
			if !file.Edit.HasEdit() {
				continue
			}
			content := string(file.Edit.Buffer().Bytes())
			overlayFS.OverrideContent(overlay.AbsFile(file.File.AbsPath), content)
			updatedFiles++
		}
	}
	logDebug("finish instruments, updated %d files", updatedFiles)
	return nil
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
