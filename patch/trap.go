package patch

import (
	"fmt"
	"os"
	"strings"

	"cmd/compile/internal/base"
	"cmd/compile/internal/ir"
	"cmd/compile/internal/typecheck"
	"cmd/compile/internal/types"

	xgo_ctxt "cmd/compile/internal/xgo_rewrite_internal/patch/ctxt"
	xgo_record "cmd/compile/internal/xgo_rewrite_internal/patch/record"
	xgo_syntax "cmd/compile/internal/xgo_rewrite_internal/patch/syntax"
)

const sig_expected__xgo_trap = `func(pkgPath string, identityName string, generic bool, recv interface{}, args []interface{}, results []interface{}) (func(), bool)`

func init() {
	if sig_gen__xgo_trap != sig_expected__xgo_trap {
		panic(fmt.Errorf("__xgo_trap signature changed, run go generate and update sig_expected__xgo_trap correspondly"))
	}
	if goMajor != 1 {
		panic(fmt.Errorf("expect goMajor to be 1, actual:%d", goMajor))
	}

	if goMinor < 17 || goMinor > 22 {
		panic(fmt.Errorf("expect goMinor 17~22, actual:%d", goMinor))
	}
}

const disableTrap bool = false

func Patch() {
	debugIR()
	if os.Getenv("COMPILER_ALLOW_IR_REWRITE") != "true" {
		return
	}
	insertTrapPoints()
	initRegFuncs()
	initClosureRegs()
}

var inited bool
var intfSlice *types.Type
var strSlice *types.Type

func ensureInit() {
	if inited {
		return
	}
	inited = true

	intf := types.Types[types.TINTER]
	intfSlice = types.NewSlice(intf)
	strSlice = types.NewSlice(types.Types[types.TSTRING])
}

func insertTrapPoints() {
	ensureInit()

	// cleanup upon exit
	defer func() {
		xgo_syntax.ClearSyntaxDeclMapping()

		// TODO: check if symlink may affect filename compared to absFilename?
		xgo_syntax.ClearFiles() // help GC
		xgo_syntax.ClearDecls()
	}()

	// printString := typecheck.LookupRuntime("printstring")
	forEachFunc(func(fn *ir.Func) bool {
		trapOrLink(fn)
		return true
	})
}

func trapOrLink(fn *ir.Func) {
	linkName, insertTrap := CanInsertTrapOrLink(fn)
	if linkName != "" {
		replaceWithRuntimeCall(fn, linkName)
		return
	}
	if !insertTrap {
		return
	}
	fnName := fn.Sym().Name
	if fnName != "__debug_ir_rewrite" {
		if !InsertTrapForFunc(fn, false) {
			return
		}
	} else {
		fn.Body = []ir.Node{ir.NewReturnStmt(fn.Pos(), []ir.Node{
			GetFieldIndex(fn.Type().Params(), 0).Nname.(*ir.Name),
		})}
	}

	typeCheckBody(fn)
	xgo_record.SetRewrittenBody(fn, fn.Body)

	// ir.Dump("after:", fn)
}

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

// for go1.17,go1.18
//
//	fn.Sym().Pkg.Path
//
// returns empty
const hasFuncPkgPath = goMajor > 1 || (goMajor == 1 && goMinor >= 19)

var trappedClosures []*ir.Func

func InsertTrapForFunc(fn *ir.Func, forGeneric bool) bool {
	ensureInit()

	var fnPkg *types.Pkg
	fnSym := fn.Sym()
	if fnSym != nil {
		fnPkg = fnSym.Pkg
	}

	// not local function, so instantiated
	// generics, currently not supported
	// TODO: solve the generic function instantiation
	if fnPkg != nil && fnPkg != types.LocalPkg {
		return false
	}

	if hasFuncPkgPath {
		curPkgPath := xgo_ctxt.GetPkgPath()
		var fnPkgPath string
		if fnPkg != nil {
			fnPkgPath = fnPkg.Path
		}
		// when package are not the same,
		// do not insert trap points
		// TODO: solve the generic package issues by
		// using generic metadata
		if fnPkgPath == "" || fnPkgPath != curPkgPath {
			return false
		}
	}

	pos := getPosInfo(fn.Pos())
	posFile := pos.AbsFilename()
	posLine := pos.Line()
	posCol := pos.Col()

	syncDeclMapping := xgo_syntax.GetSyntaxDeclMapping()

	fileMapping := syncDeclMapping[posFile]
	if fileMapping == nil {
		adjustedFile := getAdjustedFile(posFile)
		if adjustedFile != "" {
			fileMapping = syncDeclMapping[adjustedFile]
		}
	}
	decl := fileMapping[xgo_syntax.LineCol{
		Line: posLine,
		Col:  posCol,
	}]

	var isClosure bool
	// no identity name
	if decl == nil {
		if fn.OClosure == nil {
			return false
		}
		isClosure = true
	} else if decl.Interface {
		// interface just name
		return false
	}

	var identityName string
	var generic bool
	if decl != nil {
		identityName = decl.IdentityName()
		generic = decl.Generic
	}

	if !isClosure && identityName == "" {
		return false
	}
	if genericTrapNeedsWorkaround && generic != forGeneric {
		return false
	}

	pkgPath := xgo_ctxt.GetPkgPath()
	trap := typecheck.LookupRuntime("__xgo_trap")
	fnPos := fn.Pos()
	fnType := fn.Type()
	afterV := typecheck.TempAt(fnPos, fn, NewSignature(types.LocalPkg, nil, nil, nil, nil))
	stopV := typecheck.TempAt(fnPos, fn, types.Types[types.TBOOL])

	recv := fnType.Recv()
	callTrap := ir.NewCallExpr(fnPos, ir.OCALL, trap, []ir.Node{
		NewStringLit(fnPos, pkgPath),
		NewStringLit(fnPos, identityName),
		NewBoolLit(fnPos, generic),
		takeAddr(fn, recv, forGeneric),
		// newNilInterface(fnPos),
		takeAddrs(fn, fnType.Params(), forGeneric),
		// newNilInterfaceSlice(fnPos),
		takeAddrs(fn, fnType.Results(), forGeneric),
		// newNilInterfaceSlice(fnPos),
	})
	if genericTrapNeedsWorkaround && forGeneric {
		callTrap.SetType(getFuncResultsType(trap.Type()))
	}

	callAssign := ir.NewAssignListStmt(fnPos, ir.OAS2, []ir.Node{afterV, stopV}, []ir.Node{callTrap})
	callAssign.Def = true

	var assignStmt ir.Node = callAssign
	if false {
		assignStmt = callTrap
	}

	bin := ir.NewBinaryExpr(fnPos, ir.ONE, afterV, NewNilExpr(fnPos, afterV.Type()))
	if forGeneric {
		// only generic needs explicit type
		bin.SetType(types.Types[types.TBOOL])
	}

	callAfter := ir.NewIfStmt(fnPos, bin, []ir.Node{
		ir.NewGoDeferStmt(fnPos, ir.ODEFER, ir.NewCallExpr(fnPos, ir.OCALL, afterV, nil)),
	}, nil)

	origBody := fn.Body
	newBody := make([]ir.Node, 1+len(origBody))
	newBody[0] = callAfter
	for i := 0; i < len(origBody); i++ {
		newBody[i+1] = origBody[i]
	}
	ifStmt := ir.NewIfStmt(fnPos, stopV, nil, newBody)

	if isClosure {
		trappedClosures = append(trappedClosures, fn)
	}
	fn.Body = []ir.Node{assignStmt, ifStmt}
	return true
}

func CanInsertTrapOrLink(fn *ir.Func) (string, bool) {
	pkgPath := xgo_ctxt.GetPkgPath()
	// for _, fn := range typecheck.Target.Funcs {
	// NOTE: fnName is main, not main.main
	fnName := fn.Sym().Name
	// if this is a closure, skip it
	// NOTE: 'init.*' can be init function, or closure inside init functions, so they have prefix 'init.'
	if fnName == "init" || (strings.HasPrefix(fnName, "init.") && fn.OClosure == nil) {
		// the name `init` is package level auto generated init,
		// so don't trap this
		return "", false
	}
	// process link name
	// TODO: what about unnamed closure?
	linkName := linkMap[fnName]
	if linkName != "" {
		// ir.Dump("before:", fn)
		if !isLinkValid(fnName, linkName, pkgPath) {
			return "", false
		}
		return linkName, false
		// ir.Dump("after:", fn)
	}
	if xgo_ctxt.SkipPackageTrap() {
		return "", false
	}
	// TODO: read comment
	if xgo_syntax.HasSkipTrap() || strings.HasPrefix(fnName, "__xgo") || strings.HasSuffix(fnName, "_xgo_trap_skip") {
		// the __xgo prefix is reserved for xgo
		return "", false
	}
	if disableTrap {
		return "", false
	}
	if base.Flag.Std {
		// skip std lib, especially skip:
		//    runtime, runtime/internal, runtime/*, reflect, unsafe, syscall, sync, sync/atomic,  internal/*
		//
		// however, there are some funcs in stdlib that we can
		// trap, for example, db connection
		// for example:
		//     errors, math, math/bits, unicode, unicode/utf8, unicode/utf16, strconv, path, sort, time, encoding/json

		// NOTE: base.Flag.Std in does not always reflect func's package path,
		// because generic instantiation happens in other package, so this
		// func may be a foreigner.
		return "", false
	}
	if !canInsertTrap(fn) {
		return "", false
	}
	if false {
		// skip non-main package paths?
		if pkgPath != "main" {
			return "", false
		}
	}
	if fn.Body == nil {
		// in go, function can have name without body
		return "", false
	}

	// skip all packages for xgo,except test
	if strings.HasPrefix(pkgPath, xgoRuntimePkgPrefix) {
		remain := pkgPath[len(xgoRuntimePkgPrefix):]
		if !strings.HasPrefix(remain, "test/") && !strings.HasPrefix(remain, "runtime/test/") {
			return "", false
		}
	}

	// check if function body's first statement is a call to 'trap.Skip()'
	if isFirstStmtSkipTrap(fn.Body) {
		return "", false
	}

	// func marked nosplit will skip trap because
	// inserting traps when -gcflags=-N -l enabled
	// would cause stack overflow 792 bytes
	if fn.Pragma&ir.Nosplit != 0 {
		return "", false
	}
	return "", true
}

// Deprecated, there is no __xgo_register_funcs any more,
// so this function is useless
func initRegFuncs() {
	// if types.LocalPkg.Name != "main" {
	// 	return
	// }
	sym, ok := types.LocalPkg.Syms["__xgo_register_funcs"]
	if !ok {
		return
	}
	symDef := sym.Def.(*ir.Name)
	pos := symDef.Pos()
	// TODO: check sym is func, and accepts the following param
	regFunc := typecheck.LookupRuntime("__xgo_register_func")
	node := ir.NewCallExpr(pos, ir.OCALL, symDef, []ir.Node{
		regFunc,
	})
	nodes := []ir.Node{node}
	typecheck.Stmts(nodes)
	prependInit(pos, typecheck.Target, nodes)
}

func initClosureRegs() {
	if len(trappedClosures) == 0 {
		return
	}
	pkgPath := xgo_ctxt.GetPkgPath()
	locReg := types.LocalPkg.Lookup("__xgo_local_register_func")
	locRegFn := locReg.Def.(*ir.Name)

	initBody := types.LocalPkg.Lookup("__xgo_link_generate_init_regs_body")
	initBodyFn := initBody.Def.(*ir.Name)

	registerNodes := make([]ir.Node, 0, len(trappedClosures))

	pkgPathLit := NewStringLit(base.AutogeneratedPos, pkgPath)
	trueLit := NewBoolLit(base.AutogeneratedPos, true)
	emptyStrLit := NewStringLit(base.AutogeneratedPos, "")
	for _, fn := range trappedClosures {
		pos := fn.Pos()

		posInfo := getPosInfo(pos)
		fileName := posInfo.AbsFilename()
		adjustFile := getAdjustedFile(fileName)
		if adjustFile != "" {
			fileName = adjustFile
		}
		fnType := fn.Type()
		argNames := getFieldNames(fn, fnType.Params())
		resNames := getFieldNames(fn, fnType.Results())
		// TODO: why fn.Name() always returns nil?
		var fnRef ir.Node = convToEFace(pos, fn.Nname, fn.Type(), false)
		if false {
			// debug
			fnRef = newNilInterface(pos)
		}
		if false {
			// NOTE: cannot do this, must use convToEFace
			fnRef = fn.Nname
		}
		registerNode := ir.NewCallExpr(pos, ir.OCALL, locRegFn, []ir.Node{
			pkgPathLit,
			fnRef,
			trueLit,     // is closure
			emptyStrLit, // no recv name
			argNames,
			resNames,
			NewStringLit(pos, fileName),         //file
			NewIntLit(pos, int(posInfo.Line())), //line
		})
		registerNodes = append(registerNodes, registerNode)
	}

	// modify existing body
	if false {
		// debug
		debugReplaceBody(initBodyFn.Func)
		return
	}
	replaceFuncBody(initBodyFn.Func, registerNodes)
}

// for go1.20 and above, needs to convert
const needConvertArg = goMajor > 1 || (goMajor == 1 && goMinor >= 20)

func replaceWithRuntimeCall(fn *ir.Func, name string) {
	if false {
		debugReplaceBody(fn)
		// newBody = []ir.Node{debugPrint("replaced body")}
		return
	}
	isRuntime := true
	var runtimeFunc *ir.Name
	if isRuntime {
		runtimeFunc = typecheck.LookupRuntime(name)
	} else {
		// NOTE: cannot reference testing package
		// only runtime is available
		// lookup testing
		testingPkg := findTestingPkg()
		sym := testingPkg.Lookup(name)
		if sym.Def != nil {
			runtimeFunc = sym.Def.(*ir.Name)
		} else {
			runtimeFunc = NewNameAt(fn.Pos(), sym, fn.Type())
			runtimeFunc.Class = ir.PEXTERN
		}
	}
	params := fn.Type().Params()
	results := fn.Type().Results()

	paramNames := getTypeNames(params)
	resNames := getTypeNames(results)
	fnPos := fn.Pos()

	if needConvertArg && name == xgoOnTestStart {
		for i, p := range paramNames {
			paramNames[i] = convToEFace(fnPos, p, p.(*ir.Name).Type(), false)
		}
	}

	var callNode ir.Node
	callNode = ir.NewCallExpr(fnPos, ir.OCALL, runtimeFunc, paramNames)
	if len(resNames) > 0 {
		// if len(resNames) == 1 {
		// 	callNode = ir.NewAssignListStmt(fnPos, ir.OAS, resNames, []ir.Node{callNode})
		// } else {
		callNode = ir.NewReturnStmt(fnPos, []ir.Node{callNode})
		// callNode = ir.NewAssignListStmt(fnPos, ir.OAS2, resNames, []ir.Node{callNode})

		// callNode = ir.NewAssignListStmt(fnPos, ir.OAS2, resNames, []ir.Node{callNode})
		// }
	}
	replaceFuncBody(fn, []ir.Node{
		// debugPrint("debug getg"),
		callNode,
	})
}

func replaceFuncBody(fn *ir.Func, nodes []ir.Node) {
	node := ifConstant(fn.Pos(), true, nodes, fn.Body)

	fn.Body = []ir.Node{node}
	xgo_record.SetRewrittenBody(fn, fn.Body)
	typeCheckBody(fn)
}

var testingPkg *types.Pkg

func findTestingPkg() *types.Pkg {
	if testingPkg != nil {
		return testingPkg
	}
	for _, pkg := range typecheck.Target.Imports {
		if pkg.Path == "testing" {
			testingPkg = pkg
			return pkg
		}
	}
	panic("testing package not imported")
}
func debugReplaceBody(fn *ir.Func) {
	// debug
	if false {
		str := NewStringLit(fn.Pos(), "debug")
		nd := fn.Body[0]
		ue := nd.(*ir.UnaryExpr)
		ce := ue.X.(*ir.ConvExpr)
		ce.X = str
		xgo_record.SetRewrittenBody(fn, fn.Body)
		return
	}
	if false {
		fn.Body = []ir.Node{
			debugPrint("replaced body x\n"),
		}
		typeCheckBody(fn)
		xgo_record.SetRewrittenBody(fn, fn.Body)
		return
	}
	debugBody := ifConstant(fn.Pos(), true, []ir.Node{
		debugPrint("replaced body 1\n"),
		debugPrint("replaced body 2\n"),
		ir.NewReturnStmt(base.AutogeneratedPos, nil),
		fn.Body[0],
		debugPrint("replaced body 3\n"),
		// ir.NewReturnStmt(fn.Pos(), nil),
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
