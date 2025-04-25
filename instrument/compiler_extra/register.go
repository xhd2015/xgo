package compiler_extra

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/xhd2015/xgo/instrument/constants"
)

type PkgNames struct {
	REGISTER     string
	TRAP_FUNC    string
	TRAP_VAR     string
	TRAP_VAR_PTR string

	FUNC_INFO_TYPE string
	PKG_VAR        string

	XGO_INIT string
}

type Result struct {
	VarDefStmts    []string
	VarRegStmts    []string
	DelayInitStmts []string
}

func GetFileRegStmts(file *FileDecls, stdlib bool, fileVar string, fileVarForVar string, names PkgNames) Result {
	// TODO: add fn and var ptr
	regCap := len(file.TrapFuncs) + len(file.TrapVars) + len(file.InterfaceTypes)
	varDefs := make([]string, 0, regCap)
	varRegs := make([]string, 0, regCap)
	delayInits := make([]string, 0, len(file.TrapFuncs))
	addLiteral := func(lit Literal) {
		varDefs = append(varDefs, lit.VarDefs)
		varRegs = append(varRegs, lit.VarRegs)
		delayInits = append(delayInits, lit.DelayInits)
	}
	for _, varInfo := range file.TrapVars {
		lit := DefineVarLiteral(&varInfo, stdlib, fileVarForVar, names)
		addLiteral(lit)
	}

	for _, funcInfo := range file.TrapFuncs {
		funcLiteral := DefineFuncLiteral(&funcInfo, stdlib, fileVar, names)
		addLiteral(funcLiteral)
	}

	for _, intfType := range file.InterfaceTypes {
		lit := DefineIntfTypeLiteral(&intfType, stdlib, fileVar, names)
		addLiteral(lit)
	}

	return Result{
		VarDefStmts:    varDefs,
		VarRegStmts:    varRegs,
		DelayInitStmts: delayInits,
	}
}

func GenerateRegCode(__xgo_init string, varDefs []string, varRegs []string, delayInits []string) string {
	regCode := JoinStmts(varRegs...)

	delayInitCode := JoinStmts(delayInits...)
	var delayInitFunc string
	if delayInitCode != "" {
		delayInitFunc = "func init(){" + delayInitCode + ";}"
	}
	varDefStmts := JoinStmts(varDefs...)
	// call __xgo_init(), then register func infos
	regCodeInit := fmt.Sprintf("func init(){%s;}", JoinStmts(__xgo_init+"()", regCode))
	return JoinStmts(varDefStmts, delayInitFunc, regCodeInit)
}

func DefineFuncLiteral(funcInfo *FuncInfo, stdlib bool, fileVar string, names PkgNames) Literal {
	REGISTER := names.REGISTER

	FUNC_INFO_TYPE := names.FUNC_INFO_TYPE
	PKG_VAR := names.PKG_VAR
	FILE_VAR := fileVar

	identityName := funcInfo.IdentityName
	recvGeneric := funcInfo.RecvGeneric
	var extra []string
	if funcInfo.Receiver != nil {
		extra = append(extra,
			fmt.Sprintf("RecvPtr:%v", funcInfo.RecvPtr),
			fmt.Sprintf("RecvType: %q", funcInfo.RecvTypeName),
			fmt.Sprintf("RecvName:%q", funcInfo.Receiver.Name),
		)
	}

	// avoid initialization cycle
	var delayInitProp string
	var delayInitValue string
	if !recvGeneric && !funcInfo.HasGenericParams {
		delayInitProp = "Func"
		delayInitValue = identityName
	} else {
		extra = append(extra, "Generic:true")
	}
	if len(funcInfo.Params) > 0 {
		extra = append(extra, fmt.Sprintf("ArgNames:[]string{%s}", JoinQuoteNames(funcInfo.Params.Names(), ",")))
	}
	if len(funcInfo.Results) > 0 {
		extra = append(extra, fmt.Sprintf("ResNames:[]string{%s}", JoinQuoteNames(funcInfo.Results.Names(), ",")))
	}

	literal := makeLiteral(FUNC_INFO_TYPE, PKG_VAR, FILE_VAR, constants.InfoKind_Func, funcInfo.Name, identityName, funcInfo.LineNum, stdlib, extra)
	return defineLiteral(REGISTER, funcInfo.InfoVar, literal, delayInitProp, delayInitValue)
}

func DefineVarLiteral(varInfo *VarInfo, stdlib bool, FILE_VAR string, names PkgNames) Literal {
	REGISTER := names.REGISTER
	PKG_VAR := names.PKG_VAR
	FUNC_INFO_TYPE := names.FUNC_INFO_TYPE

	extra := []string{
		"Var:" + "&" + varInfo.Name,
		fmt.Sprintf("ResNames:[]string{%q}", varInfo.Name),
	}

	literal := makeLiteral(FUNC_INFO_TYPE, PKG_VAR, FILE_VAR, constants.InfoKind_Var, varInfo.Name, varInfo.Name, varInfo.LineNum, stdlib, extra)
	return defineLiteral(REGISTER, varInfo.InfoVar, literal, "", "")
}

func DefineIntfTypeLiteral(intfType *InterfaceType, stdlib bool, fileVar string, names PkgNames) Literal {
	REGISTER := names.REGISTER
	FUNC_INFO_TYPE := names.FUNC_INFO_TYPE
	PKG_VAR := names.PKG_VAR
	FILE_VAR := fileVar

	literal := makeLiteral(FUNC_INFO_TYPE, PKG_VAR, FILE_VAR, constants.InfoKind_Func, intfType.Name, intfType.Name, intfType.LineNum, stdlib, []string{
		"Interface:true",
		fmt.Sprintf("RecvType: %q", intfType.Name),
	})
	return defineLiteral(REGISTER, intfType.InfoVar, literal, "", "")
}

type Literal struct {
	VarDefs    string
	VarRegs    string
	DelayInits string
}

func defineLiteral(REGISTER string, infoVar string, literal string, delayInitProp string, delayInitValue string) Literal {
	varDefs := fmt.Sprintf("var %s=%s", infoVar, literal)
	varRegs := fmt.Sprintf("%s(%s)", REGISTER, infoVar)
	var delayInits string
	if delayInitProp != "" {
		delayInits = fmt.Sprintf("%s.%s=%s", infoVar, delayInitProp, delayInitValue)
	}
	return Literal{
		VarDefs:    varDefs,
		VarRegs:    varRegs,
		DelayInits: delayInits,
	}
}

func makeLiteral(FUNC_INFO_TYPE string, PKG_VAR string, FILE_VAR string, kind constants.InfoKind, name string, identityName string, lineNum int, stdlib bool, extra []string) string {
	var suffix string
	if len(extra) > 0 {
		suffix = "," + strings.Join(extra, ",")
	}
	if stdlib {
		suffix = suffix + ",Stdlib:true"
	}
	return fmt.Sprintf("&%s{Kind:%d,Pkg:%s,Name:%q,IdentityName:%q,File:%s,Line:%d%s}", FUNC_INFO_TYPE,
		kind,
		PKG_VAR,
		name,
		identityName,
		FILE_VAR,
		lineNum,
		suffix,
	)
}

func QuoteNames(names []string) []string {
	quotedNames := make([]string, len(names))
	for i, name := range names {
		quotedNames[i] = strconv.Quote(name)
	}
	return quotedNames
}

func JoinQuoteNames(names []string, sep string) string {
	return strings.Join(QuoteNames(names), sep)
}

func DeclarePkgStubs(pkgPath string, names PkgNames) []string {
	return []string{
		fmt.Sprintf("type %s struct{%s}", names.FUNC_INFO_TYPE, strings.Join(constants.FUNC_INFO_FIELDS, ";")),
		fmt.Sprintf("var %s = %s", names.REGISTER, constants.REGISTER_SIGNATURE),
		fmt.Sprintf("var %s = %s", names.TRAP_FUNC, constants.TRAP_FUNC_SIGNATURE),
		fmt.Sprintf("var %s = %s", names.TRAP_VAR, constants.TRAP_VAR_SIGNATURE),
		fmt.Sprintf("var %s = %s", names.TRAP_VAR_PTR, constants.TRAP_VAR_PTR_SIGNATURE),
		fmt.Sprintf("var %s=%q", names.PKG_VAR, pkgPath),
	}
}

func DeclareFileStub(fileIndex int, absFile string) string {
	return fmt.Sprintf("var %s=%q", constants.FileVar(fileIndex), absFile)
}

func DeclareFileStubForVar(fileIndex int, absFile string) string {
	return fmt.Sprintf("var %s=%q", constants.FileVarForVar(fileIndex), absFile)
}

func DeclareFileStubGc(fileIndex int, absFile string) string {
	return fmt.Sprintf("var %s=%q", constants.FileVarGc(fileIndex), absFile)
}

// make IR not inline
// check patch/link/link_ir.go
func DeclareInitFunc(names PkgNames) string {
	return fmt.Sprintf("//go:noinline\nfunc %s(){%s=%s;%s=%s;%s=%s;%s=%s;}", names.XGO_INIT,
		// self assign
		names.REGISTER, names.REGISTER,
		names.TRAP_FUNC, names.TRAP_FUNC,
		names.TRAP_VAR, names.TRAP_VAR,
		names.TRAP_VAR_PTR, names.TRAP_VAR_PTR,
	)
}

func JoinStmts(stmts ...string) string {
	var buf strings.Builder
	var nbytes int
	for _, stmt := range stmts {
		nbytes += len(stmt)
	}
	// extra for separator
	buf.Grow(nbytes + len(stmts))

	n := len(stmts)
	var lastStmt string
	for i := 0; i < n; i++ {
		stmt := stmts[i]
		if stmt == "" {
			continue
		}
		if prevStmtNeedsSeparator(lastStmt) {
			buf.WriteString(";")
		}
		lastStmt = stmt
		buf.WriteString(stmt)
	}

	return buf.String()
}

func prevStmtNeedsSeparator(s string) bool {
	if s == "" {
		return false
	}
	last := s[len(s)-1]
	return last != ';' && last != '\n'
}
