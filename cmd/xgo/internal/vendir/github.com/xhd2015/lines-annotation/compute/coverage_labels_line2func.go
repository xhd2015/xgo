package compute

import "github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/model"

func EnsureCoverageLabels_Line2Func(project *model.ProjectAnnotation) {
	if project.Has(model.AnnotationType_FuncCoverageLabels) {
		return
	}
	EnsureFuncID_ForLine(project)
	// give an empty one
	EnsureCoverageLabels_ForLine(project, nil)
	CoverageLabels_Line2Func(project)
}

func CoverageLabels_Line2Func(project *model.ProjectAnnotation) {
	// requirement
	project.MustHave(string(model.AnnotationType_FuncCoverageLabels), model.AnnotationType_LineFuncID, model.AnnotationType_LineCoverageLabels)

	for _, file := range project.Files {
		funcCoverageLabels := make(map[model.FuncID]map[string]bool, len(file.Funcs))
		for _, line := range file.Lines {
			if line.FuncID == "" {
				continue
			}
			funcLabels := funcCoverageLabels[line.FuncID]
			if funcLabels == nil {
				funcLabels = make(map[string]bool, len(line.CoverageLabels))
				funcCoverageLabels[line.FuncID] = funcLabels
			}
			for label, v := range line.CoverageLabels {
				// merge using OR
				funcLabels[label] = funcLabels[label] || v
			}
		}

		if file.Funcs == nil {
			file.Funcs = make(model.FuncAnnotationMapping, len(funcCoverageLabels))
		}
		for funcID, labels := range funcCoverageLabels {
			fn := file.Funcs[funcID]
			if fn == nil {
				fn = &model.FuncAnnotation{}
				file.Funcs[funcID] = fn
			}
			if fn.CoverageLabels == nil {
				fn.CoverageLabels = make(map[string]bool, len(labels))
			}
			fn.CoverageLabels = labels
		}
	}
	project.Set(model.AnnotationType_FuncCoverageLabels)
}
