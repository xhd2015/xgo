package goinfo

import (
	"encoding/json"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/xhd2015/xgo/support/cmd"
)

// check 'go help list'
type Package struct {
	// the abs dir
	Dir        string
	Name       string // the package name, e.g. "github.com/x/y/v2" -> y
	ImportPath string
	// the file names
	// e.g.: main.go, option.go
	GoFiles     []string
	TestGoFiles []string
	// XTestGoFiles is when a _test.go file
	// declares package as xxx_test, while the
	// package name of normal file is xxx
	XTestGoFiles []string

	// Imports are the resolved import paths of GoFiles
	Imports []string
	// imports of TestGoFiles and XTestGoFiles
	TestImports  []string
	XTestImports []string

	// Match lists the command-line patterns matching this package
	Match []string
	// ForTest is set on the variants of packages that -test lists: the
	// package whose test they are compiled for
	ForTest string

	Goroot   bool // is this package in the Go root?
	Standard bool // is this package part of the standard Go library?
	DepOnly  bool // package is only a dependency, not explicitly listed

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
	Test    bool   // -test flag, see testPackages for how its extra packages are reported
	// OverlayFile is a go -overlay JSON path. Must be passed to `go list` as
	// well as to later file reads: callers such as doctest synthesize missing
	// nested vendor go.mod files only via overlay; without -overlay here,
	// packages resolve Incomplete and instrumentation inserts no traps.
	OverlayFile string
	// GoVersion is the version of the go command, if known: from go1.19
	// only the fields of Package are listed
	GoVersion *GoVersion
}

// packageJSONFields are the fields of Package, for go list -json=fields:
// the go command then skips computing and printing the others, which are
// most of its output (embed and cgo file lists, deps, module info...).
var packageJSONFields = func() string {
	t := reflect.TypeOf(Package{})
	names := make([]string, t.NumField())
	for i := range names {
		names[i] = t.Field(i).Name
	}
	return strings.Join(names, ",")
}()

// go list -e -json ./pkg
func ListPackages(args []string, opts LoadPackageOptions) ([]*Package, error) {
	jsonFlag := "-json"
	if v := opts.GoVersion; v != nil && (v.Major > 1 || v.Major == 1 && v.Minor >= 19) {
		jsonFlag = "-json=" + packageJSONFields
	}
	flags := []string{"list", "-e", jsonFlag}
	if opts.Mod != "" {
		flags = append(flags, "-mod="+opts.Mod)
	}
	if opts.ModFile != "" {
		flags = append(flags, "-modfile="+opts.ModFile)
	}
	if opts.Deps {
		flags = append(flags, "-deps")
	}
	if opts.Test {
		flags = append(flags, "-test")
	}
	if opts.OverlayFile != "" {
		flags = append(flags, "-overlay="+opts.OverlayFile)
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
	if opts.Test {
		pkgs = testPackages(pkgs)
	}
	return pkgs, nil
}

// testPackages maps the output of `go list -test` back to plain packages.
// The test mains ("p.test") and external test packages ("p_test [p.test]")
// are dropped: the files of the latter are p's XTestGoFiles. A variant
// recompiled for a test ("q [p.test]") is dropped too when q itself is
// listed, and otherwise reported as q, which it is built from.
func testPackages(pkgs []*Package) []*Package {
	plain := make(map[string]bool, len(pkgs))
	for _, pkg := range pkgs {
		if pkg.ForTest == "" && !strings.Contains(pkg.ImportPath, " ") {
			plain[pkg.ImportPath] = true
		}
	}
	result := make([]*Package, 0, len(plain))
	for _, pkg := range pkgs {
		if pkg.Name == "main" && strings.HasSuffix(pkg.ImportPath, ".test") && plain[strings.TrimSuffix(pkg.ImportPath, ".test")] {
			continue
		}
		idx := strings.Index(pkg.ImportPath, " [")
		if idx < 0 {
			result = append(result, pkg)
			continue
		}
		base := pkg.ImportPath[:idx]
		if plain[base] || base == pkg.ForTest+"_test" {
			continue
		}
		plain[base] = true
		pkg.ImportPath = base
		pkg.ForTest = ""
		result = append(result, pkg)
	}
	return result
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
