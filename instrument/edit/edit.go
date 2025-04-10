package edit

import (
	"go/ast"
	"go/token"

	"github.com/xhd2015/xgo/instrument/load"
	"github.com/xhd2015/xgo/instrument/resolve/types"
	"github.com/xhd2015/xgo/support/edit/goedit"
)

type DeclKind int

const (
	DeclKindUnknown DeclKind = iota
	DeclKindVar
	DeclKindConst
	DeclKindType
	DeclKindFunc // not including methods
)

type Decl struct {
	Kind          DeclKind
	Ident         *ast.Ident
	Type          ast.Expr // might be nil
	Value         ast.Expr // only for var, might be nil
	ResolvedValue types.Object

	Decl           *ast.GenDecl
	HasCallRewrite bool // for var

	// for Func
	FuncDecl *FuncDecl

	// for Type
	Methods map[string]*FuncDecl
	File    *File
}

type FuncDecl struct {
	Name    string
	RecvPtr bool
	Syntax  *ast.FuncDecl
}

type Packages struct {
	Fset     *token.FileSet
	Packages []*Package

	PackageByPath map[string]*Package

	LoadOptions load.LoadOptions
}

type Package struct {
	LoadPackage *load.Package
	Files       []*File

	Decls map[string]*Decl

	Collected bool
}

type File struct {
	File  *load.File
	Index int

	Edit *goedit.Edit

	Decls []*Decl

	// the file index

	TrapFuncs      []*FuncInfo
	TrapVars       []*VarInfo
	InterfaceTypes []*InterfaceType
}

type FuncInfo struct {
	InfoVar      string
	FuncDecl     *ast.FuncDecl
	IdentityName string
	RecvPtr      bool
	RecvGeneric  bool
	RecvType     *ast.Ident

	Receiver *Field
	Params   Fields
	Results  Fields
}

type InterfaceType struct {
	InfoVar string
	Name    string
	Ident   *ast.Ident
	Type    *ast.InterfaceType
}

type Fields []*Field

func (f Fields) Names() []string {
	names := make([]string, len(f))
	for i, field := range f {
		names[i] = field.Name
	}
	return names
}

type VarInfo struct {
	InfoVar string
	Name    string
	Decl    *Decl
	Type    ast.Expr
}

type Field struct {
	Name      string
	NameIdent *ast.Ident // could be nil for anonymous field
	Type      ast.Expr
}

func (c *File) HasEdit() bool {
	return c.Edit != nil && c.Edit.Buffer().HasEdits()
}

func New(packages *load.Packages) *Packages {
	pkgs := &Packages{
		Fset:     packages.Fset,
		Packages: make([]*Package, 0, len(packages.Packages)),
	}
	pkgs.Add(packages)
	return pkgs
}

func (c *Packages) Add(packages *load.Packages) {
	if c.Fset != packages.Fset {
		panic("token.FileSet mismatch")
	}
	if c.PackageByPath == nil {
		c.PackageByPath = make(map[string]*Package, len(packages.Packages))
	}
	for _, pkg := range packages.Packages {
		files := make([]*File, len(pkg.Files))
		for j, file := range pkg.Files {
			files[j] = &File{
				File:  file,
				Index: j,
				Edit:  goedit.New(packages.Fset, file.Content),
			}
		}
		p := &Package{
			LoadPackage: pkg,
			Files:       files,
		}
		c.Packages = append(c.Packages, p)
		c.PackageByPath[pkg.GoPackage.ImportPath] = p
	}
}

func (p *Packages) Filter(f func(pkg *Package) bool) *Packages {
	filtered := &Packages{
		Fset:          p.Fset,
		Packages:      make([]*Package, 0, len(p.Packages)),
		PackageByPath: make(map[string]*Package, len(p.PackageByPath)),
		LoadOptions:   p.LoadOptions,
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
