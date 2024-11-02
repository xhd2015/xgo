package compute

import "github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/model"

// depends: BuildLineToFunctionMapping
func Changed_Line2Func(project *model.ProjectAnnotation) {
	// TODO: add pre-requiesite check
	for _, file := range project.Files {
		for _, line := range file.Lines {
			if line.FuncID != "" && line.Changed {
				fn := file.Funcs[line.FuncID]
				if fn == nil {
					if file.Funcs == nil {
						file.Funcs = make(model.FuncAnnotationMapping)
					}
					fn = &model.FuncAnnotation{
						Changed: line.Changed,
					}
					file.Funcs[line.FuncID] = fn
				} else {
					fn.Changed = fn.Changed || line.Changed
				}
			}
		}
	}
	project.Set(model.AnnotationType_FuncChanged)
}
