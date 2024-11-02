package compute

import "github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/model"

func EnsureExecLabels_Block2Line(project *model.ProjectAnnotation) {
	if project.Has(model.AnnotationType_LineExecLabels) {
		return
	}
	EnsureBlockID_ForLine(project)
	ExecLabels_Block2Line(project)
}

// from: File.Blocks.ExecLabels, File.Line.BlockID
// generate: File.Lines.ExecLabels
func ExecLabels_Block2Line(project *model.ProjectAnnotation) {
	project.MustHave(string(model.AnnotationType_LineExecLabels), model.AnnotationType_LineBlockID)
	for _, file := range project.Files {
		for _, line := range file.Lines {
			block := file.Blocks[line.BlockID]
			if block == nil || len(block.ExecLabels) == 0 {
				continue
			}

			if line.ExecLabels == nil {
				line.ExecLabels = make(map[string]bool, len(block.ExecLabels))
			}
			for label, val := range block.ExecLabels {
				if val {
					line.ExecLabels[label] = true
				}
			}
		}
	}
	project.Set(model.AnnotationType_LineExecLabels)
}
