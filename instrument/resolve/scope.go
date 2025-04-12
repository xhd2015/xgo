package resolve

import (
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"runtime"

	"github.com/xhd2015/xgo/instrument/config"
	"github.com/xhd2015/xgo/instrument/edit"
	"github.com/xhd2015/xgo/instrument/resolve/types"
)

// imports: key is local name, value is import path
type PkgScopeNames map[string]*edit.Decl

type Recorder struct {
	Pkgs map[string]*PkgRecorder
}

func (c *Recorder) GetOrInit(pkgPath string) *PkgRecorder {
	if c.Pkgs == nil {
		c.Pkgs = make(map[string]*PkgRecorder, 1)
	}
	pkg, ok := c.Pkgs[pkgPath]
	if ok {
		return pkg
	}
	pkg = &PkgRecorder{}
	c.Pkgs[pkgPath] = pkg
	return pkg
}

type PkgRecorder struct {
	Names map[string]*NameRecorder
}

func (c *PkgRecorder) GetOrInit(name string) *NameRecorder {
	if c.Names == nil {
		c.Names = make(map[string]*NameRecorder, 1)
	}
	rec, ok := c.Names[name]
	if ok {
		return rec
	}
	rec = &NameRecorder{}
	c.Names[name] = rec
	return rec
}

func (c *PkgRecorder) Get(name string) *NameRecorder {
	return c.Names[name]
}

type NameRecorder struct {
	HasMockRef      bool
	HasVarTrap      bool
	NamesHavingMock map[string]bool
}

func (c *NameRecorder) AddMockName(name string) {
	if c.NamesHavingMock == nil {
		c.NamesHavingMock = make(map[string]bool, 1)
	}
	c.NamesHavingMock[name] = true
}

type PackageRegistry interface {
	Fset() *token.FileSet
	LoadPackage(pkgPath string) (*edit.Package, bool, error)
	GetPackage(pkgPath string) *edit.Package
}

type GlobalScope struct {
	Packages PackageRegistry
	Recorder *Recorder

	detectVarTrap bool
	detectMock    bool

	// key is the expr, value is the type info
	ExprInfo map[ast.Expr]types.Info

	NamedTypeToDecl map[PkgName]*edit.Decl

	cachedFileScopes map[*edit.File]*Scope
}

type PackageScope struct {
	Package *edit.Package
	Decls   PkgScopeNames
}

func (c PackageScope) PkgPath() string {
	return c.Package.LoadPackage.GoPackage.ImportPath
}

type FileScope struct {
	File    *edit.File
	Imports Imports
}

type PkgName struct {
	PkgPath string
	Name    string
}

// a Scope provides a point where stmts can be prepended or inserted
type Scope struct {
	Global  *GlobalScope
	Package *PackageScope
	File    *FileScope

	Parent *Scope
	Defs   map[string]*Define
	Names  map[string]bool
}

type Define struct {
	// is this Def causes a split?
	// if so, should look for parent scope
	Split bool
	Expr  ast.Expr

	// if index==-1, it means exact match
	Index int
}

func newFileScope(global *GlobalScope, pkg *edit.Package, file *edit.File) *Scope {
	scope := global.cachedFileScopes[file]
	if scope != nil {
		return scope
	}
	imports := getFileImports(file.File.Syntax)
	scope = &Scope{
		Global: global,
		Package: &PackageScope{
			Package: pkg,
			Decls:   pkg.Decls,
		},
		File: &FileScope{
			File:    file,
			Imports: imports,
		},
	}
	if global.cachedFileScopes == nil {
		global.cachedFileScopes = make(map[*edit.File]*Scope, 1)
	}
	global.cachedFileScopes[file] = scope
	return scope
}

func (c *Scope) newScope() *Scope {
	return &Scope{
		Global:  c.Global,
		Package: c.Package,
		File:    c.File,
		Parent:  c,
	}
}

// tt:=tt --> creates a new scope in later phase
func (c *Scope) splitScopeWithDef(defs map[string]*Define, names map[string]bool) {
	clone := *c

	c.Parent = &clone
	c.Defs = defs
	c.Names = names
}

func (c *Scope) Add(name string) {
	if c.Names == nil {
		c.Names = make(map[string]bool, 1)
	}
	c.Names[name] = true
}

func (c *Scope) AddDef(name string, def *Define) {
	if c.Defs == nil {
		c.Defs = make(map[string]*Define, 1)
	}
	c.Defs[name] = def
}

// Has checks if the name is defined in local scope
func (c *Scope) Has(name string) bool {
	_, ok := c.Names[name]
	if ok {
		return true
	}
	_, ok = c.Defs[name]
	if ok {
		return true
	}
	return c.Parent != nil && c.Parent.Has(name)
}

func (c *Scope) GetDef(name string) *Define {
	def, ok := c.Defs[name]
	if ok {
		return def
	}
	if c.Parent != nil {
		return c.Parent.GetDef(name)
	}
	return nil
}

func (c *Scope) needDetectVarTrap() bool {
	return c.Global.detectVarTrap
}

func (c *Scope) needDetectMock() bool {
	return c.Global.detectMock
}

func (c *Scope) needRecordDef() bool {
	// need to record ident def?
	//   a:=SomeExpr()
	return c.needDetectMock()
}

func errorUnknown(expectType string, node ast.Node) {
	// unknown
	if os.Getenv("XGO_DEBUG_VAR_TRAP_STRICT") == "true" {
		panic(fmt.Errorf("unrecognized %s: %T", expectType, node))
	}
	if config.IS_DEV {
		_, file, line, _ := runtime.Caller(1)
		fmt.Fprintf(os.Stderr, "%s:%d TODO: unknown %s: %T\n", filepath.Base(file), line, expectType, node)
	}
}
