package instrument_reg

import (
	"go/token"

	astutil "github.com/xhd2015/xgo/instrument/ast"
	"github.com/xhd2015/xgo/instrument/compiler_extra"
	"github.com/xhd2015/xgo/instrument/constants"
	"github.com/xhd2015/xgo/instrument/edit"
	"github.com/xhd2015/xgo/instrument/patch"
)

func RegisterFuncTab(fset *token.FileSet, file *edit.File, pkgPath string, stdlib bool) {
	fileIndex := file.Index
	perFilePkgNames := compiler_extra.PkgNames{
		REGISTER:     constants.Register(fileIndex),
		TRAP_FUNC:    constants.Trap(fileIndex),
		TRAP_VAR:     constants.TrapVar(fileIndex),
		TRAP_VAR_PTR: constants.TrapVarPtr(fileIndex),

		FUNC_INFO_TYPE: constants.FuncInfoType(fileIndex),
		PKG_VAR:        constants.PkgVar(fileIndex),

		XGO_INIT: constants.InitFunc(fileIndex),
	}
	FILE_VAR := constants.FileVar(fileIndex)
	FILE_VAR_FOR_VAR := constants.FileVarForVar(fileIndex)
	absFile := file.File.AbsPath
	fileDecls := buildCompilerExtra(fset, file)

	res := compiler_extra.GetFileRegStmts(fileDecls, stdlib, FILE_VAR, FILE_VAR_FOR_VAR, perFilePkgNames)
	if len(res.VarDefStmts) == 0 {
		return
	}

	regCode := compiler_extra.GenerateRegCode(perFilePkgNames.XGO_INIT, res.VarDefStmts, res.VarRegStmts, res.DelayInitStmts)
	initFunc := compiler_extra.DeclareInitFunc(perFilePkgNames)

	pkgStubs := compiler_extra.DeclarePkgStubs(pkgPath, perFilePkgNames)
	fileStub := compiler_extra.DeclareFileStub(fileIndex, absFile)
	fileForVarStub := compiler_extra.DeclareFileStubForVar(fileIndex, absFile)
	stubRegCode := compiler_extra.JoinStmts(compiler_extra.JoinStmts(pkgStubs...), fileStub, fileForVarStub, regCode)

	fileSyntax := file.File.Syntax
	pos := patch.GetFuncInsertPosition(fileSyntax)
	patch.AddCode(file.Edit, pos, stubRegCode)
	patch.Append(file.Edit, fileSyntax, "\n"+initFunc)
}

func buildCompilerExtra(fset *token.FileSet, file *edit.File) *compiler_extra.FileDecls {
	decls := &compiler_extra.FileDecls{}
	for _, funcInfo := range file.TrapFuncs {
		var recvTypeName string
		var receiver *compiler_extra.Field
		var params compiler_extra.Fields
		var results compiler_extra.Fields
		if funcInfo.Receiver != nil {
			receiver = &compiler_extra.Field{
				Name: funcInfo.Receiver.Name,
			}
		}
		if funcInfo.RecvType != nil {
			recvTypeName = funcInfo.RecvType.Name
		}
		if len(funcInfo.Params) > 0 {
			for _, param := range funcInfo.Params {
				params = append(params, &compiler_extra.Field{
					Name: param.Name,
				})
			}
		}
		if len(funcInfo.Results) > 0 {
			for _, result := range funcInfo.Results {
				results = append(results, &compiler_extra.Field{
					Name: result.Name,
				})
			}
		}
		pos := funcInfo.FuncDecl.Pos()
		lineNum := fset.Position(pos).Line
		decls.TrapFuncs = append(decls.TrapFuncs, compiler_extra.FuncInfo{
			IdentityName:     funcInfo.IdentityName,
			Name:             funcInfo.FuncDecl.Name.Name,
			HasGenericParams: astutil.IsGenericFunc(funcInfo.FuncDecl),
			LineNum:          lineNum,

			InfoVar: funcInfo.InfoVar,

			RecvPtr:      funcInfo.RecvPtr,
			RecvGeneric:  funcInfo.RecvGeneric,
			RecvTypeName: recvTypeName,

			Receiver: receiver,
			Params:   params,
			Results:  results,
		})
	}
	for _, varInfo := range file.TrapVars {
		pos := varInfo.Decl.Decl.Pos()
		lineNum := fset.Position(pos).Line
		decls.TrapVars = append(decls.TrapVars, compiler_extra.VarInfo{
			Name:    varInfo.Name,
			LineNum: lineNum,
			InfoVar: varInfo.InfoVar,
		})
	}
	for _, intfType := range file.InterfaceTypes {
		pos := intfType.Ident.Pos()
		lineNum := fset.Position(pos).Line
		decls.InterfaceTypes = append(decls.InterfaceTypes, compiler_extra.InterfaceType{
			Name:    intfType.Name,
			LineNum: lineNum,
			InfoVar: intfType.InfoVar,
		})
	}
	return decls
}
