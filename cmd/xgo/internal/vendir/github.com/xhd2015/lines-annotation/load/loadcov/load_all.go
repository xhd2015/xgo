package loadcov

import (
	_ "embed"
	"fmt"

	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/ast"
	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/compute"
	ann_filter "github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/filter"
	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/merge"
	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/model"
	filter_model "github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/model/filter"
	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/path/filter"
	"github.com/xhd2015/xgo/support/fileutil"
	"github.com/xhd2015/xgo/support/goinfo"
)

type LoadAllOptions struct {
	Dir      string
	Args     []string
	Profiles []string
	Ref      string
	DiffBase string

	// file filter
	Include []string
	Exclude []string

	OnlyChangedFiles bool
}

func LoadAll(opts LoadAllOptions) (*model.ProjectAnnotation, error) {
	if opts.Ref == "" {
		return nil, fmt.Errorf("requires ref")
	}
	if opts.DiffBase == "" {
		return nil, fmt.Errorf("requires diffBase")
	}
	projectDir := opts.Dir
	modPath, err := goinfo.GetModPath(projectDir)
	if err != nil {
		return nil, err
	}

	args := opts.Args
	if len(args) == 0 {
		args = []string{modPath + "/..."}
	}
	files, err := goinfo.ListRelativeFiles(projectDir, args)
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no files")
	}
	var hasFilter bool
	if len(opts.Include) > 0 || len(opts.Exclude) > 0 {
		hasFilter = true
		fileFilter := filter.NewFileFilter(opts.Include, opts.Exclude)
		i := 0
		for j := 0; j < len(files); j++ {
			file := files[j]
			if !fileFilter.MatchFile(fileutil.Slashlize(file)) {
				continue
			}
			files[i] = file
			i++
		}
		files = files[:i]
		if len(files) == 0 {
			return nil, fmt.Errorf("all files filtered")
		}
	}

	astInfo, err := ast.LoadFiles(projectDir, files)
	if err != nil {
		return nil, err
	}
	staticInfo, err := LoadStatic(astInfo, LodOpts{
		LoadFuncInfo: false,
	})
	if err != nil {
		return nil, err
	}

	// diff
	profileData, err := LoadCoverageProfileFiles(modPath, opts.Profiles, nil)
	if err != nil {
		return nil, err
	}

	if hasFilter {
		ann_filter.FilterFiles(profileData, &filter_model.Options{
			Include: opts.Include,
			Exclude: opts.Exclude,
		})
	}

	gitDiff, err := LoadGitDiff(projectDir, opts.Ref, opts.DiffBase, files)
	if err != nil {
		return nil, err
	}

	project := merge.MergeAnnotations(staticInfo, profileData, gitDiff)
	compute.Changed_ForLineFromChanges(project)

	compute.BlockID_ForLine(project)
	compute.ExecLabels_Block2Line(project)
	compute.Uncoverable_ForLine(project)
	compute.EnsureCoverageLabels_ForLine(project, nil)

	project = ann_filter.ReserveForLineView(project, &ann_filter.ReserveOptions{
		ChangedOnly:           opts.OnlyChangedFiles,
		MissingDiffFileOption: ann_filter.MissingDiffOption_AsUnchanged,
	})
	return project, nil
}
