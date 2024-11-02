package coverage

import (
	"fmt"
	"math"
	"strconv"

	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/compute"
	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/model"
	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/model/coverage"
)

type ComputeOptions struct {
	LabelOptions      map[string]*compute.LabelOption
	DisableFunc       bool
	NeedUncoveredList bool
}

// depends: Line.CoverageLabels
func ComputeCoverageSummary(project *model.ProjectAnnotation, opts *ComputeOptions) map[string]*coverage.Summary {
	compute.LabelsOnTheFly(project)
	if opts == nil {
		opts = &ComputeOptions{}
	}
	needUncoveredList := opts.NeedUncoveredList

	if !project.Has(model.AnnotationType_FuncChanged) {
		compute.Changed_Line2Func(project)
	}

	compute.EnsureCoverageLabels_ForLine(project, opts.LabelOptions)
	if !opts.DisableFunc {
		compute.EnsureCoverageLabels_Line2Func(project)
	}

	if !project.Has(model.AnnotationType_LineUncoverable) {
		compute.Uncoverable_ForLine(project)
	}
	if !opts.DisableFunc {
		compute.EnsureCodeExcludedForFunc(project)
		compute.EnsureCodeExcludedFuncToLine(project)
	}

	// collect labels, adding default labels
	lineCoverages := compulteLineCoverage(project, needUncoveredList)
	var funcCoverages map[string]*coverage.Detail
	if !opts.DisableFunc {
		funcCoverages = compulteFuncCoverage(project, needUncoveredList)
	}

	summary := make(map[string]*coverage.Summary, len(lineCoverages))
	for label := range lineCoverages {
		labelDisplayName := label

		labelOpts := opts.LabelOptions[label]
		if labelOpts != nil && labelOpts.DisplayName != "" {
			labelDisplayName = labelOpts.DisplayName
		}

		lineState := lineCoverages[label]
		funcState := funcCoverages[label]
		sum := &coverage.Summary{
			Detail: lineState,
			Details: map[coverage.Mode]*coverage.Detail{
				// default is line
				coverage.Mode_Line: lineState,
			},
		}
		if funcState != nil {
			sum.Details[coverage.Mode_Func] = funcState
		}
		summary[labelDisplayName] = sum
	}
	updateSummary(summary)
	return summary
}

func updateSummary(summary map[string]*coverage.Summary) {
	// update percent
	for _, sum := range summary {
		if sum.Detail != nil {
			UpdateCoverageValue(sum.Detail)
		}
		for _, detail := range sum.Details {
			UpdateCoverageValue(detail)
		}
	}
}

func div(a, b int) float64 {
	if b == 0 {
		return 1 // make NaN = 100%
	}
	return math.Round(float64(a)/float64(b)*10000) / 10000
}
func divStr(a, b int) string {
	if b == 0 {
		return "1" // make NaN = 100%
	}
	return fmt.Sprintf("%.2f", float64(a)/float64(b))
}

func UpdateCoverageValue(detail *coverage.Detail) {
	updateDetail(detail)
}

func updateDetail(detail *coverage.Detail) {
	if detail.Total != nil {
		detail.Total.ValueNum = div(detail.Total.Covered, detail.Total.Total)
		detail.Total.Value = strconv.FormatFloat(detail.Total.ValueNum, 'f', 4, 64)
	}
	if detail.Incrimental != nil {
		detail.Incrimental.ValueNum = div(detail.Incrimental.Covered, detail.Incrimental.Total)
		detail.Incrimental.Value = strconv.FormatFloat(detail.Incrimental.ValueNum, 'f', 4, 64)
	}
}

func compulteLineCoverage(project *model.ProjectAnnotation, needUncoveredList bool) map[string]*coverage.Detail {
	mapping := make(map[string]*coverage.Detail)
	for file, fileData := range project.Files {
		for line, lineData := range fileData.Lines {
			if LineUncoverableOrExcluded(lineData) {
				continue
			}
			// no need to record uncovered list
			addCoverage(mapping, lineData.CoverageLabels, LineChanged(fileData, lineData), needUncoveredList, string(file), int(line), "")
		}
	}
	return mapping
}
func addCoverage(mapping map[string]*coverage.Detail, coverageLabels map[string]bool, changed bool, needUncoveredList bool, file string, lineNum int, funcName string) {
	for label, covered := range coverageLabels {
		mapping[label] = UpdateCoverageDetail(label, covered, changed, mapping[label], needUncoveredList, file, lineNum, funcName)
	}
}
func LineChanged(fileData *model.FileAnnotation, lineData *model.LineAnnotation) bool {
	if fileData != nil && fileData.ChangeDetail != nil && fileData.ChangeDetail.IsNew {
		return true
	}
	return lineData != nil && lineData.Changed
}

func LineUncoverableOrExcluded(lineData *model.LineAnnotation) bool {
	return lineData.Uncoverable || (lineData.Code != nil && lineData.Code.Excluded) || (lineData.Remark != nil && lineData.Remark.Excluded)
}

func UpdateCoverageDetail(label string, covered bool, changed bool, detail *coverage.Detail, needUncoveredList bool, file string, lineNum int, funcName string) *coverage.Detail {
	if detail == nil {
		detail = &coverage.Detail{
			Total:       &coverage.Item{},
			Incrimental: &coverage.Item{},
		}
	}
	detail.Total.Total++
	if changed {
		detail.Incrimental.Total++
	}
	if covered {
		detail.Total.Covered++
		if changed {
			detail.Incrimental.Covered++
		}
	} else if needUncoveredList && file != "" {
		item := &coverage.UncoveredItem{
			File: file,
			Line: lineNum,
			Func: funcName,
		}
		detail.Total.UncoveredList = append(detail.Total.UncoveredList, item)
		if changed {
			detail.Incrimental.UncoveredList = append(detail.Incrimental.UncoveredList, item)
		}
	}
	return detail
}

func compulteFuncCoverage(project *model.ProjectAnnotation, needUncoveredList bool) map[string]*coverage.Detail {
	mapping := make(map[string]*coverage.Detail)
	for file, fileData := range project.Files {
		for _, funcData := range fileData.Funcs {
			if (funcData.Code != nil && funcData.Code.Excluded) || funcData.FirstLineExcluded {
				continue
			}
			changed := false
			if project.Has(model.AnnotationType_ChangeDetail) {
				changed = fileData.ChangeDetail.IsNew || funcData.Changed
			}

			// TODO: string(file), funcData.Block.StartLine
			var line int
			if funcData.Block != nil {
				line = funcData.Block.StartLine
			}
			addCoverage(mapping, funcData.CoverageLabels, changed, needUncoveredList, string(file), line, funcData.Name)
		}
	}
	return mapping
}
