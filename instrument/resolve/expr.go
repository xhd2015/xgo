package resolve

import (
	"go/ast"
	"go/token"

	"github.com/xhd2015/xgo/instrument/constants"
)

func (ctx *Scope) traverseExprs(nodes []ast.Expr) {
	for _, node := range nodes {
		ctx.traverseExpr(node)
	}
}

func (c *Scope) traverseCallExpr(callExpr *ast.CallExpr) {
	if callExpr == nil {
		return
	}
	// pkg.A.Method or A.Method
	// we don't know A's type, it could be value type or pointer type
	// we can enhance this in the future by getting type info
	var isSelector bool
	fn := callExpr.Fun
	if sel, ok := fn.(*ast.SelectorExpr); ok {
		if isIdent(sel.X) || isSelectorExpr(sel.X) {
			isSelector = true
		}
		if c.needDetectMock() && len(callExpr.Args) > 0 {
			// check if mock.Patch
			if c.isMockPatch(sel) {
				// resolve the first argument's type
				// to see its type so that we
				// need to insert trap points
				c.resolveMockRef(callExpr.Args[0])
			}
		}
	}
	if !isSelector {
		c.traverseExpr(fn)
	}

	for _, arg := range callExpr.Args {
		c.traverseExpr(arg)
	}
}

func isIdent(expr ast.Expr) bool {
	_, ok := expr.(*ast.Ident)
	return ok
}
func isSelectorExpr(expr ast.Expr) bool {
	_, ok := expr.(*ast.SelectorExpr)
	return ok
}

func (c *Scope) traverseExpr(node ast.Expr) {
	if node == nil {
		return
	}

	switch node := node.(type) {
	case *ast.Ident:
		if c.trapIdent(nil, node) {
			return
		}
	case *ast.SelectorExpr:
		if c.trapSelector(nil, node) {
			return
		}
		// var lock sync.Mutex
		// lock.Lock --> lock should not be intercepted
		// because its type cannot be determined
		// see TestMockLockShouldNotCopy
		_, ok := deparen(node.X).(*ast.Ident)
		if !ok {
			c.traverseExpr(node.X)
		}
	case *ast.UnaryExpr:
		// take addr
		if node.Op == token.AND {
			switch x := node.X.(type) {
			case *ast.SelectorExpr:
				if c.trapSelector(node, x) {
					return
				}
			case *ast.Ident:
				if c.trapIdent(node, x) {
					return
				}
			}
		}
		c.traverseExpr(node.X)
	case *ast.CallExpr:
		c.traverseCallExpr(node)
	case *ast.BinaryExpr:
		c.traverseExpr(node.X)
		c.traverseExpr(node.Y)
	case *ast.ParenExpr:
		c.traverseExpr(node.X)
	case *ast.BasicLit:
		// nothing
	case *ast.FuncLit:
		c.traverseFuncLit(node)
	case *ast.StarExpr:
		// * something
		c.traverseExpr(node.X)
	case *ast.TypeAssertExpr:
		// something.(type)
		c.traverseExpr(node.X)
	case *ast.CompositeLit:
		// something{...}
		for _, elem := range node.Elts {
			c.traverseExpr(elem)
		}
	case *ast.IndexExpr:
		// something[...], only traverse the index
		c.traverseExpr(node.X)
		c.traverseExpr(node.Index)
	case *ast.SliceExpr:
		// something[...:...], only traverse the index
		c.traverseExpr(node.X)
		c.traverseExpr(node.Low)
		c.traverseExpr(node.High)
	case *ast.ChanType:
		// chan something
	case *ast.MapType:
		// map[key]value

	case *ast.StructType:
		// struct{...}
	case *ast.ArrayType:
		// [...][...]
	case *ast.InterfaceType:
		// interface{...}
	case *ast.KeyValueExpr:
		// key: value
		// map[string]int{ SomeVar: 10 }
		// v.s.
		// SomeStruct{ SomeVar: 10}
		// we cannot know if SomeVar is a field or a variable
		// if there is an identical variable, we will not intercept it
		// see test: ./runtime/test/mock/mock_var/same_field
		var keyMightConflict bool
		if idt, ok := node.Key.(*ast.Ident); ok {
			if c.Has(idt.Name) || c.Global.PkgScopeNames[idt.Name] != nil {
				keyMightConflict = true
			}
		}
		if !keyMightConflict {
			c.traverseExpr(node.Key)
		}
		c.traverseExpr(node.Value)
	default:
		errorUnknown("expr", node)
	}
}

func deparen(expr ast.Expr) ast.Expr {
	if expr == nil {
		return nil
	}
	for {
		p, ok := expr.(*ast.ParenExpr)
		if !ok {
			return expr
		}
		expr = p.X
	}
}

func (c *Scope) isMockPatch(sel *ast.SelectorExpr) bool {
	pkgPath, name, ok := c.tryAsPkgRef(sel)
	if !ok {
		return false
	}
	return pkgPath == constants.RUNTIME_MOCK_PKG && name == "Patch"
}

func (c *Scope) tryAsPkgRef(sel *ast.SelectorExpr) (pkgPath string, name string, ok bool) {
	idt, ok := sel.X.(*ast.Ident)
	if !ok {
		return "", "", false
	}
	imp, ok := c.Global.Imports[idt.Name]
	if !ok {
		return "", "", false
	}
	return imp, sel.Sel.Name, true
}

// `mock.Patch(fn,...)`
// fn examples:
//
//	(*ipc.Reader).Next
//	vr.Reader.Next where vr.Reader is an embedded field: struct { *ipc.Reader }
func (c *Scope) resolveMockRef(fn ast.Expr) {
	switch node := fn.(type) {
	case *ast.Ident:
		def := c.GetDef(node.Name)
		if def == nil {
			return
		}
		// def example:
		//   &flight.Writer{}
		if def.Index == -1 {
			// TODO
		}
	case *ast.SelectorExpr:
		name := node.Sel.Name
		if isBlankName(name) {
			return
		}
		if idt, ok := node.X.(*ast.Ident); ok {
			pkgPath, ok := c.Global.Imports[idt.Name]
			if ok {
				recorder := c.Global.Recorder
				recorder.GetOrInit(pkgPath).GetOrInit(name).HasMockPatch = true
				return
			}
			def := c.GetDef(idt.Name)
			if def != nil {
				// var.Name
				if def.Index == -1 {
					pkgPath, typeName, ok := c.tryCompositeLit(def.Expr)
					if ok {
						recorder := c.Global.Recorder
						recorder.GetOrInit(pkgPath).GetOrInit(typeName).AddMockName(name)
					}
				}
			}
			return
		}
		if pkgPath, typeName, ok := c.tryPkgType(node.X); ok {
			recorder := c.Global.Recorder
			recorder.GetOrInit(pkgPath).GetOrInit(typeName).AddMockName(name)
			return
		}
	}
}

// &flight.Writer{}
func (c *Scope) tryCompositeLit(expr ast.Expr) (pkgPath string, name string, ok bool) {
	unary, ok := expr.(*ast.UnaryExpr)
	if ok {
		if unary.Op != token.AND {
			return "", "", false
		}
		expr = unary.X
	}
	lit, ok := expr.(*ast.CompositeLit)
	if !ok {
		return "", "", false
	}
	sel, ok := lit.Type.(*ast.SelectorExpr)
	if !ok {
		return "", "", false
	}
	return c.tryAsPkgRef(sel)
}

// (*ipc.Reader).Next or ipc.Reader.Next
func (c *Scope) tryPkgType(expr ast.Expr) (pkgPath string, name string, ok bool) {
	paren, ok := expr.(*ast.ParenExpr)
	if ok {
		starExpr, ok := paren.X.(*ast.StarExpr)
		if !ok {
			return "", "", false
		}
		expr = starExpr.X
	}
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return "", "", false
	}
	return c.tryAsPkgRef(sel)
}
