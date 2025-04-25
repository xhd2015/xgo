package syntax

import (
	"cmd/compile/internal/base"
	"cmd/compile/internal/syntax"
	"cmd/compile/internal/xgo_rewrite_internal/patch/info"
)

type File struct {
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
			Syntax:  f,
			Index:   i,
			AbsPath: file,
		}
	}
	return files
}

func getFuncDecls(files []*File) []*info.DeclInfo {
	// fileInfos := make([]*FileDecl, 0, len(files))
	var declFuncs []*info.DeclInfo
	for _, f := range files {
		for declIdx, decl := range f.Syntax.DeclList {
			fnDecls := extractFuncDecls(f.Index, declIdx, f.Syntax, f.AbsPath, decl)
			declFuncs = append(declFuncs, fnDecls...)
		}
	}
	return declFuncs
}

func extractFuncDecls(fileIndex int, declIndex int, f *syntax.File, file string, decl syntax.Decl) []*info.DeclInfo {
	switch decl := decl.(type) {
	case *syntax.FuncDecl:
		fnInfo := getFuncDeclInfo(fileIndex, declIndex, f, file, decl)
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
			return []*info.DeclInfo{
				{
					RecvTypeName: decl.Name.Value,
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

func getFuncDeclInfo(fileIndex int, declIndex int, f *syntax.File, file string, fn *syntax.FuncDecl) *info.DeclInfo {
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

	return &info.DeclInfo{
		FuncDecl:     fn,
		Name:         fnName,
		RecvTypeName: recvTypeName,
		RecvPtr:      recvPtr,
		Generic:      genericFunc || genericRecv,
		RecvGeneric:  genericRecv,

		Stdlib: base.Flag.Std,

		RecvName: recvName,
		ArgNames: getFieldNames(fn.Type.ParamList),
		ResNames: getFieldNames(fn.Type.ResultList),

		FileSyntax: f,
		FileIndex:  fileIndex,
		DeclIndex:  declIndex,

		File: file,
		Line: int(line),
	}
}
