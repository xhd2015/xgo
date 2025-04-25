package syntax

import (
	"cmd/compile/internal/base"
	"cmd/compile/internal/syntax"
	xgo_ctxt "cmd/compile/internal/xgo_rewrite_internal/patch/ctxt"
	"fmt"
	"os"

	"strconv"
)

const XgoLinkTrapForGenerated = "__xgo_link_trap_for_generated"

// for closures outside stdlib
// why closure needs names?
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

const patchedTimeNow = "Now_Xgo_Original"
const patchedTimeSince = "Since_Xgo_Original"

func addTimePatch(funcDelcs []*DeclInfo) {
	for _, fn := range funcDelcs {
		if fn.RecvTypeName != "" {
			continue
		}
		if fn.Name == "Now" {
			newNow := copyFuncDecl(fn.FuncDecl, true)
			newNow.Name.Value = patchedTimeNow
			fn.FileSyntax.DeclList = append(fn.FileSyntax.DeclList, newNow)

			fillPos(fn.FuncDecl.Pos(), newNow)
			continue
		}
		if fn.Name == "Since" {
			newSince := copyFuncDecl(fn.FuncDecl, true)
			newSince.Name.Value = patchedTimeSince
			syntax.Inspect(newSince, func(n syntax.Node) bool {
				if callExpr, ok := n.(*syntax.CallExpr); ok {
					if callName, ok := callExpr.Fun.(*syntax.Name); ok && callName.Value == "Now" {
						callName.Value = patchedTimeNow
					}
				}
				return true
			})
			fn.FileSyntax.DeclList = append(fn.FileSyntax.DeclList, newSince)

			fillPos(fn.FuncDecl.Pos(), newSince)
			continue
		}
	}
}

func rewriteTimePatch(funcDelcs []*DeclInfo) {
	for _, fn := range funcDelcs {
		if fn.RecvTypeName != "" {
			continue
		}
		if fn.Name == "timeNow" {
			replaceIdent(fn.FuncDecl, "Now", patchedTimeNow)
			continue
		}
		if fn.Name == "timeSince" {
			replaceIdent(fn.FuncDecl, "Since", patchedTimeSince)
			continue
		}
	}
}

func replaceIdent(root syntax.Node, match string, to string) {
	syntax.Inspect(root, func(n syntax.Node) bool {
		if name, ok := n.(*syntax.Name); ok && name != nil && name.Value == match {
			name.Value = to
		}
		return true
	})
}

func rewriteFuncsSource(funcDecls []*DeclInfo, pkgPath string) {
	defer xgo_ctxt.LogSpan("rewriteFuncsSource")()
	for _, fn := range funcDecls {
		if !fn.Kind.IsFunc() {
			continue
		}
		if fn.Interface {
			continue
		}
		if fn.Closure {
			continue
		}
		if fn.FuncDecl.Body == nil {
			// no body, may be linked
			continue
		}

		// stdlib and generic
		if !base.Flag.Std {
			if !fn.Generic {
				continue
			}
		}
		fnDecl := fn.FuncDecl
		fnType := fnDecl.Type
		pos := fn.FuncDecl.Pos()

		// check if body contains recover(), if so
		// do not add interceptor
		// see https://github.com/xhd2015/xgo/issues/164
		// UPDATE: since we are not declaring new functions,
		// recover() is just fine
		// if hasRecoverCall(fnDecl.Body) {
		// 	continue
		// }

		preset := getPresetNames(fnDecl)
		if preset[XgoLinkTrapForGenerated] {
			// cannot trap because name conflict
			continue
		}

		// stop if __xgo_link_generated_trap conflict with recv?
		fillNames(pos, fnDecl.Recv, fnDecl.Type, preset)
		preset[XgoLinkTrapForGenerated] = true

		idName := fn.IdentityName()
		var recvRef syntax.Expr
		if fnDecl.Recv != nil {
			recvName := fnDecl.Recv.Name.Value
			recvRef = &syntax.Operation{
				Op: syntax.And,
				X:  syntax.NewName(pos, recvName),
			}
		} else {
			recvRef = syntax.NewName(pos, "nil")
		}
		argAddrs := getRefAddrSlice(pos, fnType.ParamList)
		resultAddrs := getRefAddrSlice(pos, fnType.ResultList)

		// proto type:
		//    afterV,stopV := __xgo_link_trap_for_generated(pkgPath,pc=0,identityName, isGeneric, &recv,&args,&results)
		//   if afterV!=nil{
		//	        defer afterV()
		//   }
		//   if stopV {
		//	    return
		//   }
		afterV := nextName("_after", "", preset)
		stopV := nextName("_stop", "", preset)
		prependStmts := make([]syntax.Stmt, 0, 3+len(fnDecl.Body.List))
		prependStmts = append(prependStmts,
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
		)
		for _, p := range prependStmts {
			fillPos(pos, p)
		}
		fnDecl.Body.List = append(prependStmts, fnDecl.Body.List...)
		if false {
			// debug
			syntax.Fdump(os.Stderr, fnDecl)
		}
	}
}

func makeCallToOldFuncStmt(pos syntax.Pos, fnDecl *syntax.FuncDecl, oldFnName string) syntax.Stmt {
	var callOldFunc syntax.Expr
	if fnDecl.Recv != nil {
		recvName := fnDecl.Recv.Name.Value
		callOldFunc = &syntax.SelectorExpr{
			X:   syntax.NewName(pos, recvName),
			Sel: syntax.NewName(pos, oldFnName),
		}
	} else {
		callOldFunc = syntax.NewName(pos, oldFnName)

		// need TParams
		tparams := fnDecl.TParamList
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

	fnTypeCopy := fnDecl.Type
	argValues := getRefSlice(pos, fnTypeCopy.ParamList)

	var hasDots bool
	if len(fnTypeCopy.ParamList) > 0 {
		_, hasDots = fnTypeCopy.ParamList[len(fnTypeCopy.ParamList)-1].Type.(*syntax.DotsType)
	}
	callOldExpr := &syntax.CallExpr{
		Fun:     callOldFunc,
		ArgList: argValues,
		HasDots: hasDots,
	}
	var stmt syntax.Stmt
	// var lastReturn syntax.Stmt
	if len(fnTypeCopy.ResultList) > 0 {
		stmt = &syntax.ReturnStmt{
			Results: callOldExpr,
		}
	} else {
		stmt = makeEmptyCallStmt(callOldExpr)
	}
	return stmt
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

// auto fill unnamed parameters
func fillNames(pos syntax.Pos, recv *syntax.Field, funcType *syntax.FuncType, preset map[string]bool) {
	if recv != nil {
		fillFieldNames(pos, []*syntax.Field{recv}, preset, "_x")
	}
	fillFieldNames(pos, funcType.ParamList, preset, "_a")
	fillFieldNames(pos, funcType.ResultList, preset, "_r")
}
func getFuncDeclNamesNoBlank(recv *syntax.Field, funcType *syntax.FuncType) []string {
	var names []string
	recvName := getFieldName(recv)
	if !isBlankName(recvName) {
		names = append(names, recvName)
	}
	paramNames := getFieldNames(funcType.ParamList)
	for _, name := range paramNames {
		if !isBlankName(name) {
			names = append(names, name)
		}
	}
	resultNames := getFieldNames(funcType.ResultList)
	for _, name := range resultNames {
		if !isBlankName(name) {
			names = append(names, name)
		}
	}
	return names
}

func isBlankName(name string) bool {
	return name == "" || name == "_"
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

func hasRecoverCall(node syntax.Node) bool {
	var found bool
	syntax.Inspect(node, func(n syntax.Node) bool {
		if n == nil {
			return false
		}
		if call, ok := n.(*syntax.CallExpr); ok {
			if idt, ok := call.Fun.(*syntax.Name); ok && idt.Value == "recover" {
				found = true
				return false
			}
		}
		return true
	})
	return found
}

func copyFuncDeclWithoutBody(decl *syntax.FuncDecl) *syntax.FuncDecl {
	return copyFuncDecl(decl, false)
}

func copyFuncDecl(decl *syntax.FuncDecl, withBody bool) *syntax.FuncDecl {
	if decl == nil {
		return nil
	}
	x := *decl
	x.Recv = copyField(decl.Recv)
	x.Name = copyName(decl.Name)
	x.TParamList = copyFields(decl.TParamList)
	x.Type = copyFuncType(decl.Type)
	if withBody {
		x.Body = copyBlockStmts(decl.Body)
	} else {
		x.Body = nil
	}
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

func copyBasicLiterals(literals []*syntax.BasicLit) []*syntax.BasicLit {
	c := make([]*syntax.BasicLit, len(literals))
	for i := 0; i < len(literals); i++ {
		c[i] = copyBasicLit(literals[i])
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

func copyNames(names []*syntax.Name) []*syntax.Name {
	if names == nil {
		return nil
	}
	n := len(names)
	copiedNames := make([]*syntax.Name, n)
	for i := 0; i < n; i++ {
		copiedNames[i] = copyName(names[i])
	}
	return copiedNames
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

func copySimpleStmt(stmt syntax.SimpleStmt) syntax.SimpleStmt {
	if stmt == nil {
		return nil
	}
	switch stmt := stmt.(type) {
	case *syntax.AssignStmt:
		x := *stmt
		x.Lhs = copyExpr(stmt.Lhs)
		x.Rhs = copyExpr(stmt.Rhs)
		return &x
	case *syntax.SendStmt:
		x := *stmt
		x.Chan = copyExpr(stmt.Chan)
		x.Value = copyExpr(stmt.Value)
		return &x
	case *syntax.ExprStmt:
		x := *stmt
		x.X = copyExpr(stmt.X)
		return &x
	case *syntax.EmptyStmt:
		x := *stmt
		return &x
	case *syntax.RangeClause:
		x := *stmt
		x.Lhs = copyExpr(stmt.Lhs)
		x.X = copyExpr(stmt.X)
		return &x
	default:
		panic(fmt.Errorf("unrecognized simple stmt: %T", stmt))
	}
}
func copyCaseClause(n *syntax.CaseClause) *syntax.CaseClause {
	if n == nil {
		return nil
	}
	x := *n
	x.Cases = copyExpr(n.Cases)
	x.Body = copyStmts(n.Body)
	return &x
}

func copyCaseClauses(list []*syntax.CaseClause) []*syntax.CaseClause {
	if list == nil {
		return nil
	}
	n := len(list)
	copiedList := make([]*syntax.CaseClause, n)
	for i := 0; i < n; i++ {
		copiedList[i] = copyCaseClause(list[i])
	}
	return copiedList
}

func copyCommClause(n *syntax.CommClause) *syntax.CommClause {
	if n == nil {
		return nil
	}
	x := *n
	x.Comm = copySimpleStmt(n.Comm)
	x.Body = copyStmts(n.Body)
	return &x
}

func copyCommClauses(list []*syntax.CommClause) []*syntax.CommClause {
	if list == nil {
		return nil
	}
	n := len(list)
	copiedList := make([]*syntax.CommClause, n)
	for i := 0; i < n; i++ {
		copiedList[i] = copyCommClause(list[i])
	}
	return copiedList
}
func copyDeclList(decls []syntax.Decl) []syntax.Decl {
	if decls == nil {
		return nil
	}
	n := len(decls)
	copiedDecls := make([]syntax.Decl, n)
	for i := 0; i < n; i++ {
		copiedDecls[i] = copyDecl(decls[i])
	}
	return copiedDecls
}
func copyDecl(decl syntax.Decl) syntax.Decl {
	if decl == nil {
		return nil
	}
	switch decl := decl.(type) {
	case *syntax.ConstDecl:
		x := *decl
		if decl.Group != nil {
			g := *decl.Group
			x.Group = &g
		}
		x.NameList = copyNames(decl.NameList)
		x.Type = copyExpr(decl.Type)
		x.Values = copyExpr(decl.Values)
		return &x
	case *syntax.VarDecl:
		x := *decl
		if decl.Group != nil {
			g := *decl.Group
			x.Group = &g
		}
		x.NameList = copyNames(decl.NameList)
		x.Type = copyExpr(decl.Type)
		x.Values = copyExpr(decl.Values)
		return &x
	case *syntax.TypeDecl:
		x := *decl
		if decl.Group != nil {
			g := *decl.Group
			x.Group = &g
		}
		x.Name = copyName(decl.Name)
		x.TParamList = copyFields(decl.TParamList)
		x.Type = copyExpr(decl.Type)
		return &x
	default:
		panic(fmt.Errorf("unrecognized decl: %T", decl))
	}
}
func copyStmt(stmt syntax.Stmt) syntax.Stmt {
	if stmt == nil {
		return nil
	}
	switch stmt := stmt.(type) {
	case syntax.SimpleStmt:
		return copySimpleStmt(stmt)
	case *syntax.BlockStmt:
		return copyBlockStmts(stmt)
	case *syntax.IfStmt:
		x := *stmt
		x.Init = copySimpleStmt(stmt.Init)
		x.Cond = copyExpr(stmt.Cond)
		x.Then = copyBlockStmts(stmt.Then)
		x.Else = copyStmt(stmt.Else)
		return &x
	case *syntax.CallStmt:
		x := *stmt
		x.Call = copyCallExpr(stmt.Call)
		return &x
	case *syntax.DeclStmt:
		x := *stmt
		x.DeclList = copyDeclList(stmt.DeclList)
		return &x
	case *syntax.ReturnStmt:
		x := *stmt
		x.Results = copyExpr(stmt.Results)
		return &x
	case *syntax.SwitchStmt:
		x := *stmt
		x.Init = copySimpleStmt(stmt.Init)
		x.Tag = copyExpr(stmt.Tag)
		x.Body = copyCaseClauses(stmt.Body)
		return &x
	}
	panic(fmt.Sprintf("unrecognized stmt: %T", stmt))
}
func copyStmts(stmts []syntax.Stmt) []syntax.Stmt {
	cpStmts := make([]syntax.Stmt, len(stmts))
	for i := 0; i < len(stmts); i++ {
		cpStmts[i] = copyStmt(stmts[i])
	}
	return cpStmts
}

func copyBlockStmts(stmt *syntax.BlockStmt) *syntax.BlockStmt {
	if stmt == nil {
		return nil
	}
	x := *stmt
	x.List = copyStmts(x.List)
	return &x
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
	case *syntax.CallExpr:
		x := *expr
		x.Fun = copyExpr(expr.Fun)
		x.ArgList = copyExprs(expr.ArgList)
		return &x
	case *syntax.ParenExpr:
		x := *expr
		x.X = copyExpr(expr.X)
		return &x
	case *syntax.DotsType:
		x := *expr
		x.Elem = copyExpr(expr.Elem)
		return &x
	case *syntax.StructType:
		x := *expr
		x.FieldList = copyFields(expr.FieldList)
		x.TagList = copyBasicLiterals(expr.TagList)
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
	case *syntax.CompositeLit:
		x := *expr
		x.Type = copyExpr(expr.Type)
		x.ElemList = copyExprs(expr.ElemList)
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
func takeNameAddr(pos syntax.Pos, name string) *syntax.Operation {
	return takeExprAddr(syntax.NewName(pos, name))
}

func takeExprAddr(expr syntax.Expr) *syntax.Operation {
	return &syntax.Operation{
		Op: syntax.And,
		X:  expr,
	}
}

func newIntLit(i int) *syntax.BasicLit {
	return &syntax.BasicLit{
		Value: strconv.FormatInt(int64(i), 10),
		Kind:  syntax.IntLit,
	}
}
func newBool(pos syntax.Pos, b bool) *syntax.Name {
	return syntax.NewName(pos, strconv.FormatBool(b))
}

// func newBoolLit(b bool) *syntax.BasicLit {
// 	return &syntax.BasicLit{
// 		Value: strconv.FormatBool(b),
// 		Kind:  syntax.IntLit,
// 	}
// }
