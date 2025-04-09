package resolve

import (
	"go/ast"
	"go/token"
)

func (ctx *Scope) traverseStmt(node ast.Stmt) {
	if node == nil {
		return
	}
	switch node := node.(type) {
	case *ast.ExprStmt:
		ctx.traverseExpr(node.X)
	case *ast.EmptyStmt:
		// nothing
	case *ast.RangeStmt:
		ctx.traverseExpr(node.X)
		ctx.traverseBlockStmt(node.Body)
	case *ast.DeferStmt:
		ctx.traverseCallExpr(node.Call)
	case *ast.GoStmt:
		ctx.traverseCallExpr(node.Call)
	case *ast.AssignStmt:
		for _, rhs := range node.Rhs {
			ctx.traverseExpr(rhs)
		}
		if node.Tok == token.DEFINE {
			leftTotal := len(node.Lhs)
			rightTotal := len(node.Rhs)
			// add name to current scope after RHS
			// either:
			//    a := Result()
			// or:
			//    a,b := Result()
			for i, lhs := range node.Lhs {
				if name, ok := lhs.(*ast.Ident); ok {
					if !isBlankName(name.Name) {
						var rhs ast.Expr
						defIdx := i
						if leftTotal == rightTotal {
							defIdx = -1
							rhs = node.Rhs[i]
						} else {
							defIdx = i
							rhs = node.Rhs[0]
						}
						ctx.AddDef(name.Name, &Define{
							Expr:  rhs,
							Index: defIdx,
						})
					}
				}
			}
		}
	case *ast.TypeSwitchStmt:
		ctx.traverseStmt(node.Init)
		ctx.traverseStmt(node.Assign)
		ctx.traverseBlockStmt(node.Body)
	case *ast.BadStmt:
		// nothing
	case *ast.IfStmt:
		ctx.traverseStmt(node.Init)
		ctx.traverseExpr(node.Cond)
		// node.Then =
		ctx.traverseBlockStmt(node.Body)
		ctx.traverseStmt(node.Else)
	case *ast.ForStmt:
		ctx.traverseStmt(node.Init)
		ctx.traverseExpr(node.Cond)
		ctx.traverseStmt(node.Post)
		ctx.traverseBlockStmt(node.Body)
	case *ast.SwitchStmt:
		ctx.traverseStmt(node.Init)
		ctx.traverseExpr(node.Tag)
		ctx.traverseBlockStmt(node.Body)
	case *ast.SelectStmt:
		ctx.traverseBlockStmt(node.Body)
	case *ast.DeclStmt:
		ctx.traverseDecl(node.Decl)
	case *ast.LabeledStmt:
		ctx.traverseStmt(node.Stmt)
	case *ast.BranchStmt:
		// ignore continue or continue label
	case *ast.ReturnStmt:
		ctx.traverseExprs(node.Results)
	case *ast.BlockStmt:
		ctx.traverseBlockStmt(node)
	case *ast.IncDecStmt:
		// ignore
	case *ast.CaseClause:
		ctx.traverseBlockStmt(&ast.BlockStmt{List: node.Body})
	case *ast.CommClause:
		// select {...}
		// TODO: handle name introduced in case
		ctx.traverseBlockStmt(&ast.BlockStmt{List: node.Body})
	case *ast.SendStmt:
		// ignore
		ctx.traverseExpr(node.Chan)
		ctx.traverseExpr(node.Value)
	default:
		errorUnknown("stmt", node)
	}
}
