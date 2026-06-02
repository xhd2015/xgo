package ast

import (
	"fmt"
	"go/ast"
)

func ParseReceiverInfo(fnName string, receiverType ast.Expr) (identityName string, recvPtr bool, recvGeneric bool, recvType *ast.Ident) {
	recvPtr, recvGeneric, recvType = ParseReceiverType(receiverType)
	recvTypeCode := recvType.Name
	if recvPtr {
		identityName = "(*" + recvTypeCode + ")." + fnName
	} else {
		identityName = recvTypeCode + "." + fnName
	}
	return
}

// ParseReceiverType parses a method receiver type expression, extracting
// whether it is a pointer receiver, a generic receiver, and the underlying
// type name as an *ast.Ident.
//
// It handles these receiver forms (all unwrapped to the base type name):
//
//	T         → ptr=false, generic=false, recv="T"
//	*T        → ptr=true,  generic=false, recv="T"
//	T[K]      → ptr=false, generic=true,  recv="T"
//	*T[K]     → ptr=true,  generic=true,  recv="T"
//	pkg.T     → ptr=false, generic=false, recv="T"  (ast.SelectorExpr)
//	*pkg.T    → ptr=true,  generic=false, recv="T"  (StarExpr→SelectorExpr)
//	*pkg.T[K] → ptr=true,  generic=true,  recv="T"  (StarExpr→IndexExpr→SelectorExpr)
//
// The SelectorExpr case occurs when a method receiver references a type from
// another package (e.g., *otherpkg.Type). The Go AST represents this as
// *ast.StarExpr{X: *ast.SelectorExpr{X: pkg, Sel: Type}}. Previously this
// caused a panic because the function only checked for *ast.Ident after
// unwrapping pointer and generic wrappers. The fix extracts selExpr.Sel,
// which is the *ast.Ident for the type name portion (e.g., "Type").
func ParseReceiverType(typeExpr ast.Expr) (ptr bool, generic bool, recvType *ast.Ident) {
	orig := typeExpr
	starExpr, ok := typeExpr.(*ast.StarExpr)
	if ok {
		ptr = true
		typeExpr = starExpr.X
	}
	indexExpr, ok := typeExpr.(*ast.IndexExpr)
	if ok {
		generic = true
		typeExpr = indexExpr.X
	} else {
		x, ok := extractIndexListExpr(typeExpr)
		if ok {
			generic = true
			typeExpr = x
		}
	}
	selExpr, ok := typeExpr.(*ast.SelectorExpr)
	if ok {
		return ptr, generic, selExpr.Sel
	}
	idt, ok := typeExpr.(*ast.Ident)
	if !ok {
		panic(fmt.Errorf("expect receiver to be ident, actual: %T(from %T)", typeExpr, orig))
	}
	return ptr, generic, idt
}
