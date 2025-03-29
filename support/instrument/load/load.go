package load

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"

	"github.com/xhd2015/xgo/support/goinfo"
	"github.com/xhd2015/xgo/support/instrument/overlay"
)

type LoadOptions struct {
	Dir         string
	Overlay     overlay.Overlay
	Mod         string
	IncludeTest bool
	ModFile     string // -modfile flag
}

type Package struct {
	GoPackage *goinfo.Package
	Files     []*File
}

type File struct {
	AbsPath string
	Name    string
	Content string
	Error   error
	Syntax  *ast.File
}

type Packages struct {
	Fset     *token.FileSet
	Packages []*Package
}

func LoadPackages(args []string, opts LoadOptions) (*Packages, error) {
	dir := opts.Dir
	overlayFS := opts.Overlay
	mod := opts.Mod
	modFile := opts.ModFile

	pkgs, err := goinfo.ListPackages(args, goinfo.LoadPackageOptions{
		Dir:     dir,
		Mod:     mod,
		ModFile: modFile,
	})
	if err != nil {
		return nil, err
	}

	fset := token.NewFileSet()
	loadPkgs := make([]*Package, len(pkgs))
	for i, pkg := range pkgs {
		loadPkgs[i] = &Package{
			GoPackage: pkg,
		}
	}

	// TODO: parallize
	for _, pkg := range loadPkgs {
		handleFile := func(file string) {
			absFilePath := filepath.Join(pkg.GoPackage.Dir, file)
			pkgFile := ParseFile(fset, absFilePath, overlayFS)
			pkg.Files = append(pkg.Files, pkgFile)
		}
		for _, file := range pkg.GoPackage.GoFiles {
			if !strings.HasSuffix(file, ".go") {
				continue
			}
			handleFile(file)
		}
		if opts.IncludeTest {
			for _, file := range pkg.GoPackage.TestGoFiles {
				if !strings.HasSuffix(file, ".go") {
					continue
				}
				handleFile(file)
			}
		}
	}

	return &Packages{
		Fset:     fset,
		Packages: loadPkgs,
	}, nil
}

func (c *Packages) Filter(f func(pkg *Package) bool) *Packages {
	var filtered []*Package
	for _, pkg := range c.Packages {
		if f(pkg) {
			filtered = append(filtered, pkg)
		}
	}
	return &Packages{
		Fset:     c.Fset,
		Packages: filtered,
	}
}

func ParseFile(fset *token.FileSet, asbFilePath string, overlayFS overlay.Overlay) *File {
	f := &File{
		AbsPath: asbFilePath,
		Name:    filepath.Base(asbFilePath),
	}
	_, content, err := overlayFS.Read(overlay.AbsFile(asbFilePath))
	if err != nil {
		f.Error = err
		return f
	}
	f.Content = string(content)

	file, err := parser.ParseFile(fset, string(asbFilePath), content, parser.ParseComments)
	if err != nil {
		f.Error = err
		return f
	}

	f.Syntax = file
	return f
}
