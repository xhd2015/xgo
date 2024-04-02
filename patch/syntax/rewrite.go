package syntax

import (
	"cmd/compile/internal/base"
	"cmd/compile/internal/syntax"
	"fmt"
	"os"

	"strconv"
)

const XgoLinkTrapForGenerated = "__xgo_link_trap_for_generated"

func fillFuncArgResNames(fileList []*syntax.File) {
	if base.Flag.Std {
		return
	}
	for _, file := range fileList {
		syntax.Inspect(file, func(n syntax.Node) bool {
			if decl, ok := n.(*syntax.FuncDecl); ok {
				if decl.Body == nil {
					return true
				}
				preset := getPresetNames(decl)
				fillNames(decl.Pos(), decl.Recv, decl.Type, preset)
			} else if funcLit, ok := n.(*syntax.FuncLit); ok {
				preset := getPresetNames(funcLit.Type)
				fillNames(funcLit.Pos(), nil, funcLit.Type, preset)
			}

			return true
		})
	}
}

func rewriteStdAndGenericFuncs(funcDecls []*DeclInfo, pkgPath string) {
	for _, fn := range funcDecls {
		if fn.Interface {
			continue
		}
		if fn.Closure {
			continue
		}

		// stdlib and generic
		if !base.Flag.Std {
			if !fn.Generic {
				continue
			}
		}

		fnDecl := fn.FuncDecl
		pos := fn.FuncDecl.Pos()

		fnName := fnDecl.Name.Value

		// dump
		// syntax.Fdump(os.Stderr, fn.FuncDecl.Body)

		// i
		newDecl := copyFuncDeclWithoutBody(fnDecl)
		oldFnName := "__xgo_orig_" + fnName
		fnDecl.Name.Value = oldFnName

		preset := getPresetNames(newDecl)

		fillNames(pos, newDecl.Recv, newDecl.Type, preset)
		if preset[XgoLinkTrapForGenerated] {
			// cannot trap
			continue
		}
		// stop if __xgo_link_generated_trap conflict with recv?
		preset[XgoLinkTrapForGenerated] = true

		afterV := nextName("_after", "", preset)
		stopV := nextName("_stop", "", preset)

		idName := fn.IdentityName()

		var callOldFunc syntax.Expr
		var recvRef syntax.Expr
		if newDecl.Recv != nil {
			recvName := newDecl.Recv.Name.Value
			recvRef = &syntax.Operation{
				Op: syntax.And,
				X:  syntax.NewName(pos, recvName),
			}
			callOldFunc = &syntax.SelectorExpr{
				X:   syntax.NewName(pos, recvName),
				Sel: syntax.NewName(pos, oldFnName),
			}
		} else {
			recvRef = syntax.NewName(pos, "nil")
			callOldFunc = syntax.NewName(pos, oldFnName)

			// need TParams
			tparams := newDecl.TParamList
			if len(tparams) > 0 {
				tparamExprs := make([]syntax.Expr, len(tparams))
				for i, tparam := range tparams {
					tparamExprs[i] = syntax.NewName(pos, tparam.Name.Value)
				}
				var indexExpr syntax.Expr
				if len(tparamExprs) == 1 {
					indexExpr = tparamExprs[0]
				} else {
					indexExpr = &syntax.ListExpr{
						ElemList: tparamExprs,
					}
				}
				callOldFunc = &syntax.IndexExpr{
					X:     callOldFunc,
					Index: indexExpr,
				}
			}
		}

		fnTypeCopy := newDecl.Type
		argValues := getRefSlice(pos, fnTypeCopy.ParamList)
		argAddrs := getRefAddrSlice(pos, fnTypeCopy.ParamList)
		resultAddrs := getRefAddrSlice(pos, fnTypeCopy.ResultList)

		var hasDots bool
		if len(fnTypeCopy.ParamList) > 0 {
			_, hasDots = fnTypeCopy.ParamList[len(fnTypeCopy.ParamList)-1].Type.(*syntax.DotsType)
		}
		callOldExpr := &syntax.CallExpr{
			Fun:     callOldFunc,
			ArgList: argValues,
			HasDots: hasDots,
		}
		var lastStmt syntax.Stmt
		// var lastReturn syntax.Stmt
		if len(fnTypeCopy.ResultList) > 0 {
			lastStmt = &syntax.ReturnStmt{
				Results: callOldExpr,
			}
		} else {
			lastStmt = makeEmptyCallStmt(callOldExpr)
		}

		newDecl.Body = &syntax.BlockStmt{
			Rbrace: pos,
			List: []syntax.Stmt{
				&syntax.AssignStmt{
					Op: syntax.Def,
					Lhs: &syntax.ListExpr{
						ElemList: []syntax.Expr{
							syntax.NewName(pos, afterV),
							syntax.NewName(pos, stopV),
						},
					},
					Rhs: &syntax.CallExpr{
						Fun: syntax.NewName(pos, XgoLinkTrapForGenerated),
						ArgList: []syntax.Expr{
							newStringLit(pkgPath),
							newIntLit(0), // pc, filled by IR
							newStringLit(idName),
							syntax.NewName(pos, strconv.FormatBool(fn.Generic)),
							recvRef,
							argAddrs,
							resultAddrs,
						},
					},
				},

				&syntax.IfStmt{
					Cond: &syntax.Operation{
						Op: syntax.Neq,
						X:  syntax.NewName(pos, afterV),
						Y:  syntax.NewName(pos, "nil"),
					},
					Then: &syntax.BlockStmt{
						List: []syntax.Stmt{
							&syntax.CallStmt{
								Tok: syntax.Defer,
								Call: &syntax.CallExpr{
									Fun: syntax.NewName(pos, afterV),
								},
							},
						},
					},
				},
				// debug
				// &syntax.AssignStmt{
				// 	Lhs: &syntax.ListExpr{
				// 		ElemList: []syntax.Expr{
				// 			syntax.NewName(pos, "_"),
				// 			syntax.NewName(pos, "_"),
				// 		},
				// 	},
				// 	Rhs: &syntax.ListExpr{
				// 		ElemList: []syntax.Expr{
				// 			syntax.NewName(pos, afterV),
				// 			syntax.NewName(pos, stopV),
				// 		},
				// 	},
				// },
				&syntax.IfStmt{
					Cond: syntax.NewName(pos, stopV),
					Then: &syntax.BlockStmt{
						List: []syntax.Stmt{
							&syntax.ReturnStmt{},
						},
						Rbrace: pos,
					},
				},
				lastStmt,
			},
		}
		fillPos(pos, newDecl)
		fn.FileSyntax.DeclList = append(fn.FileSyntax.DeclList, newDecl)

		if false {
			// debug
			syntax.Fdump(os.Stderr, newDecl)
		}
	}
}

func makeEmptyCallStmt(callExpr *syntax.CallExpr) syntax.Stmt {
	return &syntax.ExprStmt{
		X: callExpr,
	}
}

type ISetPos interface {
	SetPos(p syntax.Pos)
}

func fillPos(pos syntax.Pos, node syntax.Node) {
	syntax.Inspect(node, func(n syntax.Node) bool {
		if n == nil {
			return false
		}
		n.(ISetPos).SetPos(pos)
		return true
	})

}
func fillNames(pos syntax.Pos, recv *syntax.Field, funcType *syntax.FuncType, preset map[string]bool) {
	if recv != nil {
		fillFieldNames(pos, []*syntax.Field{recv}, preset, "_x")
	}
	fillFieldNames(pos, funcType.ParamList, preset, "_a")
	fillFieldNames(pos, funcType.ResultList, preset, "_r")
}

func getRefSlice(pos syntax.Pos, fields []*syntax.Field) []syntax.Expr {
	return doGetRefAddrSlice(pos, fields, false)
}
func getRefAddrSlice(pos syntax.Pos, fields []*syntax.Field) *syntax.CompositeLit {
	names := doGetRefAddrSlice(pos, fields, true)
	return &syntax.CompositeLit{
		Type: &syntax.SliceType{
			Elem: &syntax.InterfaceType{},
		},
		ElemList: names,
		Rbrace:   pos,
	}
}
func doGetRefAddrSlice(pos syntax.Pos, fields []*syntax.Field, addr bool) []syntax.Expr {
	names := make([]syntax.Expr, len(fields))
	for i, f := range fields {
		name := syntax.NewName(pos, f.Name.Value)
		if !addr {
			names[i] = name
		} else {
			names[i] = &syntax.Operation{
				Op: syntax.And,
				X:  name,
			}
		}
	}
	return names
}

// _a0,_a1
func fillFieldNames(pos syntax.Pos, fields []*syntax.Field, preset map[string]bool, prefix string) {
	for i, f := range fields {
		if f.Name == nil {
			name := &syntax.Name{}
			(ISetPos)(name).SetPos(pos)
			f.Name = name
		} else if f.Name.Value != "" && f.Name.Value != "_" {
			continue
		}
		suffix := strconv.FormatInt(int64(i), 10)
		f.Name.Value = nextName(prefix, suffix, preset)
	}
}

func nextName(prefix string, suffix string, preset map[string]bool) string {
	mid := ""
	for {
		name := prefix + mid + suffix
		if !preset[name] {
			preset[name] = true
			return name
		}
		mid = mid + "_"
	}
}
func addPresetNames(node syntax.Node, set map[string]bool) {
	m := getPresetNames(node)
	for k := range m {
		set[k] = true
	}
}

// prevent all ident appeared in func type
// NOTE: the returned map may contain "_", ""
func getPresetNames(node syntax.Node) map[string]bool {
	preset := make(map[string]bool)
	syntax.Inspect(node, func(n syntax.Node) bool {
		if n == nil {
			return false
		}
		if idt, ok := n.(*syntax.Name); ok {
			preset[idt.Value] = true
		}
		return true
	})
	return preset
}

func copyFuncDeclWithoutBody(decl *syntax.FuncDecl) *syntax.FuncDecl {
	if decl == nil {
		return nil
	}
	x := *decl
	x.Recv = copyField(decl.Recv)
	x.Name = copyName(decl.Name)
	x.TParamList = copyFields(decl.TParamList)
	x.Type = copyFuncType(decl.Type)

	return &x
}
func copyFuncType(typ_ *syntax.FuncType) *syntax.FuncType {
	if typ_ == nil {
		return nil
	}
	x := *typ_
	x.ParamList = copyFields(typ_.ParamList)
	x.ResultList = copyFields(typ_.ResultList)
	return &x
}
func copyFields(fields []*syntax.Field) []*syntax.Field {
	c := make([]*syntax.Field, len(fields))
	for i := 0; i < len(fields); i++ {
		c[i] = copyField(fields[i])
	}
	return c
}

func copyBasicLits(lits []*syntax.BasicLit) []*syntax.BasicLit {
	c := make([]*syntax.BasicLit, len(lits))
	for i := 0; i < len(lits); i++ {
		c[i] = copyBasicLit(lits[i])
	}
	return c
}
func copyBasicLit(lit *syntax.BasicLit) *syntax.BasicLit {
	if lit == nil {
		return nil
	}
	x := *lit
	return &x
}
func copyField(field *syntax.Field) *syntax.Field {
	if field == nil {
		return nil
	}
	x := *field
	x.Name = copyName(field.Name)
	x.Type = copyExpr(field.Type)
	return &x
}
func copyName(name *syntax.Name) *syntax.Name {
	if name == nil {
		return nil
	}
	x := *name
	x.Value = name.Value
	return &x
}

func copyExprs(exprs []syntax.Expr) []syntax.Expr {
	copyExprs := make([]syntax.Expr, len(exprs))
	for i, expr := range exprs {
		copyExprs[i] = copyExpr(expr)
	}
	return copyExprs
}

func copyExpr(expr syntax.Expr) syntax.Expr {
	if expr == nil {
		return nil
	}
	switch expr := expr.(type) {
	case *syntax.Name:
		return copyName(expr)
	case *syntax.BasicLit:
		x := *expr
		return &x
	case *syntax.IndexExpr:
		x := *expr
		x.X = copyExpr(expr.X)
		x.Index = copyExpr(expr.Index)
		return &x
	case *syntax.Operation:
		x := *expr
		x.X = copyExpr(expr.X)
		x.Y = copyExpr(expr.Y)
		return &x
	case *syntax.DotsType:
		x := *expr
		x.Elem = copyExpr(expr.Elem)
		return &x
	case *syntax.StructType:
		x := *expr
		x.FieldList = copyFields(expr.FieldList)
		x.TagList = copyBasicLits(expr.TagList)
		return &x
	case *syntax.FuncType:
		return copyFuncType(expr)
	case *syntax.SliceExpr:
		x := *expr
		x.X = copyExpr(expr.X)
		x.Index[0] = copyExpr(expr.Index[0])
		x.Index[1] = copyExpr(expr.Index[1])
		x.Index[2] = copyExpr(expr.Index[2])
		return &x
	case *syntax.SliceType:
		x := *expr
		x.Elem = copyExpr(expr.Elem)
		return &x
	case *syntax.ArrayType:
		x := *expr
		x.Len = copyExpr(expr.Len)
		x.Elem = copyExpr(expr.Elem)
		return &x
	case *syntax.SelectorExpr:
		x := *expr
		x.X = copyExpr(expr.X)
		x.Sel = copyName(expr.Sel)
		return &x
	case *syntax.ChanType:
		x := *expr
		x.Elem = copyExpr(expr.Elem)
		return &x
	case *syntax.InterfaceType:
		x := *expr
		x.MethodList = copyFields(expr.MethodList)
		return &x
	case *syntax.ListExpr:
		x := *expr
		x.ElemList = copyExprs(expr.ElemList)
		return &x
	case *syntax.MapType:
		x := *expr
		x.Key = copyExpr(expr.Key)
		x.Value = copyExpr(expr.Value)
		return &x
	default:
		panic(fmt.Errorf("unrecognized expr while copying: %T", expr))
	}
}

func newPanicCallExprStmt(pos syntax.Pos, s string) syntax.Stmt {
	lit := newStringLit(s)
	expr := &syntax.CallExpr{
		Fun: syntax.NewName(pos, "panic"),
		ArgList: []syntax.Expr{
			lit,
		},
	}
	exprStmt := &syntax.ExprStmt{
		X: expr,
	}
	// exprStmt.SetPos(syntax.Pos{})
	return exprStmt
}

func newStringLit(s string) *syntax.BasicLit {
	return &syntax.BasicLit{
		Value: strconv.Quote(s),
		Kind:  syntax.StringLit,
	}
}
func newIntLit(i int) *syntax.BasicLit {
	return &syntax.BasicLit{
		Value: strconv.FormatInt(int64(i), 10),
		Kind:  syntax.IntLit,
	}
}

// func newBoolLit(b bool) *syntax.BasicLit {
// 	return &syntax.BasicLit{
// 		Value: strconv.FormatBool(b),
// 		Kind:  syntax.IntLit,
// 	}
// }
