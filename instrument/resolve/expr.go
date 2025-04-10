package resolve

import (
	"go/ast"
	"go/token"
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
		// type check X.Y to get X's type
		if isIdent(sel.X) || isSelectorExpr(sel.X) {
			isSelector = true
			c.traverseExpr(sel)
		}
		if c.needDetectMock() && len(callExpr.Args) > 0 {
			// check if mock.Patch
			if c.requireTrap(sel) {
				// resolve the first argument's type
				// to see its type so that we
				// need to insert trap points
				c.recordMockRef(callExpr.Args[0])
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
			if c.Has(idt.Name) || c.Package.Decls[idt.Name] != nil {
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
