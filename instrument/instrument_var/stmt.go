package instrument_var

import (
	"go/ast"
	"go/token"
)

func (ctx *BlockContext) traverseStmt(node ast.Stmt, pkgScopeNames PkgScopeNames, imports Imports) {
	if node == nil {
		return
	}
	switch node := node.(type) {
	case *ast.ExprStmt:
		ctx.traverseExpr(node.X, pkgScopeNames, imports)
	case *ast.EmptyStmt:
		// nothing
	case *ast.RangeStmt:
		ctx.traverseExpr(node.X, pkgScopeNames, imports)
		ctx.traverseBlockStmt(node.Body, pkgScopeNames, imports)
	case *ast.DeferStmt:
		ctx.traverseCallExpr(node.Call, pkgScopeNames, imports)
	case *ast.GoStmt:
		ctx.traverseCallExpr(node.Call, pkgScopeNames, imports)
	case *ast.AssignStmt:
		if node.Tok == token.DEFINE {
			// add name to current scope
			for _, lhs := range node.Lhs {
				if name, ok := lhs.(*ast.Ident); ok {
					if !isBlankName(name.Name) {
						ctx.Add(name.Name)
					}
				}
			}
		}
		for _, rhs := range node.Rhs {
			ctx.traverseExpr(rhs, pkgScopeNames, imports)
		}
	case *ast.TypeSwitchStmt:
		ctx.traverseStmt(node.Init, pkgScopeNames, imports)
		ctx.traverseStmt(node.Assign, pkgScopeNames, imports)
		ctx.traverseBlockStmt(node.Body, pkgScopeNames, imports)
	case *ast.BadStmt:
		// nothing
	case *ast.IfStmt:
		ctx.traverseStmt(node.Init, pkgScopeNames, imports)
		ctx.traverseExpr(node.Cond, pkgScopeNames, imports)
		// node.Then =
		ctx.traverseBlockStmt(node.Body, pkgScopeNames, imports)
		ctx.traverseStmt(node.Else, pkgScopeNames, imports)
	case *ast.ForStmt:
		ctx.traverseStmt(node.Init, pkgScopeNames, imports)
		ctx.traverseExpr(node.Cond, pkgScopeNames, imports)
		ctx.traverseStmt(node.Post, pkgScopeNames, imports)
		ctx.traverseBlockStmt(node.Body, pkgScopeNames, imports)
	case *ast.SwitchStmt:
		ctx.traverseStmt(node.Init, pkgScopeNames, imports)
		ctx.traverseExpr(node.Tag, pkgScopeNames, imports)
		ctx.traverseBlockStmt(node.Body, pkgScopeNames, imports)
	case *ast.SelectStmt:
		ctx.traverseBlockStmt(node.Body, pkgScopeNames, imports)
	case *ast.DeclStmt:
		ctx.traverseDecl(node.Decl, pkgScopeNames, imports)
	case *ast.LabeledStmt:
		ctx.traverseStmt(node.Stmt, pkgScopeNames, imports)
	case *ast.BranchStmt:
		// ignore continue or continue label
	case *ast.ReturnStmt:
		ctx.traverseExprs(node.Results, pkgScopeNames, imports)
	case *ast.BlockStmt:
		ctx.traverseBlockStmt(node, pkgScopeNames, imports)
	case *ast.IncDecStmt:
		// ignore
	case *ast.CaseClause:
		ctx.traverseBlockStmt(&ast.BlockStmt{List: node.Body}, pkgScopeNames, imports)
	case *ast.CommClause:
		// select {...}
		// TODO: handle name introduced in case
		ctx.traverseBlockStmt(&ast.BlockStmt{List: node.Body}, pkgScopeNames, imports)
	case *ast.SendStmt:
		// ignore
		ctx.traverseExpr(node.Chan, pkgScopeNames, imports)
		ctx.traverseExpr(node.Value, pkgScopeNames, imports)
	default:
		errorUnknown("stmt", node)
	}
}
