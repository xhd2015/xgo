package main

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/xhd2015/xgo/support/cmd"
	"github.com/xhd2015/xgo/support/filecopy"
	"github.com/xhd2015/xgo/support/fileutil"
	"github.com/xhd2015/xgo/support/goinfo"
	"github.com/xhd2015/xgo/support/goparse"
	"github.com/xhd2015/xgo/support/instrument/overlay"
)

const RUNTIME_MODULE = "github.com/xhd2015/xgo/runtime"
const RUNTIME_TRACE_PKG = "github.com/xhd2015/xgo/runtime/trace"

type importResult struct {
	mod              string
	modfile          string
	runtimeModuleDir string
	// modReplace contains:
	//   /project/go.sum => /project/.xgo/gen/overlay/project/go.sum
	//   /project/vendor/some/pkg/go.mod => /project/.xgo/gen/overlay/project/vendor/some/pkg/go.mod
	// and also a go.sum update:
	//   /project/go.sum => /project/.xgo/gen/overlay/project/go.sum
	modReplace  map[overlay.AbsFile]overlay.AbsFile
	fileReplace map[overlay.AbsFile]overlay.AbsFile
}

//go:embed runtime_gen
var runtimeGenFS embed.FS

func getProjectRoot(projectDir string, modRootRel []string) string {
	projectRoot := projectDir
	n := len(modRootRel)
	for i := 0; i < n; i++ {
		projectRoot = filepath.Dir(projectRoot)
	}
	return projectRoot
}

// projectRoot is the directory containing go.mod
func getVendorDir(projectRoot string) (string, error) {
	// not forcing mod
	vendor := filepath.Join(projectRoot, "vendor")
	_, statErr := os.Stat(vendor)
	if statErr != nil {
		if !errors.Is(statErr, os.ErrNotExist) {
			return "", statErr
		}
		return "", nil
	}
	return vendor, nil
}

// importRuntimeDepGenOverlay detect if the target package already
// has github.com/xhd2015/xgo/runtime as dependency,
// if not, dynamically modify the go.mod to include that,
// and add a blank import in the main package
// after the dependency is added, the -mod flag will be forced setting to -mod=mod
// even if the original build has -mod=vendor.
// by employing this technique with -modfile=replacedModFile, we can instruct go
// to build with different modules.
// TODO: may apply tags
func importRuntimeDepGenOverlay(test bool, mayHaveCover bool, goroot string, goBinary string, goVersion *goinfo.GoVersion, absModFile string, xgoSrc string, projectDir string, projectRoot string, localXgoGenDir string, mainModule string, mod string, forceCopyRuntime bool, args []string) (*importResult, error) {
	if mainModule == "" {
		// only work with module
		return nil, nil
	}
	projectDir, err := filepath.Abs(projectDir)
	if err != nil {
		return nil, err
	}
	var vendorDir string
	if mod == "vendor" || mod == "" {
		vendorDir, err = getVendorDir(projectDir)
		if err != nil {
			return nil, err
		}
	}
	if mod == "vendor" && vendorDir == "" {
		return nil, fmt.Errorf("-mod=vendor: vendor dir not found")
	}
	needLoad, err := checkNeedLoadDep(goroot, goBinary, projectRoot, vendorDir, absModFile)
	if err != nil {
		return nil, err
	}

	// use .xgo/gen/runtime, .xgo/gen/project
	// to hold tmp files
	absGenDir, err := filepath.Abs(localXgoGenDir)
	if err != nil {
		return nil, err
	}
	overlayDir := filepath.Join(absGenDir, "overlay")
	err = os.MkdirAll(overlayDir, 0755)
	if err != nil {
		return nil, err
	}

	modulesDir := filepath.Join(absGenDir, "modules")
	err = os.MkdirAll(modulesDir, 0755)
	if err != nil {
		return nil, err
	}
	suffix := ""
	if isDevelopment {
		suffix = "_dev"
	}

	res := &importResult{}
	if needLoad {
		runtimeModuleDir := filepath.Join(modulesDir, filepath.Join(strings.Split(RUNTIME_MODULE, "/")...)+suffix)
		logDebug("loading dependency: %s into %s", RUNTIME_MODULE, runtimeModuleDir)
		err := os.MkdirAll(runtimeModuleDir, 0755)
		if err != nil {
			return nil, err
		}

		overlayInfo, err := loadDependency(goroot, goBinary, goVersion, absModFile, xgoSrc, projectRoot, vendorDir, overlayDir, runtimeModuleDir, forceCopyRuntime)
		if err != nil {
			return nil, err
		}
		if overlayInfo != nil {
			logDebug("go mod replace: %d", len(overlayInfo.modReplace))
			res.modReplace = overlayInfo.modReplace
			res.mod = overlayInfo.mod
			res.modfile = overlayInfo.modfile
			res.runtimeModuleDir = runtimeModuleDir
		}
	}

	pkgArgs := getPkgArgs(args)
	fileReplace, err := addBlankImports(goroot, goBinary, projectDir, pkgArgs, test, mayHaveCover, overlayDir)
	if err != nil {
		return nil, err
	}

	absFileReplace := make(map[overlay.AbsFile]overlay.AbsFile, len(fileReplace))
	for k, v := range fileReplace {
		absFileReplace[overlay.AbsFile(k)] = overlay.AbsFile(v)
	}
	res.fileReplace = absFileReplace
	logDebug("go file replace: %d", len(fileReplace))

	return res, nil
}

// when listing in vendor mod, only Path and Version is effective
type ListModule struct {
	Path    string
	Main    bool
	Version string
	Dir     string
	GoMod   string
	Error   *ListError
}
type ListError struct {
	Err string
}

func checkNeedLoadDep(goroot string, goBinary string, projectRoot string, vendorDir string, modfile string) (bool, error) {
	effectiveMod := "mod"
	if vendorDir != "" {
		effectiveMod = "vendor"
	}
	// -e: suppress error
	//  if github.com/xhd2015/xgo/runtime does not exists,
	//  err: "module github.com/xhd2015/xgo/runtime: not a known dependency"
	listArgs := []string{"list", "-m", "-json", "-mod=" + effectiveMod, "-e"}
	if modfile != "" {
		listArgs = append(listArgs, "-modfile", modfile)
	}
	listArgs = append(listArgs, RUNTIME_MODULE)

	logDebug("go %v", listArgs)
	// go list -m -json -mod=$effective_mod -modfile $modfile -e github.com/xhd2015/xgo
	output, err := cmd.Dir(projectRoot).Env([]string{
		"GOROOT=" + goroot,
	}).Output(goBinary, listArgs...)
	if err != nil {
		return false, err
	}
	var listModule *ListModule
	dec := json.NewDecoder(strings.NewReader(output))
	if dec.More() {
		err := dec.Decode(&listModule)
		if err != nil {
			return false, err
		}
	}
	if listModule == nil {
		return true, nil
	}
	if listModule.Error != nil {
		logDebug("list err: %s", listModule.Error.Err)
		return true, nil
	}
	if effectiveMod != "vendor" {
		return false, nil
	}
	// check if vendor/${trace}/trace.go exists
	if !isFile(filepath.Join(vendorDir, RUNTIME_TRACE_PKG, "trace.go")) {
		return true, nil
	}
	return false, nil
}

type dependencyInfo struct {
	modReplace map[overlay.AbsFile]overlay.AbsFile
	// the -mod flag
	mod string
	// the -modfile flag
	modfile string // alternative go.mod
}

func loadDependency(goroot string, goBinary string, goVersion *goinfo.GoVersion, modfileOption string, xgoSrc string, projectRoot string, vendorDir string, tmpOverlayDir string, tmpRuntime string, forceCopyRuntime bool) (*dependencyInfo, error) {
	if isDevelopment {
		err := filecopy.CopyReplaceDir(filepath.Join(xgoSrc, "runtime"), tmpRuntime, false)
		if err != nil {
			return nil, err
		}
	} else {
		var skipCopy bool
		if !forceCopyRuntime {
			ok, err := needCopyRuntimeFromXgo(tmpRuntime)
			if err != nil {
				return nil, err
			}
			if !ok {
				skipCopy = true
			}
		}

		//  cache copy
		if !skipCopy {
			logDebug("extracting runtime from xgo to %s", tmpRuntime)
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

	goMod := modfileOption
	if modfileOption == "" {
		goMod = filepath.Join(projectRoot, "go.mod")
	}
	tmpGoMod := fileutil.RebaseAbs(tmpOverlayDir, goMod)
	err := filecopy.CopyFileAll(goMod, tmpGoMod)
	if err != nil {
		return nil, err
	}

	// use -modfile ?
	var modfile string
	var mod string
	goModReplace := make(map[overlay.AbsFile]overlay.AbsFile)
	if vendorDir != "" {
		var hasReplace bool

		// read vendor/modules.txt,
		vendorInfo, err := goinfo.ParseVendor(vendorDir)
		if err != nil {
			return nil, err
		}
		replacedModules := make([]string, 0, len(vendorInfo.VendorList))
		// get all modules, convert all deps to replace
		for _, mod := range vendorInfo.VendorList {
			// modInfo is completely optional
			modInfo := vendorInfo.VendorMeta[mod]
			modPath := mod.Path
			if modInfo.Replacement.Path != "" &&
				!isLocalReplace(modInfo.Replacement.Path) { // !isLocalReplace, see https://github.com/xhd2015/xgo/issues/87#issuecomment-2074722912
				modPath = modInfo.Replacement.Path
			}
			vendorModPath := filepath.Join(vendorDir, modPath)
			vendorModFile := filepath.Join(vendorModPath, "go.mod")
			replaceModFile := fileutil.RebaseAbs(tmpOverlayDir, vendorModFile)
			// replace goMod => vendor=>

			// NOTE: if replace without require, go will automatically add
			//     version v0.0.0-00010101000000-000000000000
			// https://stackoverflow.com/questions/58012771/go-mod-fails-to-find-version-v0-0-0-00010101000000-000000000000-of-a-dependency
			//
			// to suppress this message, always add require
			version := mod.Version
			if version == "" {
				version = "v0.0.0-00010101000000-000000000000"
			}
			requireCmd := fmt.Sprintf("-require=%s@%s", modPath, version)

			// logDebug("replace %s -> %s", modPath, vendorModPath)
			replacedModules = append(replacedModules,
				requireCmd,
				fmt.Sprintf("-replace=%s=%s", modPath, vendorModPath),
			)
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
			goModReplace[overlay.AbsFile(vendorModFile)] = overlay.AbsFile(replaceModFile)
		}
		logDebug("num replaced modules: %d", len(replacedModules)/2)

		if hasReplace {
			if false {
				// go.sum needs to be synced?
				goSum := filepath.Join(projectRoot, "go.sum")
				tmpGoSum := fileutil.RebaseAbs(tmpOverlayDir, goSum)
				err := os.MkdirAll(filepath.Dir(tmpGoSum), 0755)
				if err != nil {
					return nil, err
				}
				// write an empty go.sum
				err = os.WriteFile(tmpGoSum, nil, 0755)
				if err != nil {
					return nil, err
				}
				goModReplace[overlay.AbsFile(goSum)] = overlay.AbsFile(tmpGoSum)
			}
			// force use -mod=mod
			mod = "mod" // force use mod after replaced vendor
			modfile = tmpGoMod
			for _, replaceModuleGroup := range splitArgsBatch(replacedModules, 100) {
				// go mod edit -replace go.mod
				editArgs := make([]string, 0, 3+len(replaceModuleGroup))
				editArgs = append(editArgs, "mod", "edit")
				editArgs = append(editArgs, replaceModuleGroup...)
				editArgs = append(editArgs, tmpGoMod)
				err = cmd.Env([]string{
					"GOROOT=" + goroot,
				}).Run(goBinary, editArgs...)
				if err != nil {
					return nil, err
				}
			}

		}
	}

	logDebug("require %s v%s, replaced: %s", RUNTIME_MODULE, VERSION, tmpRuntime)
	err = cmd.Env([]string{
		"GOROOT=" + goroot,
	}).Run(goBinary, "mod", "edit",
		fmt.Sprintf("-require=%s@v%s", RUNTIME_MODULE, VERSION),
		fmt.Sprintf("-replace=%s=%s", RUNTIME_MODULE, tmpRuntime),
		tmpGoMod,
	)
	logDebug("copy and edit go.mod: %s", tmpGoMod)
	if err != nil {
		return nil, err
	}
	if modfile == "" {
		goModReplace[overlay.AbsFile(goMod)] = overlay.AbsFile(tmpGoMod)
	}

	return &dependencyInfo{
		modReplace: goModReplace,
		mod:        mod,
		modfile:    modfile,
	}, nil
}

func needCopyRuntimeFromXgo(targetRuntimeDir string) (bool, error) {
	const runtimeGenRoot = "runtime_gen"
	// compare version
	versionFilePath := []string{"core", "version.go"}

	fsVersionFile := filepath.Join(targetRuntimeDir, filepath.Join(versionFilePath...))
	fsVersion, readFsVersionErr := os.ReadFile(fsVersionFile)
	if readFsVersionErr != nil {
		if !os.IsNotExist(readFsVersionErr) {
			return false, readFsVersionErr
		}
		return true, nil
	}
	if len(fsVersion) == 0 {
		return true, nil
	}
	// you need to run `go run ./script/generate` to sync runtime and cmd/xgo/runtime_gen
	embedPath := concatEmbedPath(runtimeGenRoot, joinEmbedPath(versionFilePath))
	emebedVersion, err := runtimeGenFS.ReadFile(embedPath)
	if err != nil {
		return false, err
	}

	fsConstants, fsErr := goparse.ExtractConstants(fsVersionFile, string(fsVersion))
	if fsErr != nil {
		return true, nil
	}
	embedConstants, err := goparse.ExtractConstants(embedPath, string(emebedVersion))
	if err != nil {
		return false, err
	}

	fsNumber, ok := fsConstants["NUMBER"]
	if !ok {
		return true, nil
	}
	fsNumberInt, _ := strconv.Atoi(fsNumber)
	if fsNumberInt <= 0 {
		return true, nil
	}
	embedNumber, ok := embedConstants["NUMBER"]
	if !ok {
		return true, nil
	}
	embedNumberInt, _ := strconv.Atoi(embedNumber)
	if embedNumberInt <= 0 {
		return true, nil
	}
	if fsNumberInt < embedNumberInt {
		logDebug("fs version number: %d < embed version number: %d, need extracting from xgo", fsNumberInt, embedNumberInt)
		return true, nil
	}
	logDebug("fs version number: %d >= embed version number: %d, skip extracting from xgo", fsNumberInt, embedNumberInt)
	return false, nil
}

// system may have limit on single go command
func splitArgsBatch(args []string, limit int) [][]string {
	if len(args) <= limit {
		return [][]string{args}
	}
	n := len(args) / limit
	remain := len(args) - n*limit
	bcap := n
	if remain > 0 {
		bcap++
	}
	groups := make([][]string, bcap)
	off := 0
	for i := 0; i < n; i++ {
		group := make([]string, limit)
		for j := 0; j < limit; j++ {
			group[j] = args[off]
			off++
		}
		groups[i] = group
	}
	if remain > 0 {
		group := make([]string, remain)
		for i := 0; i < remain; i++ {
			group[i] = args[off]
			off++
		}
		groups[bcap-1] = group
	}
	return groups
}

func isLocalReplace(modPath string) bool {
	// ./ or ../
	if modPath == "." || modPath == ".." {
		return true
	}
	if strings.HasPrefix(modPath, "./") {
		return true
	}
	if strings.HasPrefix(modPath, "../") {
		return true
	}
	return false
}

func addBlankImports(goroot string, goBinary string, projectDir string, pkgArgs []string, test bool, mayHaveCover bool, tmpProjectDir string) (replace map[string]string, err error) {
	// list files, add init
	// NOTE: go build tag applies,
	// ignored files will be placed to IgnoredGoFiles

	// meanings:
	//   TestGoFiles       []string   // _test.go files in package
	//   XTestGoFiles      []string   // _test.go files outside package
	// XTestGoFiles are files with another package name, such as api_test for api
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

	replace = make(map[string]string)
	for _, pkg := range pkgs {
		if pkg.Standard {
			// skip std packages
			continue
		}
		type pkgInfo struct {
			imports        []string
			files          []string
			addForAllFiles bool
		}
		var pkgInfos []pkgInfo

		// no matter test or not,adding to the
		// original package is always fine.
		// it is a catch-all workaround
		// when there is no source files, then we add to all tests
		// the fact is: go allows duplicate blank imports

		var hasGoFiles bool
		if len(pkg.GoFiles) > 0 {
			hasGoFiles = true
			pkgInfos = append(pkgInfos, pkgInfo{pkg.Imports, pkg.GoFiles, false})
		}
		if !hasGoFiles || mayHaveCover {
			// mayHaveCover: see https://github.com/xhd2015/xgo/issues/285
			pkgInfos = append(pkgInfos, pkgInfo{nil, pkg.TestGoFiles, true})
			pkgInfos = append(pkgInfos, pkgInfo{nil, pkg.XTestGoFiles, true})
		}
		for _, p := range pkgInfos {
			mapping, err := addBlankImportForPackage(pkg.Dir, tmpProjectDir, p.imports, p.files, p.addForAllFiles)
			if err != nil {
				return nil, err
			}
			for srcFile, dstFile := range mapping {
				replace[srcFile] = dstFile
			}
		}
	}
	return replace, nil
}
func addBlankImportForPackage(srcDir string, dstDir string, imports []string, files []string, allFile bool) (map[string]string, error) {
	if len(files) == 0 {
		// no files
		return nil, nil
	}
	if !allFile {
		// check if already has trace
		for _, imp := range imports {
			if imp == RUNTIME_TRACE_PKG {
				return nil, nil
			}
		}
		// take the first one
		files = files[0:1]
	}

	mapping := make(map[string]string, len(files))
	for _, file := range files {
		srcFile := filepath.Join(srcDir, file)
		dstFile := filepath.Join(dstDir, srcFile)
		err := filecopy.CopyFileAll(srcFile, dstFile)
		if err != nil {
			return nil, err
		}
		// add blank import
		// NOTE: go allows duplicate blank imports
		content, err := os.ReadFile(dstFile)
		if err != nil {
			return nil, err
		}
		newContent, ok := addBlankImport(string(content))
		if !ok {
			return nil, nil
		}
		err = os.WriteFile(dstFile, []byte(newContent), 0755)
		if err != nil {
			return nil, err
		}
		mapping[srcFile] = dstFile
	}
	return mapping, nil
}
func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func isDir(path string) bool {
	stat, err := os.Stat(path)
	if err != nil {
		return false
	}
	return stat.IsDir()
}

func isFile(path string) bool {
	stat, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !stat.IsDir()
}

// see go-release/go1.22.2/src/cmd/go/internal/list/list.go
type GoListPkg struct {
	Dir        string // real dir
	ImportPath string
	Root       string // project root
	Standard   bool

	GoFiles      []string // .go source files (excluding CgoFiles, TestGoFiles, XTestGoFiles)
	TestGoFiles  []string
	XTestGoFiles []string // _test.go files outside package

	// Dependency information
	Deps         []string // all (recursively) imported dependencies
	Imports      []string // import paths used by this package
	TestImports  []string // imports from TestGoFiles
	XTestImports []string // imports from XTestGoFiles

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
	q := fmt.Sprintf(";import _ %q", RUNTIME_TRACE_PKG)
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
