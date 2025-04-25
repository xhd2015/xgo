package syntax

import (
	"cmd/compile/internal/base"
	"cmd/compile/internal/syntax"
	"cmd/compile/internal/xgo_rewrite_internal/patch/info"
	"cmd/compile/internal/xgo_rewrite_internal/patch/instrument/compiler_extra"
	"os"
	"path/filepath"

	"strconv"
)

// check instrument/instrument_reg/signature.go for signature
// func(info interface{}, recvPtr interface{}, args []interface{}, results []interface{}) (func(), bool)
const XgoTrapPrefix = "__xgo_trap_"

// corresponding to legacy rewriteFuncsSource
// proto type:
//
//	  afterV,stopV := __xgo_link_trap_for_generated(pkgPath,pc=0,identityName, isGeneric, &recv,&args,&results)
//	  if afterV!=nil{
//		        defer afterV()
//	  }
//	  if stopV {
//		    return
//	  }
func trapFuncs(funcDecls []*info.DeclInfo, xgoTrap string, fileMapping map[string]*compiler_extra.FileMapping) int {
	count := 0
	for _, fn := range funcDecls {
		if isBlankName(fn.Name) {
			continue
		}
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

		// this solution was originally used for stdlib and generic
		// now it should be used for all functions
		//
		// stdlib and generic
		// if !base.Flag.Std {
		// 	if !fn.Generic {
		// 		continue
		// 	}
		// }
		baseFile := filepath.Base(fn.File)
		file := fileMapping[baseFile]
		if file == nil {
			continue
		}
		funcInfo := file.Funcs[fn.IdentityName()]
		if funcInfo == nil {
			continue
		}

		fnDecl := fn.FuncDecl
		fnType := fnDecl.Type
		pos := fn.FuncDecl.Pos()

		// fmt.Fprintf(os.Stderr, "DEBUG: trapFuncs %s\n", fn.IdentityName())
		// check if body contains recover(), if so
		// do not add interceptor
		// see https://github.com/xhd2015/xgo/issues/164
		// UPDATE: since we are not declaring new functions,
		// recover() is just fine
		// if hasRecoverCall(fnDecl.Body) {
		// 	continue
		// }

		preset := getPresetNames(fnDecl)
		if preset[xgoTrap] {
			// cannot trap because name conflict
			continue
		}

		// stop if __xgo_link_generated_trap conflict with recv?
		fillNames(pos, fnDecl.Recv, fnDecl.Type, preset)
		preset[xgoTrap] = true

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
		//   postV,stopV := __xgo_trap_{fileIndex}}(&recv,&args,&results)
		//   if postV!=nil{
		//	        defer postV()
		//   }
		//   if stopV {
		//	    return
		//   }
		postV := nextName("_post", "", preset)
		stopV := nextName("_stop", "", preset)
		prependStmts := make([]syntax.Stmt, 0, 3+len(fnDecl.Body.List))
		prependStmts = append(prependStmts,
			&syntax.AssignStmt{
				Op: syntax.Def,
				Lhs: &syntax.ListExpr{
					ElemList: []syntax.Expr{
						syntax.NewName(pos, postV),
						syntax.NewName(pos, stopV),
					},
				},
				Rhs: &syntax.CallExpr{
					Fun: syntax.NewName(pos, xgoTrap),
					ArgList: []syntax.Expr{
						syntax.NewName(pos, fn.FuncInfoVarName()),
						recvRef,
						argAddrs,
						resultAddrs,
					},
				},
			},

			&syntax.IfStmt{
				Cond: &syntax.Operation{
					Op: syntax.Neq,
					X:  syntax.NewName(pos, postV),
					Y:  syntax.NewName(pos, "nil"),
				},
				Then: &syntax.BlockStmt{
					List: []syntax.Stmt{
						&syntax.CallStmt{
							Tok: syntax.Defer,
							Call: &syntax.CallExpr{
								Fun: syntax.NewName(pos, postV),
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
		fn.HasFuncTrap = true
		if false {
			// debug
			syntax.Fdump(os.Stderr, fnDecl)
		}
		count++
	}
	return count
}

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
