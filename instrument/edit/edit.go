package edit

import (
	"fmt"
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
	Kind  DeclKind
	Ident *ast.Ident
	Type  ast.Expr // might be nil

	Value             ast.Expr // only for var, might be nil
	ResolvedValueType types.Type

	Decl *ast.GenDecl

	// Pending var refs for rewrite
	// for var only
	VarRefs []*VarRef

	// for Func
	FuncDecl *FuncDecl

	// for Type
	Methods map[string]*FuncDecl
	File    *File
}

type VarRef struct {
	File *File
	// the prefiex & token, if any
	Addr *ast.UnaryExpr
	// call get() or get_addr()
	NeedPtr   bool
	NameStart token.Pos
	// the end position of the name
	NameEnd token.Pos
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

	// Main indicates whether
	// the package is within main module
	// NOTE: this does not consider go.mod
	// boundaries, it only consider the
	// prefix of the package path.
	Main bool

	// Initial indicates whether
	// the packages are loaded via
	// package args specified by user
	Initial bool

	// Xgo indicates whether
	// the package is xgo/runtime
	Xgo bool

	// AllowInstrument indicates whether
	// the package is allowed to be instrumented
	AllowInstrument bool
}

type File struct {
	File *load.File
	// the file index
	Index int

	Edit *goedit.Edit

	Decls []*Decl

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
		// add flag
		InitPkgFlag(p)
		c.Packages = append(c.Packages, p)
		c.PackageByPath[pkg.GoPackage.ImportPath] = p
	}
}

func (p *Packages) Filter(f func(pkg *Package) bool) []*Package {
	filtered := make([]*Package, 0, len(p.Packages))
	for _, pkg := range p.Packages {
		if !f(pkg) {
			continue
		}
		filtered = append(filtered, pkg)
	}
	return filtered
}

func (p *Packages) CloneWithPackages(packages []*Package) *Packages {
	pkgMap := make(map[string]*Package, len(packages))
	for _, pkg := range packages {
		pkgMap[pkg.LoadPackage.GoPackage.ImportPath] = pkg
	}
	return &Packages{
		Fset:          p.Fset,
		Packages:      packages,
		PackageByPath: pkgMap,
		LoadOptions:   p.LoadOptions,
	}
}

func (c *Packages) Merge(pkgs *Packages, override bool) {
	if c.Fset != pkgs.Fset {
		panic("token.FileSet mismatch")
	}
	if len(pkgs.Packages) == 0 {
		return
	}
	if c.PackageByPath == nil {
		c.PackageByPath = make(map[string]*Package, len(pkgs.Packages))
	}
	for _, pkg := range pkgs.Packages {
		pkgPath := pkg.LoadPackage.GoPackage.ImportPath
		_, ok := c.PackageByPath[pkgPath]
		if ok {
			if !override {
				panic(fmt.Errorf("package %s already exists", pkgPath))
			}
			// replace
			n := len(c.Packages)
			for i := 0; i < n; i++ {
				if c.Packages[i].LoadPackage.GoPackage.ImportPath == pkgPath {
					c.Packages[i] = pkg
					break
				}
			}
		} else {
			c.Packages = append(c.Packages, pkg)
		}
		c.PackageByPath[pkgPath] = pkg
	}
}
