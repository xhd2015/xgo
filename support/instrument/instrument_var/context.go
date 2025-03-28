package instrument_var

import (
	"fmt"
	"go/ast"
	"os"
	"path/filepath"
	"runtime"

	"github.com/xhd2015/xgo/support/instrument/edit"
)

type GlobalContext struct {
	Packages *edit.Packages
	Package  *edit.Package
	File     *edit.File
}

// a BlockContext provides a point where stmts can be prepended or inserted
type BlockContext struct {
	Global *GlobalContext
	Parent *BlockContext
	Names  map[string]bool
}

func (c *BlockContext) Add(name string) {
	if c.Names == nil {
		c.Names = make(map[string]bool, 1)
	}
	c.Names[name] = true
}

func (c *BlockContext) Has(name string) bool {
	if c == nil {
		return false
	}
	_, ok := c.Names[name]
	if ok {
		return true
	}
	return c.Parent.Has(name)
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
