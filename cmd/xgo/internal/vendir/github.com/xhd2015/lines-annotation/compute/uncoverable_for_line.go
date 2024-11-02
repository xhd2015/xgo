package compute

import "github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/model"

func EnsureUncoverable_ForLine(project *model.ProjectAnnotation) {
	if project.Has(model.AnnotationType_LineUncoverable) {
		return
	}
	Uncoverable_ForLine(project)
}

func Uncoverable_ForLine(project *model.ProjectAnnotation) {
	if project == nil {
		return
	}
	for _, fileData := range project.Files {
		for _, lineData := range fileData.Lines {
			lineData.Uncoverable = LineUncoverable(lineData)
		}
	}
	project.Set(model.AnnotationType_LineUncoverable)
}

func LineUncoverable(lineData *model.LineAnnotation) bool {
	return lineData.Empty || lineData.BlockID == ""
}

var excludedTypes = []model.CommentLabel{
	model.CommentLabels_Deprecated,
	model.CommentLabels_NoCov,
	model.CommentLabels_Unreachable,
}
