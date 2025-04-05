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
	idt, ok := typeExpr.(*ast.Ident)
	if !ok {
		panic(fmt.Errorf("expect receiver to be ident, actual: %T(from %T)", typeExpr, orig))
	}
	return ptr, generic, idt
}
