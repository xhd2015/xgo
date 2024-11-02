package compute

import "github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/model"

// depends: File.Lines.ExecLabels
// generate: FuncAnnotation.ExecLabels
func ExecLabels_Line2Func(project *model.ProjectAnnotation) {
	for _, file := range project.Files {
		funcExecLabels := make(map[model.FuncID]map[string]bool, len(file.Funcs))
		for _, line := range file.Lines {
			if line.FuncID == "" {
				continue
			}
			execLabels := funcExecLabels[line.FuncID]
			if execLabels == nil {
				execLabels = make(map[string]bool, len(line.ExecLabels))
				funcExecLabels[line.FuncID] = execLabels
			}
			for label, v := range line.ExecLabels {
				if v {
					execLabels[label] = true
				}
			}
		}

		if file.Funcs == nil {
			file.Funcs = make(model.FuncAnnotationMapping, len(funcExecLabels))
		}
		for funcID, labels := range funcExecLabels {
			fn := file.Funcs[funcID]
			if fn == nil {
				fn = &model.FuncAnnotation{}
				file.Funcs[funcID] = fn
			}
			if fn.ExecLabels == nil {
				fn.ExecLabels = make(map[string]bool, len(labels))
			}
			for label := range labels {
				fn.ExecLabels[label] = true
			}
		}
	}

	project.Set(model.AnnotationType_FuncExecLabels)
}
