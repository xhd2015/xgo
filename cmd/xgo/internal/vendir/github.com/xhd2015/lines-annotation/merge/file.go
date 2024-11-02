package merge

import "github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/model"

func MergeFiles(dst model.FileAnnotationMapping, src model.FileAnnotationMapping) {
	for file, srcAnn := range src {
		dstAnn := dst[file]
		if dstAnn == nil {
			dst[file] = srcAnn
			continue
		}
		MergeFile(dstAnn, srcAnn)
	}
}

func MergeFile(dst *model.FileAnnotation, src *model.FileAnnotation) {
	if dst.ChangeDetail == nil {
		dst.ChangeDetail = src.ChangeDetail
	}
	if dst.Lines == nil {
		dst.Lines = make(model.LineAnnotationMapping)
	}
	MergeLines(dst.Lines, src.Lines)

	if dst.LineChanges == nil {
		dst.LineChanges = src.LineChanges
	}

	if dst.DeletedLines == nil {
		dst.DeletedLines = src.DeletedLines
	}

	if dst.Blocks == nil {
		dst.Blocks = make(model.BlockAnnotationMapping)
	}

	MergeBlocks(dst.Blocks, src.Blocks)

	if dst.Funcs == nil {
		dst.Funcs = make(model.FuncAnnotationMapping)
	}
	MergeFuncs(dst.Funcs, src.Funcs)
}
