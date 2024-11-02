package load

import (
	goast "go/ast"
	"go/token"

	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/ast"
	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/model"
	"github.com/xhd2015/xgo/support/fileutil"
)

// see ast.LoadFuncs
func FuncInfoMappingToAnnotation(mapping map[model.RelativeFile][]*model.FuncAnnotation) *model.ProjectAnnotation {
	if len(mapping) == 0 {
		return nil
	}
	files := make(model.FileAnnotationMapping, len(mapping))
	for file, funcInfos := range mapping {
		funcs := make(model.FuncAnnotationMapping, len(funcInfos))
		for _, funcInfo := range funcInfos {
			if funcInfo.Block == nil {
				continue
			}
			blockID := funcInfo.Block.ID()
			funcs[blockID] = funcInfo
		}
		files[file] = &model.FileAnnotation{
			Funcs: funcs,
		}
	}
	return &model.ProjectAnnotation{
		Files: files,
		Types: map[model.AnnotationType]bool{
			model.AnnotationType_FileFuncs: true,
		},
	}
}

func CollectFuncs(astInfo ast.LoadInfo) (model.FuncInfoMapping, error) {
	mapping := make(model.FuncInfoMapping)
	astInfo.RangeFiles(func(f ast.File) bool {
		mapping[model.RelativeFile(fileutil.Slashlize(f.RelPath()))] = extractFuncs(astInfo.FileSet(), f.Ast())
		return true
	})
	return mapping, nil
}

func extractFuncs(fset *token.FileSet, astFile *goast.File) []*model.FuncAnnotation {
	funcs := make([]*model.FuncAnnotation, 0)

	goast.Inspect(astFile, func(n goast.Node) bool {
		switch n := n.(type) {
		case *goast.FuncDecl:
			funcs = append(funcs, buildFunc(fset, n))
		case *goast.FuncLit:
			funcs = append(funcs, buildClosure(fset, n))
		}
		return true
	})
	return funcs
}

func buildFunc(fset *token.FileSet, fdecl *goast.FuncDecl) *model.FuncAnnotation {
	start := fset.Position(fdecl.Pos())
	end := fset.Position(fdecl.End())
	var ownerTypeName string
	var ptr bool
	if fdecl.Recv != nil && len(fdecl.Recv.List) > 0 {
		tp := fdecl.Recv.List[0].Type
		if tp, ok := tp.(*goast.StarExpr); ok {
			ptr = true
			if idt, ok := tp.X.(*goast.Ident); ok {
				ownerTypeName = idt.Name
			}
		}
		if tp, ok := tp.(*goast.Ident); ok {
			ownerTypeName = tp.Name
		}
	}
	return &model.FuncAnnotation{
		Block: &model.Block{
			StartLine: start.Line,
			StartCol:  start.Column,
			EndLine:   end.Line,
			EndCol:    end.Column,
		},
		Name:          fdecl.Name.Name,
		OwnerTypeName: ownerTypeName,
		Ptr:           ptr,
	}
}

func buildClosure(fset *token.FileSet, fn *goast.FuncLit) *model.FuncAnnotation {
	start := fset.Position(fn.Pos())
	end := fset.Position(fn.End())
	return &model.FuncAnnotation{
		Block: &model.Block{
			StartLine: start.Line,
			StartCol:  start.Column,
			EndLine:   end.Line,
			EndCol:    end.Column,
		},
		Closure: true,
	}
}
