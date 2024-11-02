package load

import (
	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/go-coverage/line"
	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/ast"
	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/model"
	"github.com/xhd2015/xgo/support/fileutil"
)

func LoadLineEmptyFromAstInfo(astLoadInfo ast.LoadInfo) *model.ProjectAnnotation {
	project := &model.ProjectAnnotation{
		Files: make(model.FileAnnotationMapping),
	}
	astLoadInfo.RangeFiles(func(f ast.File) bool {
		emptyLines := line.CollectEmptyLinesForFile(astLoadInfo.FileSet(), f.Ast())
		lines := make(model.LineAnnotationMapping, len(emptyLines))
		for _, line := range emptyLines {
			lines[model.LineNum(line)] = &model.LineAnnotation{
				Empty: true,
			}
		}

		project.Files[model.RelativeFile(fileutil.Slashlize(f.RelPath()))] = &model.FileAnnotation{
			Lines: lines,
		}
		return true
	})
	project.Set(model.AnnotationType_LineEmpty)
	return project
}
