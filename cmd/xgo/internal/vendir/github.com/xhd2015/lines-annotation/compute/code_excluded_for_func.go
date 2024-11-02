package compute

import (
	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/model"
)

// Comments to func
func CodeExcludedForFunc(project *model.ProjectAnnotation) {
	project.MustHave("func code excluded ", model.AnnotationType_FuncCodeComments)
	for _, file := range project.Files {
		for _, fn := range file.Funcs {
			if fn.Code == nil || len(fn.Code.Comments) == 0 {
				continue
			}
			fn.Code.Excluded = shouldBeExcluded(fn.Code.Comments)
		}
	}
	project.Set(model.AnnotationType_FuncCodeExcluded)
}

func CodeExcludedFuncToLine(project *model.ProjectAnnotation) {
	project.MustHave("line code excluded", model.AnnotationType_FuncCodeExcluded, model.AnnotationType_LineFuncID, model.AnnotationType_FileFuncs)
	for _, file := range project.Files {
		for _, line := range file.Lines {
			funcData := file.Funcs[line.FuncID]
			if funcData == nil || funcData.Code == nil || !funcData.Code.Excluded {
				continue
			}
			if line.Code == nil {
				line.Code = &model.CodeAnnotation{}
			}
			line.Code.Excluded = true
		}
	}
	project.Set(model.AnnotationType_LineCodeExcluded)
}

func EnsureCodeExcludedFuncToLine(project *model.ProjectAnnotation) {
	if project.Has(model.AnnotationType_LineCodeExcluded) {
		return
	}
	EnsureCodeExcludedForFunc(project)
	EnsureFuncID_ForLine(project)

	CodeExcludedFuncToLine(project)
}

// file funcs needs to be loaded
// func EnsureFileFuncs(project *model.ProjectAnnotation) {
// 	if project.Has(model.AnnotationType_FileFuncs) {
// 		return
// 	}
// }

func EnsureCodeExcludedForFunc(project *model.ProjectAnnotation) {
	if project.Has(model.AnnotationType_FuncCodeExcluded) {
		return
	}
	CodeExcludedForFunc(project)
}

func shouldBeExcluded(comments map[model.CommentLabel]*model.CodeComment) bool {
	for _, k := range excludedTypes {
		_, ok := comments[k]
		if ok {
			return true
		}
	}
	return false
}
