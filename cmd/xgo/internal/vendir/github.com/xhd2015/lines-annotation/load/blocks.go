package load

import (
	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/ast"
	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/ast/ast2ann"
	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/model"
	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/model/coverage"
	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/path"
)

func BlocksFromAstInfo(astLoadInfo ast.LoadInfo) (*model.ProjectAnnotation, error) {
	blockProfile, err := ast2ann.CollectBlockProfilesNoModPath(astLoadInfo)
	if err != nil {
		return nil, err
	}
	return binaryProfileToAnnotation(blockProfile, nil), nil
}

func BinaryProfileToAnnotation(modPath string, profile coverage.BinaryProfile) *model.ProjectAnnotation {
	return binaryProfileToAnnotation(profile, func(file string) string {
		return path.TrimModPrefixOrEmpty(file, modPath)
	})
}

func BinaryProfileTrimmedToAnnotation(profile coverage.BinaryProfile) *model.ProjectAnnotation {
	return binaryProfileToAnnotation(profile, nil)
}

func binaryProfileToAnnotation(profile coverage.BinaryProfile, trimFile func(file string) string) *model.ProjectAnnotation {
	if len(profile) == 0 {
		return nil
	}
	files := make(model.FileAnnotationMapping, len(profile))
	for file, blockStats := range profile {
		rfile := file
		if trimFile != nil {
			rfile = trimFile(rfile)
		}
		if rfile == "" {
			// filter empty files
			// fix non-main-module files
			continue
		}

		files[model.RelativeFile(rfile)] = &model.FileAnnotation{
			Blocks: BlockStatsToAnnotationMapping(blockStats),
		}
	}
	return &model.ProjectAnnotation{
		Files: files,
		Types: map[model.AnnotationType]bool{
			model.AnnotationType_Blocks: true,
		},
	}
}

func BlockStatsToAnnotationMapping(blockStats []*coverage.BlockStats) model.BlockAnnotationMapping {
	blocks := make(model.BlockAnnotationMapping, len(blockStats))
	for _, blockStat := range blockStats {
		blocks[blockStat.ID()] = BlockStatsToAnnotation(blockStat)
	}
	return blocks
}

func BlockStatsToAnnotation(blockStats *coverage.BlockStats) *model.BlockAnnotation {
	block := blockStats.Block

	labels := make(map[string]bool, len(blockStats.Count))
	for label, val := range blockStats.Count {
		if val > 0 {
			labels[label] = true
		}
	}
	return &model.BlockAnnotation{
		Block:      block,
		ExecLabels: labels,
	}
}
