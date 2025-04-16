package instrument_reg

import (
	"fmt"
	"go/token"
	"strings"

	astutil "github.com/xhd2015/xgo/instrument/ast"
	"github.com/xhd2015/xgo/instrument/constants"
	"github.com/xhd2015/xgo/instrument/edit"
	"github.com/xhd2015/xgo/instrument/patch"
)

type InfoKind int

const (
	InfoKind_Func   InfoKind = 0
	InfoKind_Var    InfoKind = 1
	InfoKind_VarPtr InfoKind = 2
	InfoKind_Const  InfoKind = 3
)

var FUNC_INFO_FIELDS = []string{
	"Kind int",
	"FullName string",
	"Pkg string",
	"IdentityName string",
	"Name string",
	"RecvType string",
	"RecvPtr bool",
	"Interface bool",
	"Generic bool",
	"Closure bool",
	"Stdlib bool",
	"File string",
	"Line int",
	"PC uintptr",
	"Func interface{}",
	"Var interface{}",
	"RecvName string",
	"ArgNames []string",
	"ResNames []string",
	"FirstArgCtx bool",
	"LastResultErr bool",
}

func RegisterFuncTab(fset *token.FileSet, file *edit.File, pkgPath string, stdlib bool) {
	fileIndex := file.Index

	REGISTER := fmt.Sprintf("%s%d", constants.LINK_REGISTER, fileIndex)
	TRAP_FUNC := fmt.Sprintf("%s%d", constants.LINK_TRAP_FUNC, fileIndex)
	TRAP_VAR := fmt.Sprintf("%s%d", constants.LINK_TRAP_VAR, fileIndex)
	TRAP_VAR_PTR := fmt.Sprintf("%s%d", constants.LINK_TRAP_VAR_PTR, fileIndex)

	FUNC_INFO_TYPE := fmt.Sprintf("__xgo_func_info_%d", fileIndex)
	pkgVar := fmt.Sprintf("__xgo_pkg_%d", fileIndex)
	fileVar := fmt.Sprintf("__xgo_file_%d", fileIndex)

	// TODO: add fn and var ptr
	regCap := len(file.TrapFuncs) + len(file.TrapVars) + len(file.InterfaceTypes)
	varDefs := make([]string, 0, regCap)
	varRegs := make([]string, 0, regCap)
	delayInits := make([]string, 0, len(file.TrapFuncs))
	idx := 0
	addLiteral := func(infoVar string, literal string, delayInitProp string, delayInitValue string) {
		idx++
		varDefs = append(varDefs, fmt.Sprintf("var %s=%s", infoVar, literal))
		varRegs = append(varRegs, fmt.Sprintf("%s(%s)", REGISTER, infoVar))
		if delayInitProp != "" {
			delayInits = append(delayInits, fmt.Sprintf("%s.%s=%s", infoVar, delayInitProp, delayInitValue))
		}
	}
	makeLiteral := func(kind InfoKind, name string, identityName string, pos token.Pos, extra []string) string {
		var suffix string
		if len(extra) > 0 {
			suffix = "," + strings.Join(extra, ",")
		}
		if stdlib {
			suffix = suffix + ",Stdlib:true"
		}
		lineNum := fset.Position(pos).Line
		return fmt.Sprintf("&%s{Kind:%d,Pkg:%s,Name:%q,IdentityName:%q,File:%s,Line:%d%s}", FUNC_INFO_TYPE,
			kind,
			pkgVar,
			name,
			identityName,
			fileVar,
			lineNum,
			suffix,
		)
	}
	for _, varInfo := range file.TrapVars {
		extra := []string{
			"Var:" + "&" + varInfo.Name,
			fmt.Sprintf("ResNames:[]string{%q}", varInfo.Name),
		}

		literal := makeLiteral(InfoKind_Var, varInfo.Name, varInfo.Name, varInfo.Decl.Decl.Pos(), extra)
		addLiteral(varInfo.InfoVar, literal, "", "")
	}
	for _, funcInfo := range file.TrapFuncs {
		identityName := funcInfo.IdentityName
		recvGeneric := funcInfo.RecvGeneric
		var extra []string
		if funcInfo.Receiver != nil {
			extra = append(extra,
				fmt.Sprintf("RecvPtr:%v", funcInfo.RecvPtr),
				fmt.Sprintf("RecvType: %q", funcInfo.RecvType.Name),
				fmt.Sprintf("RecvName:%q", funcInfo.Receiver.Name),
			)
		}

		// avoid initialization cycle
		var delayInitProp string
		var delayInitValue string
		if !recvGeneric && !astutil.IsGenericFunc(funcInfo.FuncDecl) {
			delayInitProp = "Func"
			delayInitValue = identityName
		} else {
			extra = append(extra, "Generic:true")
		}
		if len(funcInfo.Params) > 0 {
			extra = append(extra, fmt.Sprintf("ArgNames:[]string{%s}", astutil.JoinQuoteNames(funcInfo.Params.Names(), ",")))
		}
		if len(funcInfo.Results) > 0 {
			extra = append(extra, fmt.Sprintf("ResNames:[]string{%s}", astutil.JoinQuoteNames(funcInfo.Results.Names(), ",")))
		}

		literal := makeLiteral(InfoKind_Func, funcInfo.FuncDecl.Name.Name, identityName, funcInfo.FuncDecl.Pos(), extra)
		addLiteral(funcInfo.InfoVar, literal, delayInitProp, delayInitValue)
	}

	for _, intfType := range file.InterfaceTypes {
		literal := makeLiteral(InfoKind_Func, intfType.Name, intfType.Name, intfType.Ident.Pos(), []string{
			"Interface:true",
			fmt.Sprintf("RecvType: %q", intfType.Name),
		})
		addLiteral(intfType.InfoVar, literal, "", "")
	}

	if len(varDefs) == 0 {
		return
	}
	fileEdit := file.Edit
	fileSyntax := file.File.Syntax

	absFile := file.File.AbsPath
	defLines := make([]string, 0, len(varDefs)+2)
	defLines = append(defLines,
		fmt.Sprintf("type %s struct{%s}", FUNC_INFO_TYPE, strings.Join(FUNC_INFO_FIELDS, ";")),
		fmt.Sprintf("var %s = %s", REGISTER, REGISTER_SIGNATURE),
		fmt.Sprintf("var %s = %s", TRAP_FUNC, TRAP_FUNC_SIGNATURE),
		fmt.Sprintf("var %s = %s", TRAP_VAR, TRAP_VAR_SIGNATURE),
		fmt.Sprintf("var %s = %s", TRAP_VAR_PTR, TRAP_VAR_PTR_SIGNATURE),
		fmt.Sprintf("var %s=%q", pkgVar, pkgPath),
		fmt.Sprintf("var %s=%q", fileVar, absFile),
	)
	defLines = append(defLines, varDefs...)

	defCode := strings.Join(defLines, ";")
	regCode := strings.Join(varRegs, ";")
	var delayInitCode string
	if len(delayInits) > 0 {
		delayInitCode = ";func init(){" + strings.Join(delayInits, ";") + ";}"
	}
	__xgo_init_i := fmt.Sprintf("__xgo_init_%d", fileIndex)
	regCodeInit := "func init(){" + __xgo_init_i + "();" + regCode + ";}"
	pos := patch.GetFuncInsertPosition(fileSyntax)
	patch.AddCode(fileEdit, pos, defCode+delayInitCode+";"+regCodeInit)

	// make IR not inline
	// check patch/link/link_ir.go
	__xgo_init_i_func := fmt.Sprintf("func %s(){%s=%s;%s=%s;%s=%s;%s=%s;}", __xgo_init_i,
		// self assign
		REGISTER, REGISTER,
		TRAP_FUNC, TRAP_FUNC,
		TRAP_VAR, TRAP_VAR,
		TRAP_VAR_PTR, TRAP_VAR_PTR,
	)
	patch.Append(fileEdit, fileSyntax, "\n//go:noinline\n"+__xgo_init_i_func)
}
