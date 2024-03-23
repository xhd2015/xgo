package syntax

import (
	"cmd/compile/internal/syntax"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	xgo_ctxt "cmd/compile/internal/xgo_rewrite_internal/patch/ctxt"
	xgo_func_name "cmd/compile/internal/xgo_rewrite_internal/patch/func_name"
)

const sig_expected__xgo_register_func = `func(info interface{})`

func init() {
	if sig_gen__xgo_register_func != sig_expected__xgo_register_func {
		panic(fmt.Errorf("__xgo_register_func signature changed, run go generate and update sig_expected__xgo_register_func correspondly"))
	}
}

func AfterFilesParsed(fileList []*syntax.File, addFile func(name string, r io.Reader)) {
	afterFilesParsed(fileList, addFile)
}

func GetSyntaxDeclMapping() map[string]map[LineCol]*DeclInfo {
	return getSyntaxDeclMapping()
}

var allFiles []*syntax.File
var allDecls []*DeclInfo

func ClearFiles() {
	allFiles = nil
}

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

const xgoRuntimePkgPrefix = "github.com/xhd2015/xgo/runtime/"

func afterFilesParsed(fileList []*syntax.File, addFile func(name string, r io.Reader)) {
	if len(fileList) == 0 {
		return
	}
	if xgo_ctxt.SkipPackageTrap() {
		return
	}
	allFiles = fileList
	pkgPath := xgo_ctxt.GetPkgPath()

	if pkgPath == "" || pkgPath == "runtime" || strings.HasPrefix(pkgPath, "runtime/") || strings.HasPrefix(pkgPath, "internal/") || isSkippableSpecialPkg() {
		// runtime/internal should not be rewritten
		// internal/api has problem with the function register
		return
	}
	if strings.HasPrefix(pkgPath, xgoRuntimePkgPrefix) {
		if !strings.HasPrefix(pkgPath[len(xgoRuntimePkgPrefix):], "test/") {
			return
		}
	}
	pkgName := fileList[0].PkgName.Value
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

	funcDelcs := getFuncDecls(fileList)
	if len(funcDelcs) == 0 {
		return
	}
	// assign to global
	allDecls = funcDelcs

	// split fileDecls to a list of batch
	// when statements gets large, it will
	// exceeds the compiler's threshold, causing
	//     internal compiler error: NewBulk too big
	// see https://github.com/golang/go/issues/33437
	// see also: https://github.com/golang/go/issues/57832 The input code is just too big for the compiler to handle.
	// here we split the files per 1000 functions
	batchFuncDecls := splitBatch(funcDelcs, 1000)
	var subFuncNames []string
	for i, funcDecls := range batchFuncDecls {
		if len(funcDecls) == 0 {
			continue
		}

		body := generateFuncRegBody(funcDecls)
		if len(batchFuncDecls) == 1 {
			// special when there is only one file
			fileCode := generateRegFileCode(pkgName, "__xgo_register_funcs", body)
			addFile("__xgo_autogen_register_func_info.go", strings.NewReader(fileCode))
			continue
		}
		// declare sub function and sub file
		fnName := fmt.Sprintf("__xgo_register_funcs_%d", i)
		subFuncNames = append(subFuncNames, fnName)

		fileCode := generateRegFileCode(pkgName, fnName, body)
		fileName := fmt.Sprintf("__xgo_autogen_register_func_info_%d", i)
		addFile(fileName, strings.NewReader(fileCode))
	}
	if len(subFuncNames) > 0 {
		stmts := make([]string, 0, len(subFuncNames))
		for _, subFunc := range subFuncNames {
			stmts = append(stmts, fmt.Sprintf("%s(__xgo_reg_func)", subFunc))
		}
		body := strings.Join(stmts, "\n")
		fileCode := generateRegFileCode(pkgName, "__xgo_register_funcs", body)
		addFile("__xgo_autogen_register_func_info.go", strings.NewReader(fileCode))
	}
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
	var res [][]*DeclInfo

	var cur []*DeclInfo
	n := len(funcDecls)
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

type DeclInfo struct {
	FuncDecl     *syntax.FuncDecl
	Name         string
	RecvTypeName string
	RecvPtr      bool
	Generic      bool

	// arg names
	RecvName     string
	ArgNames     []string
	ResNames     []string
	FirstArgCtx  bool
	LastResError bool

	FileSyntax *syntax.File
	FileIndex  int
	File       string
	FileRef    string
	Line       int
}

func (c *DeclInfo) RefName() string {
	return xgo_func_name.FormatFuncRefName(c.RecvTypeName, c.RecvPtr, c.Name)
}

func (c *DeclInfo) RefAndGeneric() (refName string, genericName string) {
	refName = c.RefName()
	if !c.Generic {
		return refName, ""
	}
	return "nil", refName
}

func (c *DeclInfo) GenericName() string {
	if !c.Generic {
		return ""
	}
	return c.RefName()
}

func (c *DeclInfo) IdentityName() string {
	return xgo_func_name.FormatFuncRefName(c.RecvTypeName, c.RecvPtr, c.Name)
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

func getFuncDecls(files []*syntax.File) []*DeclInfo {
	// fileInfos := make([]*FileDecl, 0, len(files))
	var declFuncs []*DeclInfo
	for i, f := range files {
		file := f.Pos().RelFilename()
		for _, decl := range f.DeclList {
			fn, ok := decl.(*syntax.FuncDecl)
			if !ok {
				continue
			}
			line := fn.Pos().Line()
			if fn.Name.Value == "init" {
				continue
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
			if len(fn.Type.ParamList) > 0 && hasQualifiedName(fn.Type.ParamList[0].Type, "context", "Context") {
				firstArgCtx = true
			}
			if len(fn.Type.ResultList) > 0 && isName(fn.Type.ResultList[len(fn.Type.ResultList)-1].Type, "error") {
				lastResErr = true
			}

			declFuncs = append(declFuncs, &DeclInfo{
				FuncDecl:     fn,
				Name:         fn.Name.Value,
				RecvTypeName: recvTypeName,
				RecvPtr:      recvPtr,
				Generic:      genericFunc || genericRecv,

				RecvName: recvName,
				ArgNames: getFieldNames(fn.Type.ParamList),
				ResNames: getFieldNames(fn.Type.ResultList),

				FirstArgCtx:  firstArgCtx,
				LastResError: lastResErr,

				FileSyntax: f,
				FileIndex:  i,
				File:       file,
				Line:       int(line),
			})
		}
	}
	return declFuncs
}

func getFileRef(i int) string {
	return fmt.Sprintf("__xgo_reg_file_gen_%d", i)
}
func generateFuncRegBody(funcDecls []*DeclInfo) string {
	const regStructType = `struct {
	PkgPath      string
	Fn           interface{}
	PC           uintptr // filled later
	Generic      bool
	RecvTypeName string
	RecvPtr      bool
	Name         string
	IdentityName string // name without pkgPath

	RecvName    string
	ArgNames    []string
	ResNames    []string
	FirstArgCtx bool // first argument is context.Context or sub type?
	LastResErr  bool // last res is error or sub type?

	File string
	Line int
}
`
	fileDeclaredMapping := make(map[int]bool)
	var fileDefs []string
	stmts := make([]string, 0, len(funcDecls))
	for _, funcDecl := range funcDecls {
		if funcDecl.Name == "_" {
			// there are function with name "_"
			continue
		}
		refName, _ := funcDecl.RefAndGeneric()
		fileIdx := funcDecl.FileIndex
		fileRef := getFileRef(fileIdx)
		// func(pkgPath string, fn interface{}, recvTypeName string, recvPtr bool, name string, identityName string, generic bool, recvName string, argNames []string, resNames []string, firstArgCtx bool, lastResErr bool, file string, line int)
		fieldList := []string{
			"__xgo_regPkgPath",                        // PkgPath
			refName,                                   // Fn
			"0",                                       // PC
			strconv.FormatBool(funcDecl.Generic),      // Generic
			strconv.Quote(funcDecl.RecvTypeName),      // RecvTypeName
			strconv.FormatBool(funcDecl.RecvPtr),      // RecvPtr
			strconv.Quote(funcDecl.Name),              // Name
			strconv.Quote(funcDecl.IdentityName()),    // IdentityName
			strconv.Quote(funcDecl.RecvName),          // RecvName
			quoteNamesExpr(funcDecl.ArgNames),         // ArgNames
			quoteNamesExpr(funcDecl.ResNames),         // ResNames
			strconv.FormatBool(funcDecl.FirstArgCtx),  // FirstArgCtx
			strconv.FormatBool(funcDecl.LastResError), // LastResErr
			fileRef, /* declFunc.FileRef */ // File
			strconv.FormatInt(int64(funcDecl.Line), 10), // Line
		}
		fields := strings.Join(fieldList, ",")
		stmts = append(stmts, "__xgo_reg_func(__xgo_reg_func_struct_info{"+fields+"})")

		// add files
		if !fileDeclaredMapping[fileIdx] {
			fileDeclaredMapping[fileIdx] = true
			fileValue := funcDecl.FileSyntax.Pos().RelFilename()
			fileDefs = append(fileDefs, fmt.Sprintf("%s := %q", fileRef, fileValue))
		}
	}

	if len(stmts) == 0 {
		return ""
	}
	allStmts := make([]string, 0, 2+len(fileDefs)+len(stmts))
	allStmts = append(allStmts, `__xgo_regPkgPath := `+strconv.Quote(xgo_ctxt.GetPkgPath()))
	if false {
		// debug
		allStmts = append(allStmts, `__xgo_reg_func_old:=__xgo_reg_func; __xgo_reg_func = func(info interface{}){
			fmt.Print("reg:"+__xgo_regPkgPath+"\n")
			v := reflect.ValueOf(info)
			if v.Kind() != reflect.Struct {
				panic("non struct:"+__xgo_regPkgPath)
			}
			__xgo_reg_func_old(info)
		}`)
	}
	allStmts = append(allStmts, `type __xgo_reg_func_struct_info = `+regStructType)
	// debug, do not include file paths
	allStmts = append(allStmts, fileDefs...)
	if false {
		// debug
		pkgPath := xgo_ctxt.GetPkgPath()
		if strings.HasSuffix(pkgPath, "dao/impl") {
			if true {
				code := strings.Join(append(allStmts, stmts...), "\n")
				os.WriteFile("/tmp/test.go", []byte(code), 0755)
				panic("shit")
			}

			if len(stmts) > 100 {
				stmts = stmts[:100]
			}
		}
	}
	allStmts = append(allStmts, stmts...)
	return strings.Join(allStmts, "\n")
}
func generateRegFileCode(pkgName string, fnName string, body string) string {
	autoGenStmts := []string{
		"package " + pkgName,
		// "import \"reflect\"", // debug
		// "import \"fmt\"",     // debug
		// "const __XGO_SKIP_TRAP = true" + "\n" + // don't do this
		"func " + fnName + "(__xgo_reg_func " + sig_gen__xgo_register_func + "){",
		body,
		"}",
		"",
	}
	return strings.Join(autoGenStmts, "\n")
}

func getFieldNames(x []*syntax.Field) []string {
	names := make([]string, 0, len(x))
	for _, p := range x {
		var name string
		if p.Name != nil {
			name = p.Name.Value
		}
		names = append(names, name)
	}
	return names
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
