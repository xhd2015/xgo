package syntax

import (
	"cmd/compile/internal/base"
	"cmd/compile/internal/syntax"
	"cmd/compile/internal/xgo_rewrite_internal/patch/info"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	xgo_ctxt "cmd/compile/internal/xgo_rewrite_internal/patch/ctxt"
)

type DeclInfo = info.DeclInfo
type DeclKind = info.DeclKind

const Kind_Func = info.Kind_Func
const Kind_Var = info.Kind_Var
const Kind_VarPtr = info.Kind_VarPtr
const Kind_Const = info.Kind_Const

const XGO_TOOLCHAIN_VERSION = "XGO_TOOLCHAIN_VERSION"
const XGO_TOOLCHAIN_REVISION = "XGO_TOOLCHAIN_REVISION"
const XGO_TOOLCHAIN_VERSION_NUMBER = "XGO_TOOLCHAIN_VERSION_NUMBER"

const XGO_VERSION = "XGO_VERSION"
const XGO_REVISION = "XGO_REVISION"
const XGO_NUMBER = "XGO_NUMBER"

// --strace
const XGO_STACK_TRACE = "XGO_STACK_TRACE"
const XGO_STD_LIB_TRAP_DEFAULT_ALLOW = "XGO_STD_LIB_TRAP_DEFAULT_ALLOW"
const straceFlagConstName = "__xgo_injected_StraceFlag"
const trapStdlibFlagConstName = "__xgo_injected_StdlibTrapDefaultAllow"

// this link function is considered safe as we do not allow user
// to define such one,there will be no abuse
const XgoLinkGeneratedRegisterFunc = "__xgo_link_generated_register_func"
const XgoRegisterFuncs = "__xgo_register_funcs"
const XgoLocalFuncStub = "__xgo_local_func_stub"
const XgoLocalPkgName = "__xgo_local_pkg_name"

const sig_expected__xgo_register_func = `func(info interface{})`

func init() {
	if sig_gen__xgo_register_func != sig_expected__xgo_register_func {
		panic(fmt.Errorf("__xgo_register_func signature changed, run go generate and update sig_expected__xgo_register_func correspondly"))
	}
}

func AfterFilesParsed(fileList []*syntax.File, addFile func(name string, r io.Reader) *syntax.File) {
	xgo_ctxt.InitAfterLoad()
	defer xgo_ctxt.LogSpan("AfterFilesParsed")()
	if !xgo_ctxt.XGO_COMPILER_ENABLE_SYNTAX {
		return
	}
	debugSyntax(fileList)
	if !xgo_ctxt.XGO_COMPILER_SYNTAX_SKIP_INJECT_XGO_FLAGS {
		injectXgoFlags(fileList)
	}
	if !xgo_ctxt.XGO_COMPILER_SYNTAX_SKIP_FILL_FUNC_NAMES {
		fillFuncArgResNames(fileList)
	}
	registerAndTrapFuncs(fileList, addFile)
}

// typeinfo not used
// func AfterSyntaxTypeCheck(pkgPath string, files []*syntax.File, info *types2.Info) {
// 	if pkgPath != "github.com/xhd2015/xgo/runtime/test/debug" {
// 		return
// 	}
// 	if true {
// 		return
// 	}
// 	stmt := files[0].DeclList[2].(*syntax.FuncDecl).Body.List[0]
// 	call := stmt.(*syntax.ExprStmt).X.(*syntax.CallExpr)
// 	name := call.ArgList[0].(*syntax.Name)
// 	if false {
// 		v := &syntax.BasicLit{Value: "11", Kind: syntax.IntLit}
// 		t := syntax.TypeAndValue{
// 			Type:  name.GetTypeInfo().Type,
// 			Value: constant.MakeInt64(11),
// 		}
// 		t.SetIsValue()
// 		v.SetTypeInfo(t)
// 		call.ArgList[0] = v
// 	}

// 	_ = name
// }

func debugPkgSyntax(files []*syntax.File) {
	if false {
		return
	}
	pkgPath := xgo_ctxt.GetPkgPath()
	if pkgPath != "github.com/xhd2015/xgo/runtime/test/debug" {
		return
	}

	stmt := files[0].DeclList[2].(*syntax.FuncDecl).Body.List[1]
	call := stmt.(*syntax.ExprStmt).X.(*syntax.CallExpr)
	name := call.ArgList[0].(*syntax.Name)
	// if false {
	call.ArgList[0] = &syntax.XgoSimpleConvert{
		X: &syntax.CallExpr{
			Fun: syntax.NewName(name.Pos(), "int"),
			ArgList: []syntax.Expr{
				name,
			},
		},
	}
	// }
}

func GetSyntaxDeclMapping() map[string]map[LineCol]*info.DeclInfo {
	return getSyntaxDeclMapping()
}

var allFiles []*syntax.File
var allDecls []*info.DeclInfo

func ClearFiles() {
	allFiles = nil
}

// not used anywhere
func GetFiles() []*syntax.File {
	return allFiles
}

func GetDecls() []*DeclInfo {
	return allDecls
}

func ClearDecls() {
	allDecls = nil
}

type LineCol struct {
	Line uint
	Col  uint
}

var syntaxDeclMapping map[string]map[LineCol]*DeclInfo

func getSyntaxDeclMapping() map[string]map[LineCol]*DeclInfo {
	if syntaxDeclMapping != nil {
		return syntaxDeclMapping
	}
	// build pos -> syntax mapping
	syntaxDeclMapping = make(map[string]map[LineCol]*DeclInfo)
	for _, syntaxDecl := range allDecls {
		if syntaxDecl.Interface {
			continue
		}
		if !syntaxDecl.Kind.IsFunc() {
			continue
		}
		file := syntaxDecl.File
		fileMapping := syntaxDeclMapping[file]
		if fileMapping == nil {
			fileMapping = make(map[LineCol]*DeclInfo)
			syntaxDeclMapping[file] = fileMapping
		}
		fileMapping[LineCol{
			Line: uint(syntaxDecl.Line),
			Col:  syntaxDecl.FuncDecl.Pos().Col(),
		}] = syntaxDecl
	}
	return syntaxDeclMapping
}

var computedSkipTrap bool
var skipTrap bool

func HasSkipTrap() bool {
	if computedSkipTrap {
		return skipTrap
	}
	computedSkipTrap = true
	skipTrap = computeSkipTrap(allFiles)
	return skipTrap
}

func computeSkipTrap(files []*syntax.File) bool {
	// check if any file has __XGO_SKIP_TRAP
	for _, f := range files {
		for _, d := range f.DeclList {
			if d, ok := d.(*syntax.ConstDecl); ok && len(d.NameList) > 0 && d.NameList[0].Value == "__XGO_SKIP_TRAP" {
				return true
			}
		}
	}
	return false
}

func ClearSyntaxDeclMapping() {
	syntaxDeclMapping = nil
}

func registerAndTrapFuncs(fileList []*syntax.File, addFile func(name string, r io.Reader) *syntax.File) {
	defer xgo_ctxt.LogSpan("registerAndTrapFuncs")()
	allFiles = fileList

	pkgPath := xgo_ctxt.GetPkgPath()

	var needTimePatch bool
	var needTimeRewrite bool
	if base.Flag.Std && pkgPath == "time" {
		needTimePatch = true
	} else if pkgPath == xgo_ctxt.XgoRuntimeTracePkg {
		needTimeRewrite = true
	}

	skipTrap := xgo_ctxt.XGO_COMPILER_SYNTAX_SKIP_ALL_TRAP || xgo_ctxt.SkipPackageTrap()

	// debugPkgSyntax(fileList)
	// if true {
	// 	return
	// }
	// cannot directly import the runtime package
	// but we can first:
	//  1.modify the importcfg
	//  2.do not import anything, rely on IR to finish remaining steps
	//
	// I feel the second is more proper as importcfg is an extra layer of
	// complexity, and runtime can be compiled or cached, we cannot locate
	// where its _pkg_.a is.

	varTrap := !xgo_ctxt.XGO_COMPILER_SYNTAX_SKIP_VAR_TRAP && allowVarTrap()
	var funcDelcs []*info.DeclInfo
	if needTimePatch || needTimeRewrite || !skipTrap {
		funcDelcs = getFuncDecls(fileList, varTrap)
	}
	if needTimePatch {
		// add time.Now_Xgo_Original() and time.Since_Xgo_Original()
		addTimePatch(funcDelcs)
	}
	if needTimeRewrite {
		rewriteTimePatch(funcDelcs)
	}

	if skipTrap {
		return
	}

	for _, funcDecl := range funcDelcs {
		if funcDecl.RecvTypeName == "" && funcDecl.Name == XgoLinkGeneratedRegisterFunc {
			// ensure we are safe
			// refuse to link such volatile package
			return
		}
	}

	var pkgName string
	if len(fileList) > 0 {
		pkgName = fileList[0].PkgName.Value
	}

	// filterFuncDecls
	// NOTE: stdlib is only available via source rewrite
	// IR is turned off.
	// so closure is not rewritten in stdlib
	funcDelcs = filterFuncDecls(funcDelcs, pkgPath)
	// assign to global
	allDecls = funcDelcs

	// std lib, and generic functions
	// normal functions uses IR
	rewriteFuncsSource(funcDelcs, pkgPath)

	if varTrap {
		// write package data
		err := writePkgData(pkgPath, funcDelcs)
		if err != nil {
			base.Fatalf("write pkg data: %v", err)
		}
		trapVariables(fileList, funcDelcs)
		// debug
		// fmt.Fprintf(os.Stderr, "ast:")
		// syntax.Fdump(os.Stderr, fileList[0])
	}

	// always generate a helper to aid IR
	helperFile := addFile("__xgo_autogen_register_func_helper.go", strings.NewReader(generateRegHelperCode(pkgName)))

	// change __xgo_local_pkg_name
	for _, decl := range helperFile.DeclList {
		if constDecl, ok := decl.(*syntax.ConstDecl); ok && constDecl.NameList[0].Value == XgoLocalPkgName {
			constDecl.Values.(*syntax.BasicLit).Value = strconv.Quote(pkgPath)
			break
		}
	}
	// split fileDecls to a list of batch
	// when statements gets large, it will
	// exceeds the compiler's threshold, causing
	//     internal compiler error: NewBulk too big
	// see https://github.com/golang/go/issues/33437
	// see also: https://github.com/golang/go/issues/57832 The input code is just too big for the compiler to handle.
	// here we split the files per 1000 functions
	if !xgo_ctxt.XGO_COMPILER_SYNTAX_SKIP_GEN_CODE {
		batchFuncDecls := splitBatch(funcDelcs, 1000)
		logBatchGen := xgo_ctxt.LogSpan("batch gen")
		for i, funcDecls := range batchFuncDecls {
			if len(funcDecls) == 0 {
				continue
			}
			// NOTE: here the trick is to use init across multiple files,
			// in go, init can appear more than once even in single file
			fileNameBase := "__xgo_autogen_register_func_info"
			if len(batchFuncDecls) > 1 {
				// special when there are multiple files
				fileNameBase += fmt.Sprintf("_%d", i)
			}
			initFile := addFile(fileNameBase+".go", strings.NewReader(generateRegFileCode(pkgName, "init", "")))

			initFn := initFile.DeclList[0].(*syntax.FuncDecl)
			pos := initFn.Pos()
			// use XgoLinkRegFunc for general purepose
			body := generateFuncRegBody(pos, funcDecls, XgoLinkGeneratedRegisterFunc, XgoLocalFuncStub)

			for _, stmt := range body {
				fillPos(pos, stmt)
			}
			initFn.Body.List = append(initFn.Body.List, body...)
		}
		logBatchGen()
	}
}

func injectXgoFlags(fileList []*syntax.File) {
	pkgPath := xgo_ctxt.GetPkgPath()
	switch pkgPath {
	case xgo_ctxt.XgoRuntimeCorePkg:
		patchXgoRuntimeCoreVersions(fileList)
	case xgo_ctxt.XgoRuntimeTracePkg:
		injectXgoStraceFlag(fileList)
	}
}

func findFile(fileList []*syntax.File, name string) *syntax.File {
	n := len(name)
	for _, file := range fileList {
		relFileName := file.Pos().RelFilename()
		if !strings.HasSuffix(relFileName, name) {
			continue
		}
		fn := len(relFileName)
		if fn == n {
			return file
		}
		c := relFileName[fn-n-1]

		// must be a separator
		if c == '/' || c == '\\' || c == filepath.Separator {
			return file
		}
	}
	return nil
}

func patchXgoRuntimeCoreVersions(fileList []*syntax.File) {
	version := os.Getenv(XGO_TOOLCHAIN_VERSION)
	if version == "" {
		return
	}
	revision := os.Getenv(XGO_TOOLCHAIN_REVISION)
	if revision == "" {
		return
	}
	versionNumEnv := os.Getenv(XGO_TOOLCHAIN_VERSION_NUMBER)
	versionNum, err := strconv.ParseInt(versionNumEnv, 10, 64)
	if err != nil || versionNum <= 0 {
		return
	}

	versionFile := findFile(fileList, "version.go")
	if versionFile == nil {
		return
	}
	forEachConst(versionFile.DeclList, func(constDecl *syntax.ConstDecl) bool {
		for _, name := range constDecl.NameList {
			switch name.Value {
			case XGO_VERSION:
				constDecl.Values = newStringLit(version)
			case XGO_REVISION:
				constDecl.Values = newStringLit(revision)
			case XGO_NUMBER:
				constDecl.Values = newIntLit(int(versionNum))
			}
		}
		return false
	})
}
func forEachConst(declList []syntax.Decl, fn func(constDecl *syntax.ConstDecl) bool) {
	for _, decl := range declList {
		constDecl, ok := decl.(*syntax.ConstDecl)
		if !ok {
			continue
		}
		if fn(constDecl) {
			return
		}
	}
}

func injectXgoStraceFlag(fileList []*syntax.File) {
	straceFlag := os.Getenv(XGO_STACK_TRACE)
	trapStdlibFlag := os.Getenv(XGO_STD_LIB_TRAP_DEFAULT_ALLOW)
	if straceFlag == "" && trapStdlibFlag == "" {
		return
	}

	traceFile := findFile(fileList, "trace.go")
	if traceFile == nil {
		return
	}

	forEachConst(traceFile.DeclList, func(constDecl *syntax.ConstDecl) bool {
		for _, name := range constDecl.NameList {
			switch name.Value {
			case straceFlagConstName:
				constDecl.Values = newStringLit(straceFlag)
			case trapStdlibFlagConstName:
				constDecl.Values = newStringLit(trapStdlibFlag)
			}
		}
		return false
	})
}

func getFileIndexMapping(files []*syntax.File) map[*syntax.File]int {
	m := make(map[*syntax.File]int, len(files))
	for i, file := range files {
		m[file] = i
	}
	return m
}
func splitBatch(funcDecls []*DeclInfo, batch int) [][]*DeclInfo {
	if batch <= 0 {
		panic("invalid batch")
	}
	n := len(funcDecls)
	if n <= batch {
		// fast path
		return [][]*DeclInfo{funcDecls}
	}
	var res [][]*DeclInfo

	var cur []*DeclInfo
	for i := 0; i < n; i++ {
		cur = append(cur, funcDecls[i])
		if len(cur) >= batch {
			res = append(res, cur)
			cur = nil
		}
	}
	if len(cur) > 0 {
		res = append(res, cur)
		cur = nil
	}
	return res
}

type FileDecl struct {
	File  *syntax.File
	Funcs []*DeclInfo
}

func fillMissingArgNames(fn *syntax.FuncDecl) {
	if fn.Recv != nil {
		fillName(fn.Recv, "__xgo_recv_auto_filled")
	}
	for i, p := range fn.Type.ParamList {
		fillName(p, fmt.Sprintf("__xgo_arg_auto_filled_%d", i))
	}
}

func fillName(field *syntax.Field, namePrefix string) {
	if field.Name == nil {
		field.Name = syntax.NewName(field.Pos(), namePrefix)
		return
	}
	if field.Name.Value == "_" {
		field.Name.Value = namePrefix + "_blank"
		return
	}
}

// collect funcs from files, register each of them by
// calling to __xgo_reg_func with names and func pointer

var AbsFilename func(name string) string
var TrimFilename func(b *syntax.PosBase) string

func getFuncDecls(files []*syntax.File, varTrap bool) []*info.DeclInfo {
	// fileInfos := make([]*FileDecl, 0, len(files))
	var declFuncs []*DeclInfo
	for i, f := range files {
		var file string
		var trimmed bool
		if base.Flag.Std && false {
			file = f.Pos().RelFilename()
		} else if TrimFilename != nil {
			// >= go1.18
			file = TrimFilename(f.Pos().Base())
			trimmed = true
		} else if AbsFilename != nil {
			file = AbsFilename(f.Pos().Base().Filename())
			trimmed = true
		} else {
			// fallback to default
			file = f.Pos().RelFilename()
		}

		// see https://github.com/xhd2015/xgo/issues/80
		if trimmed && strings.HasPrefix(file, "_cgo") {
			file = f.Pos().RelFilename()
		}
		for _, decl := range f.DeclList {
			fnDecls := extractFuncDecls(i, f, file, decl, varTrap)
			declFuncs = append(declFuncs, fnDecls...)
		}
	}
	// compute __xgo_trap_xxx
	n := len(declFuncs)
	if n == 0 {
		return nil
	}
	j := n
	for i := n - 1; i >= 0; i-- {
		if i == 0 || !isTrapped(declFuncs, i) {
			j--
			declFuncs[j] = declFuncs[i]
			continue
		}
		// a special comment
		declFuncs[i].FollowingTrapConst = true
		j--
		declFuncs[j] = declFuncs[i]
		// remove the comment by skipping next
		i--
	}
	return declFuncs[j:]
}

func isTrapped(declFuncs []*DeclInfo, i int) bool {
	fn := declFuncs[i]
	if fn.Kind != info.Kind_Var {
		return false
	}
	last := declFuncs[i-1]
	if last.Kind != info.Kind_Const {
		return false
	}
	const xgoTrapPrefix = "__xgo_trap_"
	if !strings.HasPrefix(last.Name, xgoTrapPrefix) {
		return false
	}
	subName := last.Name[len(xgoTrapPrefix):]
	if !strings.EqualFold(fn.Name, subName) {
		return false
	}
	return true
}

func filterFuncDecls(funcDecls []*info.DeclInfo, pkgPath string) []*info.DeclInfo {
	n := len(funcDecls)
	i := 0
	for j := 0; j < n; j++ {
		fn := funcDecls[j]

		action := xgo_ctxt.GetAction(fn)
		if action == "" {
			// disable part of stdlibs
			if !xgo_ctxt.AllowPkgFuncTrap(pkgPath, base.Flag.Std, fn.IdentityName()) {
				action = "exclude"
			}
		}
		if action == "" || action == "include" {
			funcDecls[i] = fn
			i++
		}
	}
	return funcDecls[:i]
}

func extractFuncDecls(fileIndex int, f *syntax.File, file string, decl syntax.Decl, varTrap bool) []*DeclInfo {
	switch decl := decl.(type) {
	case *syntax.FuncDecl:
		info := getFuncDeclInfo(fileIndex, f, file, decl)
		if info == nil {
			return nil
		}
		return []*DeclInfo{info}
	case *syntax.VarDecl:
		if !varTrap {
			return nil
		}
		varDecls := collectVarDecls(Kind_Var, decl.NameList, decl.Type)
		for _, varDecl := range varDecls {
			varDecl.VarDecl = decl

			varDecl.FileSyntax = f
			varDecl.FileIndex = fileIndex
			varDecl.File = file
		}
		return varDecls
	case *syntax.ConstDecl:
		if !varTrap {
			return nil
		}
		constDecls := collectVarDecls(Kind_Const, decl.NameList, decl.Type)
		for _, constDecl := range constDecls {
			constDecl.ConstDecl = decl

			constDecl.FileSyntax = f
			constDecl.FileIndex = fileIndex
			constDecl.File = file
		}
		return constDecls
	case *syntax.TypeDecl:
		if decl.Alias {
			return nil
		}
		// TODO: test generic interface
		if len(decl.TParamList) > 0 {
			return nil
		}

		// NOTE: for interface type, we only set a marker
		// because we cannot handle Embed interface if
		// the that comes from other package
		if _, ok := decl.Type.(*syntax.InterfaceType); ok {
			return []*DeclInfo{
				{
					RecvTypeName: decl.Name.Value,
					Interface:    true,

					FileSyntax: f,
					FileIndex:  fileIndex,
					File:       file,
					Line:       int(decl.Pos().Line()),
				},
			}
		}
	}
	return nil
}

func getFuncDeclInfo(fileIndex int, f *syntax.File, file string, fn *syntax.FuncDecl) *DeclInfo {
	if fn.Body == nil {
		// see bug https://github.com/xhd2015/xgo/issues/202
		return nil
	}
	line := fn.Pos().Line()
	fnName := fn.Name.Value
	// there are cases where fnName is _
	if fnName == "" || fnName == "_" || fnName == "init" || strings.HasPrefix(fnName, "_cgo") || strings.HasPrefix(fnName, "_Cgo") {
		// skip cgo also,see https://github.com/xhd2015/xgo/issues/80#issuecomment-2067976575
		return nil
	}
	var genericFunc bool
	if len(fn.TParamList) > 0 {
		genericFunc = true
	}
	var recvTypeName string
	var recvPtr bool
	var recvName string
	var genericRecv bool
	fillMissingArgNames(fn)
	if fn.Recv != nil {
		recvName = "_"
		if fn.Recv.Name != nil {
			recvName = fn.Recv.Name.Value
		}

		recvTypeExpr := fn.Recv.Type

		// *A
		if starExpr, ok := fn.Recv.Type.(*syntax.Operation); ok && starExpr.Op == syntax.Mul {
			// *A
			recvTypeExpr = starExpr.X
			recvPtr = true
		}
		// check if generic
		if indexExpr, ok := recvTypeExpr.(*syntax.IndexExpr); ok {
			// *A[T] or A[T]
			// the generic receiver
			// currently not handled
			genericRecv = true
			recvTypeExpr = indexExpr.X
		}

		recvTypeName = recvTypeExpr.(*syntax.Name).Value
	}
	var firstArgCtx bool
	var lastResErr bool
	if false {
		// NOTE: these fields will be retrieved at runtime dynamically
		if len(fn.Type.ParamList) > 0 && hasQualifiedName(fn.Type.ParamList[0].Type, "context", "Context") {
			firstArgCtx = true
		}
		if len(fn.Type.ResultList) > 0 && isName(fn.Type.ResultList[len(fn.Type.ResultList)-1].Type, "error") {
			lastResErr = true
		}
	}

	return &DeclInfo{
		FuncDecl:     fn,
		Name:         fnName,
		RecvTypeName: recvTypeName,
		RecvPtr:      recvPtr,
		Generic:      genericFunc || genericRecv,

		Stdlib: base.Flag.Std,

		RecvName: recvName,
		ArgNames: getFieldNames(fn.Type.ParamList),
		ResNames: getFieldNames(fn.Type.ResultList),

		FirstArgCtx:  firstArgCtx,
		LastResError: lastResErr,

		FileSyntax: f,
		FileIndex:  fileIndex,
		File:       file,
		Line:       int(line),
	}
}

func getFileRef(i int) string {
	return fmt.Sprintf("__xgo_reg_file_gen_%d", i)
}

func gen() string {
	return ""
}

type StructDef struct {
	Field string
	Type  string
}

func genStructType(fields []*StructDef) string {
	concats := make([]string, 0, len(fields))
	for _, field := range fields {
		concats = append(concats, fmt.Sprintf("%s %s", field.Field, field.Type))
	}
	return fmt.Sprintf("struct{\n%s\n}\n", strings.Join(concats, "\n"))
}

// return a list of statements
//
//	   fileA := "..."
//	   fileB := "..."
//	   __xgo_link_generated_register_func(__xgo_local_func_stub{
//		     __xgo_local_pkg_name,
//	      0, // kind
//	      ...
//	   })
func generateFuncRegBody(pos syntax.Pos, funcDecls []*DeclInfo, xgoRegFunc string, xgoLocalFuncStub string) []syntax.Stmt {
	fileDeclaredMapping := make(map[int]bool)
	var fileDefs []syntax.Stmt
	stmts := make([]syntax.Stmt, 0, len(funcDecls))
	if xgo_ctxt.XGO_COMPILER_LOG_COST {
		fmt.Fprintf(os.Stderr, "funcDecls: %d\n", len(funcDecls))
	}

	logBodyGen := xgo_ctxt.LogSpan("funcDecl body gen")
	for _, funcDecl := range funcDecls {
		if funcDecl.Name == "_" {
			// there are function with name "_"
			continue
		}
		// type unknown constant expr, will not be registered,
		// see bug https://github.com/xhd2015/xgo/issues/53
		// why not filter them earlier?
		// because the node still needs to be marked, but effectively skipped
		if funcDecl.Kind == info.Kind_Const && funcDecl.ConstDecl.Type == nil {
			untypedConstType := getConstDeclValueType(funcDecl.ConstDecl.Values)
			if untypedConstType == "" || untypedConstType == UNKNOWN_CONST_TYPE {
				// exclude
				continue
			}
		}

		var fnRefName syntax.Expr
		var varRefName syntax.Expr
		if funcDecl.Kind.IsFunc() {
			if !funcDecl.Generic {
				fnRefName = funcDecl.RefNameSyntax(pos)
			}
		} else if funcDecl.Kind == info.Kind_Var {
			varRefName = &syntax.Operation{
				Op: syntax.And,
				X:  funcDecl.RefNameSyntax(pos),
			}
		} else if funcDecl.Kind == info.Kind_Const {
			varRefName = funcDecl.RefNameSyntax(pos)
		}
		fileIdx := funcDecl.FileIndex
		fileRef := getFileRef(fileIdx)

		if fnRefName == nil {
			fnRefName = syntax.NewName(pos, "nil")
		}
		if varRefName == nil {
			varRefName = syntax.NewName(pos, "nil")
		}

		// check expected__xgo_stub_def and __xgo_local_func_stub for correctness
		var _ = expected__xgo_stub_def
		regKind := func(kind info.DeclKind, identityName string) {
			fieldList := [...]syntax.Expr{
				syntax.NewName(pos, XgoLocalPkgName),         // PkgPath
				newIntLit(int(kind)),                         // Kind
				fnRefName,                                    // Fn
				varRefName,                                   // Var
				newIntLit(0),                                 // PC, filled later
				newBool(pos, funcDecl.Interface),             // Interface
				newBool(pos, funcDecl.Generic),               // Generic
				newBool(pos, funcDecl.Closure),               // Closure
				newBool(pos, funcDecl.Stdlib),                // Stdlib
				newStringLit(funcDecl.RecvTypeName),          // RecvTypeName
				newBool(pos, funcDecl.RecvPtr),               // RecvPtr
				newStringLit(funcDecl.Name),                  // Name
				newStringLit(identityName),                   // IdentityName
				newStringLit(funcDecl.RecvName),              // RecvName
				quoteNamesExprSyntax(pos, funcDecl.ArgNames), // ArgNames
				quoteNamesExprSyntax(pos, funcDecl.ResNames), // ResNames
				newBool(pos, funcDecl.FirstArgCtx),           // FirstArgCtx
				newBool(pos, funcDecl.LastResError),          // LastResErr
				syntax.NewName(pos, fileRef),                 /* declFunc.FileRef */ // File
				newIntLit(funcDecl.Line),                     // Line
			}
			// fields := strings.Join(fieldList, ",")
			// stmts = append(stmts, fmt.Sprintf("%s(%s{%s})", xgoRegFunc, xgoLocalFuncStub, fields))
			stmts = append(stmts, &syntax.ExprStmt{
				X: &syntax.CallExpr{
					Fun: syntax.NewName(pos, xgoRegFunc),
					ArgList: []syntax.Expr{
						&syntax.CompositeLit{
							Type:     syntax.NewName(pos, xgoLocalFuncStub),
							ElemList: fieldList[:],
							Rbrace:   pos,
						},
					},
				},
			})
		}
		identityName := funcDecl.IdentityName()
		regKind(funcDecl.Kind, identityName)
		if funcDecl.Kind == info.Kind_Var {
			regKind(info.Kind_VarPtr, "*"+identityName)
		}

		// add files
		if !fileDeclaredMapping[fileIdx] {
			fileDeclaredMapping[fileIdx] = true
			fileValue := funcDecl.File
			fileDefs = append(fileDefs, &syntax.AssignStmt{
				Op:  syntax.Def,
				Lhs: syntax.NewName(pos, fileRef),
				Rhs: newStringLit(fileValue),
			})
		}
	}
	logBodyGen()

	if len(stmts) == 0 {
		return nil
	}
	allStmts := make([]syntax.Stmt, 0, 2+len(fileDefs)+len(stmts))
	allStmts = append(allStmts, fileDefs...)
	allStmts = append(allStmts, stmts...)
	return allStmts
}
func generateRegFileCode(pkgName string, fnName string, body string) string {
	autoGenStmts := []string{
		"package " + pkgName,
		// "import \"reflect\"", // debug
		// "import \"fmt\"",     // debug
		// "const __XGO_SKIP_TRAP = true" + "\n" + // don't do this
		"func " + fnName + "(){",
		body,
		"}",
		"",
	}
	return strings.Join(autoGenStmts, "\n")
}

func generateRegHelperCode(pkgName string) string {
	code := helperCode
	if false {
		// debug
		code = strings.ReplaceAll(code, "__PKG__", xgo_ctxt.GetPkgPath())
	}
	autoGenStmts := []string{
		// padding for debug
		"// padding",
		"",
		"",
		"package " + pkgName,
		code,
	}
	return strings.Join(autoGenStmts, "\n")
}

func getFieldNames(fields []*syntax.Field) []string {
	names := make([]string, 0, len(fields))
	for _, field := range fields {
		names = append(names, getFieldName(field))
	}
	return names
}
func getFieldName(f *syntax.Field) string {
	if f == nil || f.Name == nil {
		return ""
	}
	return f.Name.Value
}

func quoteNamesExpr(names []string) string {
	if len(names) == 0 {
		return "nil"
	}
	qNames := make([]string, 0, len(names))
	for _, name := range names {
		qNames = append(qNames, strconv.Quote(name))
	}
	return "[]string{" + strings.Join(qNames, ",") + "}"
}

func quoteNamesExprSyntax(pos syntax.Pos, names []string) syntax.Expr {
	if len(names) == 0 {
		return syntax.NewName(pos, "nil")
	}
	qNames := make([]syntax.Expr, 0, len(names))
	for _, name := range names {
		qNames = append(qNames, newStringLit(name))
	}
	return &syntax.CompositeLit{
		Type: &syntax.SliceType{
			Elem: syntax.NewName(pos, "string"),
		},
		ElemList: qNames,
		Rbrace:   pos,
	}
}

func isName(expr syntax.Expr, name string) bool {
	nameExp, ok := expr.(*syntax.Name)
	if !ok {
		return false
	}
	return nameExp.Value == name
}
func hasQualifiedName(expr syntax.Expr, pkg string, name string) bool {
	sel, ok := expr.(*syntax.SelectorExpr)
	if !ok {
		return false
	}
	if sel.Sel.Value != name {
		return false
	}
	return isName(sel.X, pkg)
}
