package syntax

import (
	"cmd/compile/internal/base"
	"cmd/compile/internal/syntax"
	"cmd/compile/internal/xgo_rewrite_internal/patch/info"
	"cmd/compile/internal/xgo_rewrite_internal/patch/instrument/compiler_extra"
	"path/filepath"
)

type File struct {
	Name    string
	Syntax  *syntax.File
	Index   int
	AbsPath string
}

func initFiles(syntaxFiles []*syntax.File) []*File {
	files := make([]*File, len(syntaxFiles))
	for i, f := range syntaxFiles {
		var file string
		if TrimFilename != nil {
			// >= go1.18
			file = TrimFilename(f.Pos().Base())
		} else if AbsFilename != nil {
			file = AbsFilename(f.Pos().Base().Filename())
		} else {
			// fallback to default
			file = f.Pos().RelFilename()
		}
		files[i] = &File{
			Name:    filepath.Base(file),
			Syntax:  f,
			Index:   i,
			AbsPath: file,
		}
	}
	return files
}

func getFuncDecls(files []*File, fileMapping map[string]*compiler_extra.FileMapping) ([]*File, []*info.DeclInfo) {
	// fileInfos := make([]*FileDecl, 0, len(files))
	var declFuncs []*info.DeclInfo

	i := 0
	for _, f := range files {
		fmapping := fileMapping[f.Name]
		if fmapping == nil {
			continue
		}
		var hasDecl bool
		for declIdx, decl := range f.Syntax.DeclList {
			fnDecls := extractFuncDecls(f.Index, declIdx, f.Syntax, f.AbsPath, decl, fmapping)
			if len(fnDecls) > 0 {
				hasDecl = true
			}
			declFuncs = append(declFuncs, fnDecls...)
		}
		if hasDecl {
			files[i] = f
			i++
		}
	}
	return files[:i], declFuncs
}

func extractFuncDecls(fileIndex int, declIndex int, f *syntax.File, file string, decl syntax.Decl, fileMapping *compiler_extra.FileMapping) []*info.DeclInfo {
	switch decl := decl.(type) {
	case *syntax.FuncDecl:
		fnInfo := getFuncDeclInfo(fileIndex, declIndex, f, file, decl, fileMapping.Funcs)
		if fnInfo == nil {
			return nil
		}
		return []*info.DeclInfo{fnInfo}
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
			idName := decl.Name.Value
			_, ok := fileMapping.Interfaces[idName]
			if !ok {
				return nil
			}
			return []*info.DeclInfo{
				{
					RecvTypeName: idName,
					Interface:    true,

					FileSyntax: f,
					FileIndex:  fileIndex,
					DeclIndex:  declIndex,
					File:       file,
					Line:       int(decl.Pos().Line()),
				},
			}
		}
	}
	return nil
}

func getFuncDeclInfo(fileIndex int, declIndex int, f *syntax.File, file string, fn *syntax.FuncDecl, funcsMapping map[string]*compiler_extra.FuncMapping) *info.DeclInfo {
	if fn.Body == nil {
		// see bug https://github.com/xhd2015/xgo/issues/202
		return nil
	}
	line := fn.Pos().Line()
	fnName := fn.Name.Value
	// there are cases where fnName is _
	if fnName == "" || fnName == "_" || fnName == "init" {
		// || strings.HasPrefix(fnName, "_cgo") || strings.HasPrefix(fnName, "_Cgo") {
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

	declInfo := &info.DeclInfo{
		FuncDecl:     fn,
		Name:         fnName,
		RecvTypeName: recvTypeName,
		RecvPtr:      recvPtr,
		Generic:      genericFunc || genericRecv,
		RecvGeneric:  genericRecv,

		Stdlib: base.Flag.Std,

		RecvName: recvName,
		// filled later
		// ArgNames: getFieldNames(fn.Type.ParamList),
		// ResNames: getFieldNames(fn.Type.ResultList),

		FileSyntax: f,
		FileIndex:  fileIndex,
		DeclIndex:  declIndex,

		File: file,
		Line: int(line),
	}
	idName := declInfo.IdentityName()
	_, ok := funcsMapping[idName]
	if !ok {
		return nil
	}
	declInfo.ArgNames = getFieldNames(fn.Type.ParamList)
	declInfo.ResNames = getFieldNames(fn.Type.ResultList)
	return declInfo
}
