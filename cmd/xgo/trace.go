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
)

const RUNTIME_MODULE = "github.com/xhd2015/xgo/runtime"
const TRACE_PKG = RUNTIME_MODULE + "/trace"

type importResult struct {
	overlayFile string
}

//go:embed runtime_gen
var runtimeGenFS embed.FS

// TODO: may apply tags
func importRuntimeDep(test bool, goroot string, goBinary string, xgoSrc string, projectDir string, modRootRel []string, mainModule string, args []string) (*importResult, error) {
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
	// assert no vendor
	vendorDir := filepath.Join(projectRoot, "vendor")
	_, statErr := os.Stat(vendorDir)
	if statErr != nil {
		if !errors.Is(statErr, os.ErrNotExist) {
			return nil, statErr
		}
	} else {
		// no vendor support, suppress
		return nil, nil
	}

	overlayFile, err := createOverlay(goroot, goBinary, xgoSrc, test, projectRoot, projectDir, pkgArgs)
	if err != nil {
		return nil, err
	}

	return &importResult{
		overlayFile: overlayFile,
	}, nil
}

func getPkgArgs(args []string) []string {
	n := len(args)
	newArgs := make([]string, 0, n)
	for i := 0; i < n; i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "-") {
			// stop at first non-arg
			newArgs = append(newArgs, args[i:]...)
			break
		}
		if arg == "-args" {
			// go test -args: pass everything after to underlying program
			break
		}
		eqIdx := strings.Index(arg, "=")
		if eqIdx >= 0 {
			// self hosted arg
			continue
		}
		// make --opt equivalent with -opt
		if strings.HasPrefix(arg, "--") {
			arg = arg[1:]
		}
		switch arg {
		case "-a", "-n", "-race", "-masan", "-asan", "-cover", "-v", "-work", "-x", "-linkshared", "-buildvcs", // shared among build,test,run
			"-args", "-c", "-json": // -json for test
			// zero arg
		default:
			// 1 arg
			i++
		}
	}
	return newArgs
}

type Overlay struct {
	Replace map[string]string
}

func createOverlay(goroot string, goBinary string, xgoSrc string, test bool, projectRoot string, projectDir string, pkgArgs []string) (string, error) {
	// try /tmp first
	tmpDir := "/tmp"
	_, statErr := os.Stat(tmpDir)
	if statErr != nil {
		tmpDir = os.TempDir()
	}

	xgoTmp := filepath.Join(tmpDir, "xgo_"+fileutil.CleanSpecial(getRevision()))
	err := os.MkdirAll(xgoTmp, 0755)
	if err != nil {
		return "", err
	}
	logDebug("xgo tmp dir: %s", xgoTmp)

	suffix := ""
	if isDevelopment {
		suffix = "_dev"
	}
	tmpRuntime := filepath.Join(xgoTmp, "runtime"+suffix)
	err = os.MkdirAll(tmpRuntime, 0755)
	if err != nil {
		return "", err
	}
	if isDevelopment {
		err = filecopy.CopyReplaceDir(filepath.Join(xgoSrc, "runtime"), tmpRuntime, false)
		if err != nil {
			return "", err
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
				return "", err
			}
			err = copyEmbedDir(runtimeGenFS, "runtime_gen", tmpRuntime)
			if err != nil {
				return "", err
			}
			err = os.Rename(filepath.Join(tmpRuntime, "go.mod.txt"), filepath.Join(tmpRuntime, "go.mod"))
			if err != nil {
				return "", err
			}
		}
	}

	// create project
	projectSum, err := pathsum.PathSum("", projectRoot)
	if err != nil {
		return "", err
	}
	tmpProjectDir := filepath.Join(xgoTmp, "projects", projectSum)
	logDebug("tmp project dir: %s", tmpProjectDir)
	err = os.MkdirAll(tmpProjectDir, 0755)
	if err != nil {
		return "", err
	}

	goMod := filepath.Join(projectRoot, "go.mod")
	tmpGoMod := filepath.Join(tmpProjectDir, "go.mod")
	err = filecopy.CopyFile(goMod, tmpGoMod)
	if err != nil {
		return "", err
	}

	err = cmd.Env([]string{
		"GOROOT=" + goroot,
	}).Run(goBinary, "mod", "edit",
		fmt.Sprintf("-require=%s@v%s", RUNTIME_MODULE, VERSION),
		fmt.Sprintf("-replace=%s=%s", RUNTIME_MODULE, tmpRuntime),
		tmpGoMod,
	)
	if err != nil {
		return "", err
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
		return "", err
	}
	var pkgs []*GoListPkg
	dec := json.NewDecoder(strings.NewReader(output))
	for dec.More() {
		var pkg *GoListPkg
		err := dec.Decode(&pkg)
		if err != nil {
			return "", err
		}
		pkgs = append(pkgs, pkg)
	}

	replace := make(map[string]string)
	for _, pkg := range pkgs {
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
			return "", err
		}
		// add blank import
		content, err := os.ReadFile(dstFile)
		if err != nil {
			return "", err
		}
		newContent, ok := addBlankImport(string(content))
		if !ok {
			continue
		}
		err = os.WriteFile(dstFile, []byte(newContent), 0755)
		if err != nil {
			return "", err
		}
		replace[srcFile] = dstFile
	}

	replace[goMod] = tmpGoMod

	overlay := Overlay{Replace: replace}
	overlayData, err := json.Marshal(overlay)
	if err != nil {
		return "", err
	}
	overlayFile := filepath.Join(tmpProjectDir, "overlay.json")
	err = os.WriteFile(overlayFile, overlayData, 0755)
	if err != nil {
		return "", err
	}
	return overlayFile, nil
}

type GoListPkg struct {
	Root        string
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
