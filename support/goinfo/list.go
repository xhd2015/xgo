package goinfo

import (
	"encoding/json"
	"path/filepath"
	"strings"

	"github.com/xhd2015/xgo/support/cmd"
)

// check 'go help list'
type Package struct {
	// the abs dir
	Dir        string
	Name       string
	ImportPath string
	// the file names
	// e.g.: main.go, option.go
	GoFiles     []string
	TestGoFiles []string
	// XTestGoFiles is when a _test.go file
	// declares package as xxx_test, while the
	// package name of normal file is xxx
	XTestGoFiles []string

	Goroot   bool // is this package in the Go root?
	Standard bool // is this package part of the standard Go library?

	Incomplete bool          // this package or a dependency has an error
	Error      *PackageError // error loading package
}

type PackageError struct {
	ImportStack []string // shortest path from package named on command line to this one
	Pos         string   // position of error (if present, file:line:col)
	Err         string   // the error itself
}

type LoadPackageOptions struct {
	Dir     string
	Mod     string
	ModFile string // -modfile flag
	Goroot  string // GOROOT env
	Deps    bool   // -deps flag
}

// go list -e -json ./pkg
func ListPackages(args []string, opts LoadPackageOptions) ([]*Package, error) {
	flags := []string{"list", "-e", "-json"}
	if opts.Mod != "" {
		flags = append(flags, "-mod="+opts.Mod)
	}
	if opts.ModFile != "" {
		flags = append(flags, "-modfile="+opts.ModFile)
	}
	if opts.Deps {
		flags = append(flags, "-deps")
	}
	flags = append(flags, args...)
	var env []string
	goCmd := "go"
	if opts.Goroot != "" {
		goCmd = filepath.Join(opts.Goroot, "bin", "go")
		env = append(env, "GOROOT="+opts.Goroot)
	}
	output, err := cmd.Dir(opts.Dir).Env(env).Output(goCmd, flags...)
	if err != nil {
		return nil, err
	}
	var pkgs []*Package
	dec := json.NewDecoder(strings.NewReader(output))
	for dec.More() {
		var pkg Package
		err := dec.Decode(&pkg)
		if err != nil {
			return nil, err
		}
		pkgs = append(pkgs, &pkg)
	}
	return pkgs, nil
}

// go list ./pkg
func ListPackagePaths(dir string, mod string, args []string) ([]string, error) {
	flags := []string{"list"}
	if mod != "" {
		flags = append(flags, "-mod="+mod)
	}
	flags = append(flags, args...)
	output, err := cmd.Dir(dir).Output("go", flags...)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(output, "\n")
	return lines, nil
}

// ListFiles list all go files, return a list of absolute paths
// go list -e -json ./pkg
func ListFiles(dir string, args []string) ([]string, error) {
	return lisFiles(dir, false, args)
}

// ListRelativeFiles list all go files, return a list of relative paths
// go list -e -json ./pkg
func ListRelativeFiles(dir string, args []string) ([]string, error) {
	return lisFiles(dir, true, args)
}

func lisFiles(dir string, relativeOnly bool, args []string) ([]string, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	flags := []string{"list", "-e", "-json"}
	flags = append(flags, args...)
	res, err := cmd.Dir(dir).Output("go", flags...)
	if err != nil {
		return nil, err
	}
	var resultFiles []string
	dec := json.NewDecoder(strings.NewReader(res))
	for dec.More() {
		// check 'go help list'
		var pkg struct {
			// the abs dir
			Dir        string
			ImportPath string
			// the file names
			GoFiles     []string
			TestGoFiles []string
		}
		err := dec.Decode(&pkg)
		if err != nil {
			return nil, err
		}
		goFiles := make([]string, 0, len(pkg.GoFiles)+len(pkg.TestGoFiles))
		goFiles = append(goFiles, pkg.GoFiles...)
		goFiles = append(goFiles, pkg.TestGoFiles...)
		if len(goFiles) > 0 {
			absPkgDir, err := filepath.Abs(pkg.Dir)
			if err != nil {
				return nil, err
			}
			for _, goFile := range goFiles {
				if !strings.HasSuffix(goFile, ".go") {
					// some cache files
					continue
				}
				absGoFile := goFile
				if !filepath.IsAbs(goFile) {
					// go list outputs file names
					// in most cases this goFile is not absolute
					absGoFile = filepath.Join(absPkgDir, goFile)
				}
				if relativeOnly {
					// this is just for compatibility
					if !strings.HasPrefix(absGoFile, absDir) {
						continue
					}
					relFile := strings.TrimPrefix(absGoFile[len(absDir):], string(filepath.Separator))
					if relFile != "" {
						resultFiles = append(resultFiles, relFile)
					}
					continue
				}
				resultFiles = append(resultFiles, absGoFile)
			}
		}
	}
	return resultFiles, nil
}
