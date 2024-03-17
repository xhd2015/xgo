package syntax

import (
	"cmd/compile/internal/syntax"
	"cmd/compile/internal/types"
	"fmt"
	"io"
	"strconv"
	"strings"

	xgo_func_name "cmd/compile/internal/xgo_rewrite_internal/patch/func_name"
)

const sig_expected__xgo_register_func = "func(pkgPath string, fn interface{}, recvTypeName string, recvPtr bool, name string, identityName string, generic bool, recvName string, argNames []string, resNames []string, firstArgCtx bool, lastResErr bool)"

func init() {
	if sig_gen__xgo_register_func != sig_expected__xgo_register_func {
		panic(fmt.Errorf("__xgo_register_func signature changed, run go generate and update sig_expected__xgo_register_func correspondly"))
	}
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

func GetSyntaxDeclMapping() map[string]map[LineCol]*DeclInfo {
	if syntaxDeclMapping != nil {
		return syntaxDeclMapping
	}
	// build pos -> syntax mapping
	syntaxDeclMapping = make(map[string]map[LineCol]*DeclInfo)
	for _, syntaxDecl := range allDecls {
		pos := syntaxDecl.FuncDecl.Pos()
		file := pos.RelFilename()
		fileMapping := syntaxDeclMapping[file]
		if fileMapping == nil {
			fileMapping = make(map[LineCol]*DeclInfo)
			syntaxDeclMapping[file] = fileMapping
		}
		fileMapping[LineCol{
			Line: pos.Line(),
			Col:  pos.Col(),
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

func AfterFilesParsed(fileList []*syntax.File, addFile func(name string, r io.Reader)) {
	if len(fileList) == 0 {
		return
	}
	allFiles = fileList
	pkgPath := types.LocalPkg.Path

	// if pkgPath == "github.com/xhd2015/xgo/runtime/core/functab_debug" {

	// }

	if pkgPath == "" || pkgPath == "runtime" || strings.HasPrefix(pkgPath, "runtime/") || strings.HasPrefix(pkgPath, "internal/") || strings.HasPrefix(pkgPath, "github.com/xhd2015/xgo/runtime/") {
		// runtime/internal should not be rewritten
		// internal/api has problem with the function register
		return
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
	decls, body := getRegFuncsBody(fileList)
	allDecls = decls
	if body == "" {
		return
	}

	autoGen :=
		"package " + pkgName + "\n" +
			// "const __XGO_SKIP_TRAP = true" + "\n" + // don't do this
			"func __xgo_register_funcs(__xgo_reg_func " + sig_gen__xgo_register_func + "){\n" +
			body +
			"\n}"
	// ioutil.WriteFile("test.log", []byte(autoGen), 0755)
	addFile("__xgo_autogen.go", strings.NewReader(autoGen))
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

// collect funcs from files, register each of them by
// calling to __xgo_reg_func with names and func pointer
func getRegFuncsBody(files []*syntax.File) ([]*DeclInfo, string) {
	var declFuncs []*DeclInfo
	for _, f := range files {
		for _, decl := range f.DeclList {
			fn, ok := decl.(*syntax.FuncDecl)
			if !ok {
				continue
			}
			if fn.Name.Value == "init" {
				continue
			}
			var genericFunc bool
			if len(fn.TParamList) > 0 {
				genericFunc = true
				// cannot handle generic
			}
			var recvTypeName string
			var recvPtr bool
			var recvName string
			var genericRecv bool
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
			})
		}
	}

	stmts := make([]string, 0, len(declFuncs))
	for _, declFunc := range declFuncs {
		if declFunc.Name == "_" {
			// there are function with name "_"
			continue
		}
		refName, _ := declFunc.RefAndGeneric()
		// pkgPath string, fn interface{}, recvTypeName string, recvPtr bool, name string,identityName string, generic bool, recvName string, argNames []string, resNames []string, firstArgCtx bool, lastResErr bool
		stmts = append(stmts, fmt.Sprintf("__xgo_reg_func(__xgo_regPkgPath,%s)",
			strings.Join([]string{
				refName,
				strconv.Quote(declFunc.RecvTypeName), strconv.FormatBool(declFunc.RecvPtr), strconv.Quote(declFunc.Name),
				strconv.Quote(declFunc.IdentityName()), strconv.FormatBool(declFunc.Generic), // generic
				strconv.Quote(declFunc.RecvName), quoteNamesExpr(declFunc.ArgNames), quoteNamesExpr(declFunc.ResNames),
				strconv.FormatBool(declFunc.FirstArgCtx), strconv.FormatBool(declFunc.LastResError)}, ","),
		))
	}
	if len(stmts) == 0 {
		return nil, ""
	}
	return declFuncs, fmt.Sprintf(" __xgo_regPkgPath := %q\n", types.LocalPkg.Path) + strings.Join(stmts, "\n")
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
