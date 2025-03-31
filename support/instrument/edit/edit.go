package edit

import (
	"go/ast"
	"go/token"

	"github.com/xhd2015/xgo/support/edit/goedit"
	"github.com/xhd2015/xgo/support/instrument/load"
)

type DeclKind int

const (
	DeclKindUnknown DeclKind = iota
	DeclKindVar
	DeclKindConst
	DeclKindType
)

type Decl struct {
	Kind           DeclKind
	Ident          *ast.Ident
	Type           ast.Expr // might be nil
	Decl           *ast.GenDecl
	HasCallRewrite bool
}

type Packages struct {
	Fset     *token.FileSet
	Packages []*Package

	PackageByPath map[string]*Package
}

type Package struct {
	LoadPackage *load.Package
	Files       []*File

	Decls map[string]*Decl
}

type File struct {
	File *load.File
	Edit *goedit.Edit

	Decls []*Decl

	TrapFuncs []*FuncInfo
	TrapVars  []*VarInfo
}

type FuncInfo struct {
	FuncDecl *ast.FuncDecl
	Receiver *Field
	Params   []*Field
	Results  []*Field
}

type VarInfo struct {
	Name string
	Decl *Decl
	Type ast.Expr
}

type Field struct {
	Name      string
	NameIdent *ast.Ident // could be nil for anonymous field
	Type      ast.Expr
}

func (c *File) HasEdit() bool {
	return c.Edit != nil && c.Edit.Buffer().HasEdits()
}

func Edit(packages *load.Packages) *Packages {
	pkgs := &Packages{
		Fset:          packages.Fset,
		Packages:      make([]*Package, len(packages.Packages)),
		PackageByPath: make(map[string]*Package, len(packages.Packages)),
	}
	for i, pkg := range packages.Packages {
		files := make([]*File, len(pkg.Files))
		for j, file := range pkg.Files {
			files[j] = &File{
				File: file,
				Edit: goedit.New(packages.Fset, file.Content),
			}
		}
		p := &Package{
			LoadPackage: pkg,
			Files:       files,
		}
		pkgs.Packages[i] = p
		pkgs.PackageByPath[pkg.GoPackage.ImportPath] = p
	}

	return pkgs
}

func (p *Packages) Filter(f func(pkg *Package) bool) *Packages {
	filtered := &Packages{
		Fset:          p.Fset,
		PackageByPath: make(map[string]*Package),
	}
	for _, pkg := range p.Packages {
		if !f(pkg) {
			continue
		}

		filtered.Packages = append(filtered.Packages, pkg)
		filtered.PackageByPath[pkg.LoadPackage.GoPackage.ImportPath] = pkg
	}
	return filtered
}
