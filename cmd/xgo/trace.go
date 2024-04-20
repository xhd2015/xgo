package main

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/xgo/cmd/xgo/pathsum"
	"github.com/xhd2015/xgo/support/cmd"
	"github.com/xhd2015/xgo/support/filecopy"
	"github.com/xhd2015/xgo/support/fileutil"
	"github.com/xhd2015/xgo/support/goinfo"
)

const RUNTIME_MODULE = "github.com/xhd2015/xgo/runtime"
const TRACE_PKG = RUNTIME_MODULE + "/trace"

type importResult struct {
	overlayFile string
	mod         string
	modfile     string
}

//go:embed runtime_gen
var runtimeGenFS embed.FS

// TODO: may apply tags
func importRuntimeDep(test bool, goroot string, goBinary string, goVersion *goinfo.GoVersion, modfile string, xgoSrc string, projectDir string, modRootRel []string, mainModule string, mod string, args []string) (*importResult, error) {
	if mainModule == "" {
		// only work with module
		return nil, nil
	}
	projectDir, err := filepath.Abs(projectDir)
	if err != nil {
		return nil, err
	}
	projectRoot := projectDir
	n := len(modRootRel)
	for i := 0; i < n; i++ {
		projectRoot = filepath.Dir(projectRoot)
	}

	pkgArgs := getPkgArgs(args)
	// check if trace package already exists
	listArgs := []string{"list", "-deps"}
	if test {
		listArgs = append(listArgs, "-test")
	}
	listArgs = append(listArgs, pkgArgs...)
	logDebug("go %v", listArgs)
	output, err := cmd.Dir(projectDir).Env([]string{
		"GOROOT=" + goroot,
	}).Output(goBinary, listArgs...)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		pkg := strings.TrimSpace(line)
		if pkg == TRACE_PKG {
			return nil, nil
		}
	}
	var vendorDir string
	if mod != "mod" {
		// not forcing mod
		testVendorDir := filepath.Join(projectRoot, "vendor")
		_, statErr := os.Stat(testVendorDir)
		if statErr != nil {
			if !errors.Is(statErr, os.ErrNotExist) {
				return nil, statErr
			}
		} else {
			vendorDir = testVendorDir
		}
	}

	overlayInfo, err := createOverlay(goroot, goBinary, goVersion, modfile, xgoSrc, test, projectRoot, projectDir, vendorDir, pkgArgs)
	if err != nil {
		return nil, err
	}
	res := &importResult{}
	if overlayInfo != nil {
		res.overlayFile = overlayInfo.overlayFile
		res.mod = overlayInfo.mod
		res.modfile = overlayInfo.modfile
	}

	return res, nil
}

type Overlay struct {
	Replace map[string]string
}

type overlayInfo struct {
	overlayFile string
	mod         string
	modfile     string // alternative go.mod
}

func createOverlay(goroot string, goBinary string, goVersion *goinfo.GoVersion, modfileOption string, xgoSrc string, test bool, projectRoot string, projectDir string, vendorDir string, pkgArgs []string) (*overlayInfo, error) {
	// try /tmp first
	tmpDir := "/tmp"
	_, statErr := os.Stat(tmpDir)
	if statErr != nil {
		tmpDir = os.TempDir()
	}

	xgoTmp := filepath.Join(tmpDir, "xgo_"+fileutil.CleanSpecial(getRevision()))
	err := os.MkdirAll(xgoTmp, 0755)
	if err != nil {
		return nil, err
	}
	logDebug("xgo tmp dir: %s", xgoTmp)

	suffix := ""
	if isDevelopment {
		suffix = "_dev"
	}
	tmpRuntime := filepath.Join(xgoTmp, "runtime"+suffix)
	err = os.MkdirAll(tmpRuntime, 0755)
	if err != nil {
		return nil, err
	}
	if isDevelopment {
		err = filecopy.CopyReplaceDir(filepath.Join(xgoSrc, "runtime"), tmpRuntime, false)
		if err != nil {
			return nil, err
		}
	} else {
		exists := func(path string) bool {
			_, err := os.Stat(path)
			return err == nil
		}
		files := []string{"core", "trap", "trace", "go.mod"}
		allExists := true
		for _, file := range files {
			if !exists(filepath.Join(tmpRuntime, file)) {
				allExists = false
				break
			}
		}
		//  cache copy
		if !allExists {
			err := os.RemoveAll(tmpRuntime)
			if err != nil {
				return nil, err
			}
			err = copyEmbedDir(runtimeGenFS, "runtime_gen", tmpRuntime)
			if err != nil {
				return nil, err
			}
			err = os.Rename(filepath.Join(tmpRuntime, "go.mod.txt"), filepath.Join(tmpRuntime, "go.mod"))
			if err != nil {
				return nil, err
			}
		}
	}

	// create project
	projectSum, err := pathsum.PathSum("", projectRoot)
	if err != nil {
		return nil, err
	}
	tmpProjectDir := filepath.Join(xgoTmp, "projects", projectSum)
	logDebug("tmp project dir: %s", tmpProjectDir)
	err = os.MkdirAll(tmpProjectDir, 0755)
	if err != nil {
		return nil, err
	}

	goMod := modfileOption
	if modfileOption == "" {
		goMod = filepath.Join(projectRoot, "go.mod")
	}

	tmpGoMod := filepath.Join(tmpProjectDir, "go.mod")
	err = filecopy.CopyFile(goMod, tmpGoMod)
	if err != nil {
		return nil, err
	}

	// use -modfile ?
	var modfile string
	var mod string
	goModReplace := make(map[string]string)
	if vendorDir != "" {
		var hasReplace bool
		editArgs := []string{"mod", "edit"}

		// read vendor/modules.txt,
		vendorInfo, err := goinfo.ParseVendor(vendorDir)
		if err != nil {
			return nil, err
		}
		// get all modules, convert all deps to replace
		for _, mod := range vendorInfo.VendorList {
			modInfo, ok := vendorInfo.VendorMeta[mod]
			if !ok {
				continue
			}
			modPath := mod.Path
			if modInfo.Replacement.Path != "" {
				modPath = modInfo.Replacement.Path
			}
			vendorModPath := filepath.Join(vendorDir, modPath)
			vendorModFile := filepath.Join(vendorModPath, "go.mod")
			replaceModFile := filepath.Join(tmpProjectDir, vendorModFile)
			// replace goMod => vendor=>

			editArgs = append(editArgs, fmt.Sprintf("-replace=%s=%s", modPath, vendorModPath))
			hasReplace = true
			// create placeholder go.mod for each module
			modGoVersion := modInfo.GoVersion
			if modGoVersion == "" {
				modGoVersion = fmt.Sprintf("%d.%d", goVersion.Major, goVersion.Minor)
			}
			err := createGoModPlaceholder(replaceModFile, modPath, modGoVersion)
			if err != nil {
				return nil, err
			}
			goModReplace[vendorModFile] = replaceModFile
		}

		if hasReplace {
			if false {
				// go.sum needs to be synced?
				goSum := filepath.Join(projectRoot, "go.sum")
				tmpGoSum := filepath.Join(tmpProjectDir, "go.sum")
				// write an empty go.sum
				err := os.WriteFile(tmpGoSum, nil, 0755)
				if err != nil {
					return nil, err
				}
				goModReplace[goSum] = tmpGoSum
			}
			// force use -mod=mod
			mod = "mod" // force use mod after replaced vendor
			modfile = tmpGoMod
			editArgs = append(editArgs, tmpGoMod)
			err = cmd.Env([]string{
				"GOROOT=" + goroot,
			}).Run(goBinary, editArgs...)
			if err != nil {
				return nil, err
			}
		}
	}

	err = cmd.Env([]string{
		"GOROOT=" + goroot,
	}).Run(goBinary, "mod", "edit",
		fmt.Sprintf("-require=%s@v%s", RUNTIME_MODULE, VERSION),
		fmt.Sprintf("-replace=%s=%s", RUNTIME_MODULE, tmpRuntime),
		tmpGoMod,
	)
	if err != nil {
		return nil, err
	}

	// list files, add init
	// NOTE: go build tag applies,
	// ignored files will be placed to IgnoredGoFiles

	listArgs := []string{"list", "-json"}
	listArgs = append(listArgs, pkgArgs...)
	output, err := cmd.Dir(projectDir).Env([]string{
		"GOROOT=" + goroot,
	}).Output(goBinary, listArgs...)
	if err != nil {
		return nil, err
	}
	var pkgs []*GoListPkg
	dec := json.NewDecoder(strings.NewReader(output))
	for dec.More() {
		var pkg *GoListPkg
		err := dec.Decode(&pkg)
		if err != nil {
			return nil, err
		}
		pkgs = append(pkgs, pkg)
	}

	replace := make(map[string]string)
	for _, pkg := range pkgs {
		if pkg.Standard {
			continue
		}
		var file string
		if test && len(pkg.TestGoFiles) > 0 {
			file = pkg.TestGoFiles[0]
		} else if len(pkg.GoFiles) > 0 {
			file = pkg.GoFiles[0]
		}
		if file == "" {
			// no files
			continue
		}
		srcFile := filepath.Join(pkg.Root, file)
		dstFile := filepath.Join(tmpProjectDir, srcFile)
		err := filecopy.CopyFileAll(srcFile, dstFile)
		if err != nil {
			return nil, err
		}
		// add blank import
		content, err := os.ReadFile(dstFile)
		if err != nil {
			return nil, err
		}
		newContent, ok := addBlankImport(string(content))
		if !ok {
			continue
		}
		err = os.WriteFile(dstFile, []byte(newContent), 0755)
		if err != nil {
			return nil, err
		}
		replace[srcFile] = dstFile
	}

	if modfile == "" {
		replace[goMod] = tmpGoMod
	}
	for k, v := range goModReplace {
		replace[k] = v
	}

	overlay := Overlay{Replace: replace}
	overlayData, err := json.Marshal(overlay)
	if err != nil {
		return nil, err
	}
	overlayFile := filepath.Join(tmpProjectDir, "overlay.json")
	err = os.WriteFile(overlayFile, overlayData, 0755)
	if err != nil {
		return nil, err
	}
	return &overlayInfo{
		overlayFile: overlayFile,
		mod:         mod,
		modfile:     modfile,
	}, nil
}

type GoListPkg struct {
	Dir         string
	ImportPath  string
	Root        string
	Standard    bool
	GoFiles     []string
	TestGoFiles []string
}

func addBlankImport(content string) (string, bool) {
	var base int
	idx := strings.Index(content, "package ")
	if idx < 0 {
		return "", false
	}
	base += idx
	subContent := content[idx:]
	rIdx := strings.Index(subContent, "\r")
	if rIdx > 0 {
		base += rIdx
		subContent = subContent[rIdx+1:]
	}
	q := fmt.Sprintf(";import _ %q", TRACE_PKG)
	nIdx := strings.Index(subContent, "\n")
	if nIdx < 0 {
		return content + q, true
	}
	base += nIdx
	return content[:base] + q + content[base:], true
}

func createGoModPlaceholder(file string, modPath string, goVersion string) error {
	dir := filepath.Dir(file)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}
	content := fmt.Sprintf("module %s\ngo %s\n", modPath, goVersion)
	err = os.WriteFile(file, []byte(content), 0755)
	if err != nil {
		return err
	}
	return nil
}
