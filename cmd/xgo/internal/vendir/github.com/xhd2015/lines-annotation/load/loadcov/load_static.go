package loadcov

import (
	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/ast"
	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/ast/ast2ann"
	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/compute"
	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/load"
	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/merge"
	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/model"
)

type LodOpts struct {
	LoadFuncInfo bool
}

func LoadStatic(astLoadInfo ast.LoadInfo, opts LodOpts) (*model.ProjectAnnotation, error) {
	var funcInfo model.FuncInfoMapping
	var funcComments *model.ProjectAnnotation

	lineEmptyAnnotation := load.LoadLineEmptyFromAstInfo(astLoadInfo)

	trimedEmptyProfile, err := ast2ann.CollectBlockProfilesNoModPath(astLoadInfo)
	if err != nil {
		return nil, err
	}

	if opts.LoadFuncInfo {
		funcInfo, err = load.CollectFuncs(astLoadInfo)
		if err != nil {
			return nil, err
		}

		funcComments = load.FuncCommentsFromAstInfoMapping(astLoadInfo, funcInfo)

		// pick comments with "labels" to "RC"
		compute.LineLabelsFromLineComment(funcComments)
	}

	// sort blocks
	trimedEmptyProfile.SortAll()

	funcInfoMapping := load.FuncInfoMappingToAnnotation(funcInfo)
	blockInfo := load.BinaryProfileTrimmedToAnnotation(trimedEmptyProfile)
	if blockInfo == nil {
		// if all files excluded, the block info can be nil
		blockInfo = &model.ProjectAnnotation{
			Types: map[model.AnnotationType]bool{
				model.AnnotationType_Blocks: true,
			},
		}
	}

	merge.MergeAnnotationsInto(blockInfo, lineEmptyAnnotation, funcInfoMapping, funcComments)
	return blockInfo.Simplified(), nil
}
