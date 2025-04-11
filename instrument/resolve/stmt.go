package resolve

import (
	"go/ast"
	"go/token"
)

func (c *Scope) traverseStmt(node ast.Stmt) {
	if node == nil {
		return
	}
	switch node := node.(type) {
	case *ast.ExprStmt:
		c.traverseExpr(node.X)
	case *ast.EmptyStmt:
		// nothing
	case *ast.RangeStmt:
		if node.Tok == token.DEFINE {
			if key, ok := node.Key.(*ast.Ident); ok && !isBlankName(key.Name) {
				c.Add(key.Name)
			}
			if value, ok := node.Value.(*ast.Ident); ok && !isBlankName(value.Name) {
				c.Add(value.Name)
			}
		}
		c.traverseExpr(node.X)
		c.traverseBlockStmt(node.Body)
	case *ast.DeferStmt:
		c.traverseCallExpr(node.Call)
	case *ast.GoStmt:
		c.traverseCallExpr(node.Call)
	case *ast.AssignStmt:
		for _, rhs := range node.Rhs {
			c.traverseExpr(rhs)
		}

		// edge case:
		// tt:=tt --> before this line, tt is another define
		if node.Tok == token.DEFINE {
			leftTotal := len(node.Lhs)
			rightTotal := len(node.Rhs)
			// add name to current scope after RHS
			// either:
			//    a := Result()
			// or:
			//    a,b := Result()
			var rhsNames []string
			getRhsNames := func() []string {
				if rhsNames != nil {
					return rhsNames
				}
				rhsNames = make([]string, 0, rightTotal)
				for _, rhs := range node.Rhs {
					if name, ok := rhs.(*ast.Ident); ok {
						if isBlankName(name.Name) {
							continue
						}
						rhsNames = append(rhsNames, name.Name)
					}
				}
				return rhsNames
			}

			// has name conflicts
			var hasNameConflicts bool
			var defs map[string]*Define
			for i, lhs := range node.Lhs {
				if name, ok := lhs.(*ast.Ident); ok {
					if isBlankName(name.Name) {
						continue
					}
					if !hasNameConflicts && inList(name.Name, getRhsNames()) {
						hasNameConflicts = true
					}
					var rhs ast.Expr
					defIdx := i
					if leftTotal == rightTotal {
						defIdx = -1
						rhs = node.Rhs[i]
					} else {
						defIdx = i
						rhs = node.Rhs[0]
					}
					if defs == nil {
						defs = make(map[string]*Define)
					}
					defs[name.Name] = &Define{
						Expr:  rhs,
						Index: defIdx,
					}
				}
			}

			// should start a new scope
			if !hasNameConflicts {
				for name, def := range defs {
					c.AddDef(name, def)
				}
			} else {
				for _, def := range defs {
					def.Split = true
				}
				c.splitScopeWithDef(defs, nil)
			}
		}
	case *ast.TypeSwitchStmt:
		c.traverseStmt(node.Init)
		c.traverseStmt(node.Assign)
		c.traverseBlockStmt(node.Body)
	case *ast.BadStmt:
		// nothing
	case *ast.IfStmt:
		c.traverseStmt(node.Init)
		c.traverseExpr(node.Cond)
		// node.Then =
		c.traverseBlockStmt(node.Body)
		c.traverseStmt(node.Else)
	case *ast.ForStmt:
		c.traverseStmt(node.Init)
		c.traverseExpr(node.Cond)
		c.traverseStmt(node.Post)
		c.traverseBlockStmt(node.Body)
	case *ast.SwitchStmt:
		c.traverseStmt(node.Init)
		c.traverseExpr(node.Tag)
		c.traverseBlockStmt(node.Body)
	case *ast.SelectStmt:
		c.traverseBlockStmt(node.Body)
	case *ast.DeclStmt:
		c.traverseDecl(node.Decl)
	case *ast.LabeledStmt:
		c.traverseStmt(node.Stmt)
	case *ast.BranchStmt:
		// ignore continue or continue label
	case *ast.ReturnStmt:
		c.traverseExprs(node.Results)
	case *ast.BlockStmt:
		c.traverseBlockStmt(node)
	case *ast.IncDecStmt:
		// ignore
	case *ast.CaseClause:
		c.traverseBlockStmt(&ast.BlockStmt{List: node.Body})
	case *ast.CommClause:
		// select {...}
		// TODO: handle name introduced in case
		c.traverseBlockStmt(&ast.BlockStmt{List: node.Body})
	case *ast.SendStmt:
		// ignore
		c.traverseExpr(node.Chan)
		c.traverseExpr(node.Value)
	default:
		errorUnknown("stmt", node)
	}
}

func inList(name string, list []string) bool {
	for _, n := range list {
		if n == name {
			return true
		}
	}
	return false
}
