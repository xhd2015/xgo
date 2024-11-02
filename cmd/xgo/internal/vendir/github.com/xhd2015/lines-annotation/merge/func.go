package merge

import (
	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/model"
)

func MergeFuncs(dst model.FuncAnnotationMapping, src model.FuncAnnotationMapping) {
	for funcID, srcAnn := range src {
		dstAnn := dst[funcID]
		if dstAnn == nil {
			dst[funcID] = srcAnn
			continue
		}
		MergeFunc(dstAnn, srcAnn)
	}
}

func MergeFunc(dst *model.FuncAnnotation, src *model.FuncAnnotation) {
	if dst.Block == nil {
		dst.Block = src.Block
	}
	dst.Changed = dst.Changed || src.Changed
	dst.FirstLineExcluded = dst.FirstLineExcluded || src.FirstLineExcluded

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

	if dst.Code == nil {
		dst.Code = src.Code
	} else {
		MergeCode(dst.Code, src.Code)
	}
}

func MergeCode(dst *model.CodeAnnotation, src *model.CodeAnnotation) {
	if src == nil {
		return
	}
	dst.Excluded = dst.Excluded || src.Excluded

	if dst.Comments == nil {
		if len(src.Comments) == 0 {
			return
		}
		dst.Comments = make(map[model.CommentLabel]*model.CodeComment, len(src.Comments))
	}
	for prop, comment := range src.Comments {
		dstComment := dst.Comments[prop]
		if dstComment == nil {
			dst.Comments[prop] = comment.Clone()
			// dstComment = comment
		} else {
			if dstComment.Author == "" {
				dstComment.Author = comment.Author
			}
			for _, v := range comment.Values {
				if !strListContains(dstComment.Values, v) {
					dstComment.Values = append(dstComment.Values, v)
				}
			}
		}
	}
}

func strListContains(set []string, v string) bool {
	for _, s := range set {
		if s == v {
			return true
		}
	}
	return false
}
