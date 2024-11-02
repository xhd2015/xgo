package compute

import (
	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/model"
)

type MatchMode string

const (
	MatchMode_Exact MatchMode = "" // default is exact
	MatchMode_Any   MatchMode = "any"
)

type LabelOption struct {
	DisplayName string // when not empty, use it instead
	Alias       map[string]bool
	MatchMode   MatchMode // need match exec label
}

func EnsureCoverageLabels_ForLine(project *model.ProjectAnnotation, labelOptions map[string]*LabelOption) {
	if project.Has(model.AnnotationType_LineCoverageLabels) {
		return
	}
	EnsureExecLabels_Block2Line(project)
	CoverageLabels_ForLine(project, labelOptions)
}

// optional: AnnotationType_LineLabels
func CoverageLabels_ForLine(project *model.ProjectAnnotation, labelOptions map[string]*LabelOption) {
	project.MustHave(string(model.AnnotationType_LineCoverageLabels), model.AnnotationType_LineExecLabels)

	// hasLineLabels := project.Has(model.AnnotationType_LineLabels)

	// collect labels, adding default labels
	labels := collectMarkLabels(project)
	labels[""] = true
	for label := range labelOptions {
		labels[label] = true
	}

	hasLabelMark := func(line *model.LineAnnotation, label string) bool {
		return checkHasLabelMark(line.Labels, label, labelOptions[label])
	}
	hasExecLabel := func(line *model.LineAnnotation, label string) bool {
		return checkHasExecLabel(line.ExecLabels, label, labelOptions[label])
	}
	for _, fileData := range project.Files {
		for _, lineData := range fileData.Lines {
			for label := range labels {
				if hasLabelMark(lineData, label) {
					if lineData.CoverageLabels == nil {
						lineData.CoverageLabels = make(map[string]bool)
					}
					lineData.CoverageLabels[label] = hasExecLabel(lineData, label)
				}
			}
		}
	}
	project.Set(model.AnnotationType_LineCoverageLabels)
}

func collectMarkLabels(project *model.ProjectAnnotation) map[string]bool {
	m := make(map[string]bool)
	for _, file := range project.Files {
		for _, line := range file.Lines {
			for label, v := range line.Labels {
				if label == "" {
					continue
				}
				if v {
					m[label] = true
				}
			}
		}
	}
	return m
}

func checkHasLabelMark(labels map[string]bool, label string, opts *LabelOption) bool {
	if label == "" {
		return true
	}
	if opts != nil && len(opts.Alias) > 0 {
		for labelAlias := range opts.Alias {
			if labels[labelAlias] {
				return true
			}
		}
		return false
	}
	return labels[label]
}

func checkHasExecLabel(execLabels map[string]bool, label string, opts *LabelOption) bool {
	if opts != nil && opts.MatchMode != MatchMode_Exact {
		if opts.MatchMode == MatchMode_Any {
			return len(execLabels) > 0
		}
		return false
	}
	if opts != nil && len(opts.Alias) > 0 {
		for labelAlias := range opts.Alias {
			if execLabels[labelAlias] {
				return true
			}
		}
		return false
	}
	return execLabels[label]
}
