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

func RegisterFuncTab(fset *token.FileSet, file *edit.File, pkgPath string, stdlib bool) {
	fileIndex := file.Index

	const FUNC_INFO = constants.RUNTIME_FUNC_TYPE
	const PKG_FUNC_INFO = constants.RUNTIME_PKG_FUNC_INFO_REF
	const REGISTER = constants.RUNTIME_REGISTER_FUNC

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
		varRegs = append(varRegs, fmt.Sprintf("%s.%s(%s)", PKG_FUNC_INFO, REGISTER, infoVar))
		if delayInitProp != "" {
			delayInits = append(delayInits, fmt.Sprintf("%s.%s=%s", infoVar, delayInitProp, delayInitValue))
		}
	}
	makeLiteral := func(kind string, name string, identityName string, pos token.Pos, extra []string) string {
		var suffix string
		if len(extra) > 0 {
			suffix = "," + strings.Join(extra, ",")
		}
		if stdlib {
			suffix = suffix + ",Stdlib:true"
		}
		lineNum := fset.Position(pos).Line
		return fmt.Sprintf("&%s.%s{Kind:%s.%s,Pkg:%s,Name:%q,IdentityName:%q,File:%s,Line:%d%s}", PKG_FUNC_INFO, FUNC_INFO, PKG_FUNC_INFO, kind,
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

		literal := makeLiteral("Kind_Var", varInfo.Name, varInfo.Name, varInfo.Decl.Decl.Pos(), extra)
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

		literal := makeLiteral("Kind_Func", funcInfo.FuncDecl.Name.Name, identityName, funcInfo.FuncDecl.Pos(), extra)
		addLiteral(funcInfo.InfoVar, literal, delayInitProp, delayInitValue)
	}
	for _, intfType := range file.InterfaceTypes {
		literal := makeLiteral("Kind_Func", intfType.Name, intfType.Name, intfType.Ident.Pos(), []string{
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

	patch.AddImport(fileEdit, fileSyntax, PKG_FUNC_INFO, constants.RUNTIME_CORE_INFO_PKG)

	absFile := file.File.AbsPath
	defLines := make([]string, 0, len(varDefs)+2)
	defLines = append(defLines,
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
	regCodeInit := "func init(){" + regCode + ";}"
	pos := patch.GetFuncInsertPosition(fileSyntax)
	patch.AddCode(fileEdit, pos, defCode+delayInitCode+";"+regCodeInit)
}
