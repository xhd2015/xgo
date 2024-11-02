package compute

import "github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/model"

// LabelsOnTheFly
// on the fly because these processes are lightweight
// the compute is simply, but the storage is massive
// so named with on the fly
func LabelsOnTheFly(project *model.ProjectAnnotation) {
	// already converted block to line
	// ExecLabels_Block2Line(project)
	ExecLabels_Line2Func(project)
	Labels_Line2Func(project)
}
