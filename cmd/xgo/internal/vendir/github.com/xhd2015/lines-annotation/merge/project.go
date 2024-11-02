package merge

import (
	"fmt"

	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/model"
)

func MergeAnnotations(annotations ...*model.ProjectAnnotation) *model.ProjectAnnotation {
	return MergeProjectAnnotation(annotations)
}

func MergeProjectAnnotation(annotations []*model.ProjectAnnotation) *model.ProjectAnnotation {
	if len(annotations) == 0 {
		return nil
	}
	if len(annotations) == 1 {
		return annotations[0]
	}
	res := &model.ProjectAnnotation{
		Files: make(model.FileAnnotationMapping),
		Types: make(map[model.AnnotationType]bool),
	}
	MergeAnnotationsInto(res, annotations...)
	return res
}

func MergeAnnotationsInto(res *model.ProjectAnnotation, annotations ...*model.ProjectAnnotation) {
	if res == nil {
		panic(fmt.Errorf("res should not be nil"))
	}
	for _, src := range annotations {
		if src == nil {
			continue
		}
		if res.Files == nil {
			res.Files = make(model.FileAnnotationMapping, len(src.Files))
		}
		MergeFiles(res.Files, src.Files)

		if res.Types == nil {
			res.Types = make(map[model.AnnotationType]bool, len(src.Types))
		}
		MergeTypes(res.Types, src.Types)
	}

}

func MergeTypes(dst map[model.AnnotationType]bool, src map[model.AnnotationType]bool) {
	for typ := range src {
		dst[typ] = true
	}
}
