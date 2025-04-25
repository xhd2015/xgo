package syntax

import (
	"cmd/compile/internal/base"
	"cmd/compile/internal/syntax"
	"cmd/compile/internal/xgo_rewrite_internal/patch/ctxt"
	"cmd/compile/internal/xgo_rewrite_internal/patch/info"
	"cmd/compile/internal/xgo_rewrite_internal/patch/instrument/compiler_extra"
	"cmd/compile/internal/xgo_rewrite_internal/patch/instrument/constants"
	"fmt"
	"io"
	"strconv"
	"strings"
)

func AfterFilesParsed(syntaxFiles []*syntax.File, addFile func(name string, r io.Reader) *syntax.File) {
	if len(syntaxFiles) == 0 {
		return
	}
	pkgPath := ctxt.GetPkgPath()
	// mainModule := ctxt.XGO_MAIN_MODULE
	packagesFile := ctxt.XGO_COMPILER_SYNTAX_REWRITE_PACKAGES_FILE

	var mapping *compiler_extra.PackagesMapping
	if packagesFile != "" {
		var err error
		mapping, err = compiler_extra.ParseMapping(packagesFile)
		if err != nil {
			panic(fmt.Errorf("failed to parse packages file: %w", err))
		}
	}

	var pkgMapping *compiler_extra.PackageMapping
	if mapping != nil {
		pkgMapping = mapping.Packages[pkgPath]
	}
	if pkgMapping == nil || len(pkgMapping.Files) == 0 {
		return
	}

	files := initFiles(syntaxFiles)
	funcDelcs := getFuncDecls(files)

	// always __xgo_trap_0
	__xgo_trap := constants.Trap(0)
	trapCount := trapFuncs(funcDelcs, __xgo_trap, pkgMapping.Files)
	if trapCount == 0 {
		return
	}
	// filter
	if trapCount != len(funcDelcs) {
		funcDelcs = filterFuncDecls(funcDelcs, trapCount)
	}
	batchFuncDecls := splitBatch(funcDelcs, 1024)

	names := compiler_extra.PkgNames{
		REGISTER:       constants.Register(0),
		TRAP_FUNC:      __xgo_trap,
		TRAP_VAR:       constants.TrapVar(0),
		TRAP_VAR_PTR:   constants.TrapVarPtr(0),
		FUNC_INFO_TYPE: constants.FuncInfoType(0),
		PKG_VAR:        constants.PkgVar(0),
		XGO_INIT:       constants.InitFunc(0),
	}

	var fileStubs []string
	for _, file := range files {
		fileStub := compiler_extra.DeclareFileStubGc(file.Index, file.AbsPath)
		fileStubs = append(fileStubs, fileStub)
	}

	stdlib := base.Flag.Std

	pkgName := syntaxFiles[0].PkgName.Value
	totalBatch := len(batchFuncDecls)

	var declaredPkgAndFileStubs bool

	// with var trap, names will have already been
	// declared
	hasVarTrap := pkgMapping.HasVarTrap
	for i, funcDecls := range batchFuncDecls {
		// each batch represents a file
		if len(funcDecls) == 0 {
			continue
		}
		// NOTE: here the trick is to use init across multiple files,
		// in go, init can appear more than once even in single file
		fileNameBase := "__xgo_autogen_register_func_info"
		if totalBatch > 1 {
			// only when there are many functions,
			// we need to generate divide our declares
			// into multiple files to avoid large file
			// special when there are multiple files
			fileNameBase += "_" + strconv.Itoa(i)
		}

		var pkgStmts []string
		if !declaredPkgAndFileStubs {
			pkgStmts = make([]string, 0, len(fileStubs))
			pkgStmts = append(pkgStmts, fileStubs...)
			if !hasVarTrap {
				pkgStubs := compiler_extra.DeclarePkgStubs(pkgPath, names)
				pkgStmts = append(pkgStmts, pkgStubs...)
				// the init func should have `//noinline` directive, so ensure
				// it is separated from other stmts
				pkgStmts = append(pkgStmts, "\n"+compiler_extra.DeclareInitFunc(names)+"\n")
			}
			declaredPkgAndFileStubs = true
		}

		var varDefs []string
		var varRegs []string
		var delayInits []string
		for _, funcDecl := range funcDecls {
			var lit compiler_extra.Literal
			if funcDecl.Kind == info.Kind_Func {
				if funcDecl.Interface {
					lit = compiler_extra.DefineIntfTypeLiteral(convertInterfaceTypeDecl(funcDecl), stdlib, constants.FileVarGc(funcDecl.FileIndex), names)
				} else {
					lit = compiler_extra.DefineFuncLiteral(convertFuncDecl(funcDecl), stdlib, constants.FileVarGc(funcDecl.FileIndex), names)
				}
			}

			varDefs = append(varDefs, lit.VarDefs)
			varRegs = append(varRegs, lit.VarRegs)
			delayInits = append(delayInits, lit.DelayInits)
		}

		// __xgo_init() can be called multiple times
		// we don't need special handling here
		// the rule is, we must ensure __xgo_init() is called
		// before registering func infos
		regCode := compiler_extra.GenerateRegCode(names.XGO_INIT, varDefs, varRegs, delayInits)
		pkgStmts = append(pkgStmts, regCode)

		body := compiler_extra.JoinStmts(pkgStmts...)

		regFileCode := fmt.Sprintf("package %s\n%s", pkgName, body)
		addFile(fileNameBase+".go", strings.NewReader(regFileCode))
	}
}

func filterFuncDecls(funcDecls []*info.DeclInfo, n int) []*info.DeclInfo {
	res := make([]*info.DeclInfo, 0, n)
	for _, funcDecl := range funcDecls {
		if !funcDecl.HasFuncTrap {
			continue
		}
		res = append(res, funcDecl)
	}
	return res
}

func convertFuncDecl(funcDecl *info.DeclInfo) *compiler_extra.FuncInfo {
	if funcDecl == nil {
		return nil
	}
	if funcDecl.Interface {
		panic("interface use convertInterfaceTypeDecl instead")
	}
	var receiver *compiler_extra.Field
	if funcDecl.RecvTypeName != "" {
		receiver = &compiler_extra.Field{
			Name: funcDecl.RecvTypeName,
		}
	}
	return &compiler_extra.FuncInfo{
		IdentityName:     funcDecl.IdentityName(),
		Name:             funcDecl.Name,
		HasGenericParams: funcDecl.Generic,
		LineNum:          funcDecl.Line,
		InfoVar:          funcDecl.FuncInfoVarName(),

		RecvPtr:      funcDecl.RecvPtr,
		RecvGeneric:  funcDecl.RecvGeneric,
		RecvTypeName: funcDecl.RecvTypeName,

		Receiver: receiver,
		Params:   convertFields(funcDecl.ArgNames),
		Results:  convertFields(funcDecl.ResNames),
	}
}

func convertInterfaceTypeDecl(typeDecl *info.DeclInfo) *compiler_extra.InterfaceType {
	if typeDecl == nil {
		return nil
	}
	return &compiler_extra.InterfaceType{
		Name:    typeDecl.Name,
		LineNum: typeDecl.Line,
		InfoVar: typeDecl.FuncInfoVarName(),
	}
}

func convertFields(fields []string) []*compiler_extra.Field {
	if fields == nil {
		return nil
	}
	res := make([]*compiler_extra.Field, 0, len(fields))
	for _, field := range fields {
		res = append(res, &compiler_extra.Field{
			Name: field,
		})
	}
	return res
}
