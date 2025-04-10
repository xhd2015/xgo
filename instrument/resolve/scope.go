package resolve

import (
	"fmt"
	"go/ast"
	"os"
	"path/filepath"
	"runtime"

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
	NamesHavingMock map[string]bool
}

func (c *NameRecorder) AddMockName(name string) {
	if c.NamesHavingMock == nil {
		c.NamesHavingMock = make(map[string]bool, 1)
	}
	c.NamesHavingMock[name] = true
}

type GlobalScope struct {
	Packages *edit.Packages
	Package  *edit.Package
	File     *edit.File

	PkgScopeNames PkgScopeNames
	Imports       Imports

	Recorder *Recorder

	detectVarTrap bool
	detectMock    bool

	// key is the expr, value is the type info
	ObjectInfo map[ast.Expr]types.Type

	NamedTypeToDecl map[types.NamedType]*edit.Decl
}

// a Scope provides a point where stmts can be prepended or inserted
type Scope struct {
	Global *GlobalScope
	Parent *Scope
	Defs   map[string]*Define
	Names  map[string]bool
}

type Define struct {
	Expr ast.Expr

	// if index==-1, it means exact match
	Index int
}

func newFileScope(global *GlobalScope) *Scope {
	return &Scope{
		Global: global,
	}
}

func (c *Scope) newScope() *Scope {
	return &Scope{
		Global: c.Global,
		Parent: c,
	}
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
	if IS_DEV {
		_, file, line, _ := runtime.Caller(1)
		fmt.Fprintf(os.Stderr, "%s:%d TODO: unknown %s: %T\n", filepath.Base(file), line, expectType, node)
	}
}
