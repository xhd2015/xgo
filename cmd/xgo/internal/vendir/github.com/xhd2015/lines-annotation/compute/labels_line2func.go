package compute

import "github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/model"

// depends: File.Lines.Labels
// generate: FuncAnnotation.Labels
func Labels_Line2Func(project *model.ProjectAnnotation) {
	for _, file := range project.Files {
		funcLabels := make(map[model.FuncID]map[string]bool, len(file.Funcs))
		for _, line := range file.Lines {
			if line.FuncID == "" {
				continue
			}
			labels := funcLabels[line.FuncID]
			if labels == nil {
				labels = make(map[string]bool, len(line.Labels))
				funcLabels[line.FuncID] = labels
			}
			for label, v := range line.Labels {
				if v {
					labels[label] = true
				}
			}
		}

		if file.Funcs == nil {
			file.Funcs = make(model.FuncAnnotationMapping, len(funcLabels))
		}
		for funcID, labels := range funcLabels {
			fn := file.Funcs[funcID]
			if fn == nil {
				fn = &model.FuncAnnotation{}
				file.Funcs[funcID] = fn
			}
			if fn.Labels == nil {
				fn.Labels = make(map[string]bool, len(labels))
			}
			for label := range labels {
				fn.Labels[label] = true
			}
		}
	}
	project.Set(model.AnnotationType_FuncLabels)
}
