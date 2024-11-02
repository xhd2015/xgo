package merge

import "github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/model"

func MergeBlocks(dst model.BlockAnnotationMapping, src model.BlockAnnotationMapping) {
	for blockID, srcAnn := range src {
		dstAnn := dst[blockID]
		if dstAnn == nil {
			dst[blockID] = srcAnn
			continue
		}
		MergeBlock(dstAnn, srcAnn)
	}
}

func MergeBlock(dst *model.BlockAnnotation, src *model.BlockAnnotation) {
	if dst.Block == nil {
		dst.Block = src.Block
	}
	if dst.FuncID == "" {
		dst.FuncID = src.FuncID
	}

	if dst.ExecLabels == nil {
		dst.ExecLabels = make(map[string]bool, len(src.ExecLabels))
	}
	MergeStringSet(dst.ExecLabels, src.ExecLabels)
}
