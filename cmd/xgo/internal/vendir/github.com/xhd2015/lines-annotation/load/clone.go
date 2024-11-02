package load

import (
	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/merge"
	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/model"
)

func Clone(project *model.ProjectAnnotation) *model.ProjectAnnotation {
	if project == nil {
		return nil
	}
	res := &model.ProjectAnnotation{
		Files: make(model.FileAnnotationMapping),
		Types: make(map[model.AnnotationType]bool),
	}
	merge.MergeAnnotationsInto(res, project)
	return res
}
