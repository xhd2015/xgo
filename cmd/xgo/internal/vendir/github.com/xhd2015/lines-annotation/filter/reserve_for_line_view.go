package filter

import (
	"fmt"

	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/compute"
	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/model"
)

type MissingDiffFileOption int

const (
	MissingDiffOption_Error MissingDiffFileOption = 0
	// when will this be useful? a go file is invovled in compilation, but excluded by git,
	// for example, local debug files
	MissingDiffOption_AsUnchanged MissingDiffFileOption = 1
)

type ReserveOptions struct {
	ChangedOnly           bool
	MissingDiffFileOption MissingDiffFileOption
}

func ReserveForLineView(project *model.ProjectAnnotation, opts *ReserveOptions) *model.ProjectAnnotation {
	if project == nil {
		return nil
	}
	// project.MustHave("reserve for line view", model.AnnotationType_LineChanges, model.AnnotationType_LineChanged)

	changedOnly := false
	if opts != nil {
		changedOnly = opts.ChangedOnly
	}
	res := &model.ProjectAnnotation{
		Files: make(model.FileAnnotationMapping, len(project.Files)),
	}
	for file, fileData := range project.Files {
		fileChanged := false
		if project.Has(model.AnnotationType_ChangeDetail) {
			if fileData.ChangeDetail == nil {
				if opts != nil && opts.MissingDiffFileOption == MissingDiffOption_AsUnchanged {
					// ignore
				} else {
					panic(fmt.Errorf("missing diff: %s", file))
				}
			} else {
				fileChanged = fileData.ChangeDetail.IsNew || fileData.ChangeDetail.ContentChanged
			}
		}
		if changedOnly && !fileChanged {
			continue
		}
		lines := make(model.LineAnnotationMapping, len(fileData.Lines))
		for line, lineData := range fileData.Lines {
			// TODO: verify that the project has these data
			if changedOnly && !fileChanged && !lineData.Changed {
				continue
			}
			ann := &model.LineAnnotation{
				OldLine:        lineData.OldLine,
				Changed:        lineData.Changed,
				CoverageLabels: lineData.CoverageLabels,
				Uncoverable:    compute.LineUncoverable(lineData),
				Remark:         lineData.Remark, // inherit remark
			}
			if lineData.Code != nil && lineData.Code.Excluded {
				ann.Code = &model.CodeAnnotation{
					Excluded: true,
				}
			}
			lines[line] = ann
		}
		res.Files[file] = &model.FileAnnotation{
			Lines:        lines,
			ChangeDetail: fileData.ChangeDetail,
		}
	}
	res.Set(model.AnnotationType_LineOldLine)
	res.Set(model.AnnotationType_LineChanged)
	res.Set(model.AnnotationType_LineCoverageLabels)
	res.Set(model.AnnotationType_LineUncoverable)
	res.Set(model.AnnotationType_ChangeDetail)
	res.Set(model.AnnotationType_LineCodeExcluded)
	if project.Has(model.AnnotationType_LineRemark) {
		res.Set(model.AnnotationType_LineRemark)
	}

	return res
}
