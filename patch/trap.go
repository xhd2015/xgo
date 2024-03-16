package patch

import (
	"fmt"
	"go/constant"
	"os"
	"strings"

	"cmd/compile/internal/base"
	"cmd/compile/internal/ir"
	"cmd/compile/internal/reflectdata"
	"cmd/compile/internal/syntax"
	"cmd/compile/internal/typecheck"
	"cmd/compile/internal/types"
	"cmd/internal/src"

	xgo_func_name "cmd/compile/internal/xgo_rewrite_internal/patch/func_name"
	xgo_record "cmd/compile/internal/xgo_rewrite_internal/patch/record"
	xgo_syntax "cmd/compile/internal/xgo_rewrite_internal/patch/syntax"
)

const disableXgoLink bool = false
const disableTrap bool = false

func Patch() {
	debugIR()
	if os.Getenv("COMPILER_ALLOW_IR_REWRITE") != "true" {
		return
	}
	insertTrapPoints()
	initRegFuncs()
}

const xgoRuntimePkgPrefix = "github.com/xhd2015/xgo/runtime/"
const xgoRuntimeTrapPkg = xgoRuntimePkgPrefix + "trap"

func insertTrapPoints() {
	files := xgo_syntax.GetFiles()
	syntaxDecls := xgo_syntax.GetDecls()
	xgo_syntax.ClearFiles() // help GC
	xgo_syntax.ClearDecls()

	// check if any file has __XGO_SKIP_TRAP
	var skipTrap bool
	for _, f := range files {
		for _, d := range f.DeclList {
			if d, ok := d.(*syntax.ConstDecl); ok && len(d.NameList) > 0 && d.NameList[0].Value == "__XGO_SKIP_TRAP" {
				skipTrap = true
				break
			}
		}
		if skipTrap {
			break
		}
	}

	type lineCol struct {
		line uint
		col  uint
	}
	// build pos -> syntax mapping
	syncDeclMapping := make(map[string]map[lineCol]*xgo_syntax.DeclInfo)
	for _, syntaxDecl := range syntaxDecls {
		pos := syntaxDecl.FuncDecl.Pos()
		file := pos.RelFilename()
		fileMapping := syncDeclMapping[file]
		if fileMapping == nil {
			fileMapping = make(map[lineCol]*xgo_syntax.DeclInfo)
			syncDeclMapping[file] = fileMapping
		}
		fileMapping[lineCol{
			line: pos.Line(),
			col:  pos.Col(),
		}] = syntaxDecl
	}

	setTrap := "__xgo_set_trap"
	linkMap := map[string]string{
		"__xgo_link_for_each_func": "__xgo_for_each_func",
		"__xgo_link_getcurg":       "__xgo_getcurg",
		"__xgo_link_set_trap":      setTrap,
	}

	// var fileNames []string
	// for _, file := range files {
	// 	fileNames = append(fileNames, file.Pos().Base().Filename())
	// }

	intf := types.Types[types.TINTER]
	intfSlice = types.NewSlice(intf)
	// printString := typecheck.LookupRuntime("printstring")
	trap := typecheck.LookupRuntime("__xgo_trap")
	forEachFunc(func(fn *ir.Func) bool {
		// TODO: if we find the func appears as one argument of trap.AddInterceptor,
		// or if its first statement is 'trap.Mark()'
		if false {
			// TODO: is it ok to parse comment here?
			fnPos := base.Ctxt.OutermostPos(fn.Pos())
			var syntaxFile *syntax.File

			fileName := fnPos.AbsFilename() // symlink may affect filename vs absFilename?
			for _, file := range files {
				syntaxFileName := file.Pos().Base().Filename()
				if syntaxFileName == fileName {
					syntaxFile = file
					break
				}
			}
			if syntaxFile != nil {
				lineNum := fnPos.Line()

				_ = lineNum
				// find matching comment
			}
		}

		// })
		// for _, fn := range typecheck.Target.Funcs {
		// NOTE: fnName is main, not main.main
		fnName := fn.Sym().Name
		// if this is a closure, skip it
		// NOTE: 'init.*' can be init function, or closure inside init functions, so they have prefix 'init.'
		if fnName == "init" || (strings.HasPrefix(fnName, "init.") && fn.OClosure == nil) {
			// the name `init` is package level auto generated init,
			// so don't trap this
			return true
		}
		// process link name
		// TODO: what about unnamed closure?
		linkName := linkMap[fnName]
		if linkName != "" {
			if !strings.HasPrefix(types.LocalPkg.Path, xgoRuntimePkgPrefix) {
				return true
			}
			// ir.Dump("before:", fn)
			if !disableXgoLink {
				if linkName == setTrap && types.LocalPkg.Path != xgoRuntimeTrapPkg {
					return true
				}
				replaceWithRuntimeCall(fn, linkName)
			}
			// ir.Dump("after:", fn)
			return true
		}
		// TODO: read comment
		if skipTrap || strings.HasPrefix(fnName, "__xgo") || strings.HasSuffix(fnName, "_xgo_trap_skip") {
			// the __xgo prefix is reserved for xgo
			return true
		}
		if disableTrap {
			return true
		}
		if base.Flag.Std {
			// skip std lib, especially skip:
			//    runtime, runtime/internal, runtime/*, reflect, unsafe, syscall, sync, sync/atomic,  internal/*
			//
			// however, there are some funcs in stdlib that we can
			// trap, for example, db connection
			// for example:
			//     errors, math, math/bits, unicode, unicode/utf8, unicode/utf16, strconv, path, sort, time, encoding/json
			return true
		}
		pkgPath := types.LocalPkg.Path
		if false {
			// skip non-main package paths?
			if pkgPath != "main" {
				return true
			}
		}
		if fn.Body == nil {
			// in go, function can have name without body
			return true
		}

		if strings.HasPrefix(pkgPath, xgoRuntimePkgPrefix) && !strings.HasPrefix(pkgPath[len(xgoRuntimePkgPrefix):], "test/") {
			// skip all packages for xgo,except test
			return true
		}

		// check if function body's first statement is a call to 'trap.Skip()'
		if isFirstStmtSkipTrap(fn.Body) {
			return true
		}

		// func marked nosplit will skip trap because
		// inserting traps when -gcflags=-N -l enabled
		// would cause stack overflow 792 bytes
		if fn.Pragma&ir.Nosplit != 0 {
			return true
		}

		// ir.Dump("before:", fn)
		// fn.Body =
		t := fn.Type()
		ft := t.FuncType()
		_ = ft

		_ = t

		/*
			equivalent go code:
			func orig(a string) error {
				something....
				return nil
			}
			==>
			func orig_trap(a string) (err error) {
				after,stop := __trap(nil,[]interface{}{&a},[]interface{}{&err})
				if stop {
				}else{
					if after!=nil{
						defer after()
					}
					something....
					return nil
				}
			}
		*/

		recv := t.Recv()
		// tParams := t.TParams()
		// if tParams != nil && tParams.NumFields() > 0 {
		n := t.NumTParams()
		_ = n
		defT := fn.Nname.Defn.Type()
		b := defT.HasTParam()
		_ = b
		pos := base.Ctxt.PosTable.Pos(fn.Pos())
		posFile := pos.AbsFilename()
		posLine := pos.Line()
		posCol := pos.Col()

		decl := syncDeclMapping[posFile][lineCol{
			line: posLine,
			col:  posCol,
		}]

		var identityName string
		var generic bool
		if decl != nil {
			identityName = decl.IdentityName()
			generic = decl.Generic
		}

		if false {
			// t.HasTParam() always returns false because
			// the funcs appears here are all instansitated
			if t.HasTParam() {
				var recvTypeName string
				var recvPtr bool
				if recv != nil {
					recvTypeName = recv.Sym.Name
					if recv.Type.IsPtr() {
						recvPtr = true
						identityName = "*"
					}
				}
				identityName = xgo_func_name.FormatFuncRefName(recvTypeName, recvPtr, fnName)
			}
		}
		afterV := typecheck.TempAt(base.AutogeneratedPos, fn, NewSignature(types.LocalPkg, nil, nil, nil, nil))
		stopV := typecheck.TempAt(base.AutogeneratedPos, fn, types.Types[types.TBOOL])

		callTrap := ir.NewCallExpr(base.AutogeneratedPos, ir.OCALL, trap, []ir.Node{
			NewStringLit(base.AutogeneratedPos, pkgPath),
			NewStringLit(base.AutogeneratedPos, identityName),
			NewBoolLit(base.AutogeneratedPos, generic),
			takeAddr(fn, recv),
			// newNilInterface(base.AutogeneratedPos),
			takeAddrs(fn, t.Params()),
			// newNilInterfaceSlice(base.AutogeneratedPos),
			takeAddrs(fn, t.Results()),
			// newNilInterfaceSlice(base.AutogeneratedPos),
		})

		callAssign := ir.NewAssignListStmt(base.AutogeneratedPos, ir.OAS2, []ir.Node{afterV, stopV}, []ir.Node{callTrap})
		callAssign.Def = true

		var assignStmt ir.Node = callAssign
		if false {
			assignStmt = callTrap
		}

		callAfter := ir.NewIfStmt(base.AutogeneratedPos, ir.NewBinaryExpr(base.AutogeneratedPos, ir.ONE, afterV, NewNilExpr(base.AutogeneratedPos, afterV.Type())), []ir.Node{
			ir.NewGoDeferStmt(base.AutogeneratedPos, ir.ODEFER, ir.NewCallExpr(base.AutogeneratedPos, ir.OCALL, afterV, nil)),
		}, nil)

		origBody := fn.Body
		newBody := make([]ir.Node, 1+len(origBody))
		newBody[0] = callAfter
		for i := 0; i < len(origBody); i++ {
			newBody[i+1] = origBody[i]
		}
		ifStmt := ir.NewIfStmt(base.AutogeneratedPos, stopV, nil, newBody)

		fn.Body = []ir.Node{assignStmt, ifStmt}
		if false {
			// fn.Body = []ir.Node{assignStmt /* ifStmt */}
		}
		typeCheckBody(fn)
		xgo_record.SetRewrittenBody(fn, fn.Body)

		// ir.Dump("after:", fn)

		return true
	})
}

func isFirstStmtSkipTrap(nodes ir.Nodes) bool {
	for _, node := range nodes {
		if isCallTo(node, xgoRuntimeTrapPkg, "Skip") {
			return true
		}
	}
	return false
}

func isCallTo(node ir.Node, pkgPath string, name string) bool {
	callNode, ok := node.(*ir.CallExpr)
	if !ok {
		return false
	}
	nameNode, ok := callNode.X.(*ir.Name)
	if !ok {
		return false
	}
	sym := nameNode.Sym()
	if sym == nil {
		return false
	}
	return sym.Pkg != nil && sym.Name == name && sym.Pkg.Path == pkgPath
}

func initRegFuncs() {
	// if types.LocalPkg.Name != "main" {
	// 	return
	// }
	sym, ok := types.LocalPkg.Syms["__xgo_register_funcs"]
	if !ok {
		return
	}
	// TODO: check sym is func, and accepts the following param
	regFunc := typecheck.LookupRuntime("__xgo_register_func")
	node := ir.NewCallExpr(base.AutogeneratedPos, ir.OCALL, sym.Def.(*ir.Name), []ir.Node{
		regFunc,
	})
	nodes := []ir.Node{node}
	typecheck.Stmts(nodes)
	prependInit(typecheck.Target, nodes)
}

func ForEachField(params *types.Type, callback func(field *types.Field) bool) {
	n := params.NumFields()
	for i := 0; i < n; i++ {
		if !callback(params.Field(i)) {
			return
		}
	}
}

func replaceWithRuntimeCall(fn *ir.Func, name string) {
	if false {
		debugReplaceBody(fn)
		// newBody = []ir.Node{debugPrint("replaced body")}
		return
	}
	runtimeFunc := typecheck.LookupRuntime(name)
	params := fn.Type().Params()
	results := fn.Type().Results()

	paramNames := getTypeNames(params)
	resNames := getTypeNames(results)

	var callNode ir.Node
	callNode = ir.NewCallExpr(base.AutogeneratedPos, ir.OCALL, runtimeFunc, paramNames)
	if len(resNames) > 0 {
		// if len(resNames) == 1 {
		// 	callNode = ir.NewAssignListStmt(base.AutogeneratedPos, ir.OAS, resNames, []ir.Node{callNode})
		// } else {
		callNode = ir.NewReturnStmt(base.AutogeneratedPos, []ir.Node{callNode})
		// callNode = ir.NewAssignListStmt(base.AutogeneratedPos, ir.OAS2, resNames, []ir.Node{callNode})

		// callNode = ir.NewAssignListStmt(base.AutogeneratedPos, ir.OAS2, resNames, []ir.Node{callNode})
		// }
	}
	var node ir.Node
	node = ifConstant(true, []ir.Node{
		// debugPrint("debug getg"),
		callNode,
	}, fn.Body)

	fn.Body = []ir.Node{node}
	xgo_record.SetRewrittenBody(fn, fn.Body)
	typeCheckBody(fn)
}

func debugReplaceBody(fn *ir.Func) {
	// debug
	if false {
		str := NewStringLit(base.AutogeneratedPos, "shit")
		nd := fn.Body[0]
		ue := nd.(*ir.UnaryExpr)
		ce := ue.X.(*ir.ConvExpr)
		ce.X = str
		xgo_record.SetRewrittenBody(fn, fn.Body)
		return
	}
	debugBody := ifConstant(true, []ir.Node{
		debugPrint("replaced body\n"),
		// ir.NewReturnStmt(base.AutogeneratedPos, nil),
	}, nil)
	// debugBody := debugPrint("replaced body\n")
	fn.Body = []ir.Node{debugBody}
	typeCheckBody(fn)
	xgo_record.SetRewrittenBody(fn, fn.Body)
}

func typeCheckBody(fn *ir.Func) {
	savedFunc := ir.CurFunc
	ir.CurFunc = fn
	typecheck.Stmts(fn.Body)
	ir.CurFunc = savedFunc
}

func ifConstant(b bool, body []ir.Node, els []ir.Node) *ir.IfStmt {
	return ir.NewIfStmt(base.AutogeneratedPos,
		NewBoolLit(base.AutogeneratedPos, true),
		body,
		els,
	)
}

func NewStringLit(pos src.XPos, s string) ir.Node {
	return NewBasicLit(pos, types.Types[types.TSTRING], constant.MakeString(s))
}
func NewBoolLit(pos src.XPos, b bool) ir.Node {
	return NewBasicLit(pos, types.Types[types.TBOOL], constant.MakeBool(b))
}

func regFuncsV1() {
	files := xgo_syntax.GetFiles()
	xgo_syntax.ClearFiles() // help GC

	type declName struct {
		name         string
		recvTypeName string
		recvPtr      bool
	}
	var declFuncNames []*declName
	for _, f := range files {
		for _, decl := range f.DeclList {
			fn, ok := decl.(*syntax.FuncDecl)
			if !ok {
				continue
			}
			if fn.Name.Value == "init" {
				continue
			}
			var recvTypeName string
			var recvPtr bool
			if fn.Recv != nil {
				if starExpr, ok := fn.Recv.Type.(*syntax.Operation); ok && starExpr.Op == syntax.Mul {
					recvTypeName = starExpr.X.(*syntax.Name).Value
					recvPtr = true
				} else {
					recvTypeName = fn.Recv.Type.(*syntax.Name).Value
				}
			}
			declFuncNames = append(declFuncNames, &declName{
				name:         fn.Name.Value,
				recvTypeName: recvTypeName,
				recvPtr:      recvPtr,
			})
		}
	}

	regFunc := typecheck.LookupRuntime("__xgo_register_func")
	regMethod := typecheck.LookupRuntime("__xgo_register_method")
	_ = regMethod

	var regNodes []ir.Node
	for _, declName := range declFuncNames {
		var valNode ir.Node
		fnSym, ok := types.LocalPkg.LookupOK(declName.name)
		if !ok {
			panic(fmt.Errorf("func name symbol not found: %s", declName.name))
		}
		if declName.recvTypeName != "" {
			typeSym, ok := types.LocalPkg.LookupOK(declName.recvTypeName)
			if !ok {
				panic(fmt.Errorf("type name symbol not found: %s", declName.recvTypeName))
			}
			var recvNode ir.Node
			if !declName.recvPtr {
				recvNode = typeSym.Def.(*ir.Name)
				// recvNode = ir.NewNameAt(base.AutogeneratedPos, typeSym, nil)
			} else {
				// types.TypeSymLookup are for things like "int","func(){...}"
				//
				// typeSym2 := types.TypeSymLookup(declName.recvTypeName)
				// if typeSym2 == nil {
				// 	panic("empty typeSym2")
				// }
				// types.TypeSym()
				recvNode = ir.TypeNode(typeSym.Def.(*ir.Name).Type())
			}
			valNode = ir.NewSelectorExpr(base.AutogeneratedPos, ir.OMETHEXPR, recvNode, fnSym)
			continue
		} else {
			valNode = fnSym.Def.(*ir.Name)
			// valNode = ir.NewNameAt(base.AutogeneratedPos, fnSym, fnSym.Def.Type())
			// continue
		}
		_ = valNode

		node := ir.NewCallExpr(base.AutogeneratedPos, ir.OCALL, regFunc, []ir.Node{
			// NewNilExpr(base.AutogeneratedPos, types.AnyType),
			ir.NewConvExpr(base.AutogeneratedPos, ir.OCONV, types.Types[types.TINTER] /*types.AnyType*/, valNode),
			// ir.NewBasicLit(base.AutogeneratedPos, types.Types[types.TSTRING], constant.MakeString("hello init\n")),
		})

		// ir.MethodExprFunc()
		regNodes = append(regNodes, node)
	}

	// this typecheck is required
	// to make subsequent steps work
	typecheck.Stmts(regNodes)

	// regFuncs.Body = []ir.Node{
	// 	ir.NewCallExpr(base.AutogeneratedPos, ir.OCALL, typecheck.LookupRuntime("printstring"), []ir.Node{
	// 		ir.NewBasicLit(base.AutogeneratedPos, types.Types[types.TSTRING], constant.MakeString("hello init\n")),
	// 	}),
	// }
	prependInit(typecheck.Target, regNodes)
}

// how to delcare a new function?
// init names are usually init.0, init.1, ...
//
// NOTE: when there is already an init function, declare new init function
// will give an error: main..inittask: relocation target main.init.1 not defined
func prependInit(target *ir.Package, body []ir.Node) {
	if len(target.Inits) > 0 {
		target.Inits[0].Body.Prepend(body...)
		return
	}

	sym := types.LocalPkg.Lookup(fmt.Sprintf("init.%d", len(target.Inits)))
	regFunc := NewFunc(base.AutogeneratedPos, base.AutogeneratedPos, sym, NewSignature(types.LocalPkg, nil, nil, nil, nil))
	regFunc.Body = body

	target.Inits = append(target.Inits, regFunc)
	AddFuncs(regFunc)
}

func takeAddr(fn *ir.Func, field *types.Field) ir.Node {
	if field == nil {
		return newNilInterface(base.AutogeneratedPos)
	}
	// go1.20 only? no Nname, so cannot refer to it
	// but we should display it in trace?(better to do so)
	if field.Nname == nil {
		return newNilInterface(base.AutogeneratedPos)
	}
	// if name is "_", return nil
	if field.Sym != nil && field.Sym.Name == "_" {
		return newNilInterface(base.AutogeneratedPos)
	}
	arg := ir.NewAddrExpr(base.AutogeneratedPos, field.Nname.(*ir.Name))
	conv := ir.NewConvExpr(base.AutogeneratedPos, ir.OCONV, types.Types[types.TINTER], arg)
	conv.SetImplicit(true)
	// go1.20: must have Typeword
	conv.TypeWord = reflectdata.TypePtrAt(base.Pos, types.NewPtr(field.Type))
	return conv
}

func newNilInterface(pos src.XPos) ir.Expr {
	return NewNilExpr(pos, types.Types[types.TINTER])
}
func newNilInterfaceSlice(pos src.XPos) ir.Expr {
	return NewNilExpr(pos, types.NewSlice(types.Types[types.TINTER]))
}
