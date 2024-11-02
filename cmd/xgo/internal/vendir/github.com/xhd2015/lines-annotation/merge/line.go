package merge

import "github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/model"

func MergeLines(dst model.LineAnnotationMapping, src model.LineAnnotationMapping) {
	for lineNum, srcAnn := range src {
		dstAnn := dst[lineNum]
		if dstAnn == nil {
			dst[lineNum] = srcAnn
			continue
		}
		MergeLine(dstAnn, srcAnn)
	}
}

func MergeLine(dst *model.LineAnnotation, src *model.LineAnnotation) {
	dst.Changed = dst.Changed || src.Changed
	dst.Empty = dst.Empty || src.Empty
	if dst.BlockID == "" {
		dst.BlockID = src.BlockID
	}
	if dst.FuncID == "" {
		dst.FuncID = src.FuncID
	}
	if dst.Labels == nil {
		dst.Labels = make(map[string]bool, len(src.Labels))
	}
	MergeStringSet(dst.Labels, src.Labels)

	if dst.ExecLabels == nil {
		dst.ExecLabels = make(map[string]bool, len(src.ExecLabels))
	}
	MergeStringSet(dst.ExecLabels, src.ExecLabels)

	if dst.CoverageLabels == nil {
		dst.CoverageLabels = make(map[string]bool, len(src.CoverageLabels))
	}
	MergeMap(dst.CoverageLabels, src.CoverageLabels)

	// merge remark using replace
	if dst.Remark == nil {
		dst.Remark = src.Remark
	}

	if dst.Code == nil {
		dst.Code = src.Code
	} else {
		MergeCode(dst.Code, src.Code)
	}
}

func MergeStringSet(dst map[string]bool, src map[string]bool) {
	for key := range src {
		dst[key] = true
	}
}
func MergeMap(dst map[string]bool, src map[string]bool) {
	for key := range src {
		dst[key] = true
	}
}
