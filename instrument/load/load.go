package load

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"time"

	"github.com/xhd2015/xgo/instrument/config"
	"github.com/xhd2015/xgo/instrument/overlay"
	"github.com/xhd2015/xgo/support/goinfo"
)

type LoadOptions struct {
	Dir         string
	Overlay     overlay.Overlay
	Mod         string
	IncludeTest bool
	ModFile     string // -modfile flag
	Deps        bool   // -deps flag

	// max file size to parse
	// if file size is larger than this
	// see https://github.com/xhd2015/xgo/issues/303
	// for more background
	MaxFileSize int64

	Goroot string

	FilterErrorFile bool

	Fset *token.FileSet
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
	deps := opts.Deps
	mod := opts.Mod
	modFile := opts.ModFile
	maxFileSize := opts.MaxFileSize
	filterErrorFile := opts.FilterErrorFile
	goroot := opts.Goroot
	fset := opts.Fset

	pkgs, err := goinfo.ListPackages(args, goinfo.LoadPackageOptions{
		Dir:     dir,
		Mod:     mod,
		ModFile: modFile,
		Goroot:  goroot,
		Deps:    deps,
		Test:    false, // NOTE: don't set it
	})
	if err != nil {
		return nil, err
	}

	if fset == nil {
		fset = token.NewFileSet()
	}
	loadPkgs := make([]*Package, len(pkgs))
	for i, pkg := range pkgs {
		loadPkgs[i] = &Package{
			GoPackage: pkg,
		}
	}

	begin := time.Now()
	var numFiles int
	// TODO: parallelize
	for _, pkg := range loadPkgs {
		addFile := func(file string) {
			absFilePath := filepath.Join(pkg.GoPackage.Dir, file)
			pkgFile, err := doParseFile(fset, overlayFS, absFilePath, maxFileSize)
			numFiles++
			if err != nil {
				if filterErrorFile {
					return
				}
				pkg.Files = append(pkg.Files, &File{Error: err})
				return
			}
			if pkgFile.Error != nil && filterErrorFile {
				return
			}
			pkg.Files = append(pkg.Files, pkgFile)
		}
		for _, file := range pkg.GoPackage.GoFiles {
			if !strings.HasSuffix(file, ".go") {
				continue
			}
			addFile(file)
		}
		if opts.IncludeTest {
			for _, file := range pkg.GoPackage.TestGoFiles {
				if !strings.HasSuffix(file, ".go") {
					continue
				}
				addFile(file)
			}
			for _, file := range pkg.GoPackage.XTestGoFiles {
				if !strings.HasSuffix(file, ".go") {
					continue
				}
				addFile(file)
			}
		}
	}

	// LoadPackages: loaded 173 packages, parsed 1636 files, cost 1.054140333s
	// LoadPackages: loaded 2250 packages, parsed 14565 files, cost 4.262132333s
	config.LogDebug("LoadPackages: loaded %d packages, parsed %d files, cost %v", len(loadPkgs), numFiles, time.Since(begin))

	return &Packages{
		Fset:     fset,
		Packages: loadPkgs,
	}, nil
}

func doParseFile(fset *token.FileSet, overlayFS overlay.Overlay, absFilePath string, maxFileSize int64) (*File, error) {
	if maxFileSize > 0 {
		size, err := overlayFS.Size(overlay.AbsFile(absFilePath))
		if err != nil {
			return nil, err
		}
		if size > maxFileSize {
			return nil, fmt.Errorf("file size %d large than %d", size, maxFileSize)
		}
	}
	return parseFile(fset, absFilePath, overlayFS), nil
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

func parseFile(fset *token.FileSet, asbFilePath string, overlayFS overlay.Overlay) *File {
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
