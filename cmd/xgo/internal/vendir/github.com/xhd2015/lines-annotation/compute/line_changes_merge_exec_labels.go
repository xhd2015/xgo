package compute

import (
	"fmt"

	diff "github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/go-coverage/diff/vscode"

	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/model"
)

// Three-way Merge(Phase 1)
// `base`: contains labels data, line's blockID
// `changesForBase`: contain mapping
// `oldData`: label data for older commit

func LineChangesMergeExecLabels(base *model.ProjectAnnotation, changesForBase *model.ProjectAnnotation, oldData *model.ProjectAnnotation) {
	lineChangesMergeExecLabels(base, changesForBase, oldData, true, true)
}
func LineChangesMergeExecLabelsOnlyChanged(base *model.ProjectAnnotation, changesForBase *model.ProjectAnnotation, oldData *model.ProjectAnnotation) {
	lineChangesMergeExecLabels(base, changesForBase, oldData, true, false)
}
func LineChangesMergeExecLabelsOnlyUnchanged(base *model.ProjectAnnotation, changesForBase *model.ProjectAnnotation, oldData *model.ProjectAnnotation) {
	lineChangesMergeExecLabels(base, changesForBase, oldData, false, true)
}

// LineChangesMergeLineRemark try best effort to map old line data to new line data
func LineChangesMergeLineRemark(base *model.ProjectAnnotation, changesForBase *model.ProjectAnnotation, oldData *model.ProjectAnnotation) {
	mapLineRemarks := func(baseLineData *model.LineAnnotation, oldLineData *model.LineAnnotation) {
		if baseLineData.Remark == nil {
			// set remark only if there is no new one
			baseLineData.Remark = oldLineData.Remark
		}
	}
	mapLineData(base, changesForBase, oldData, false /*disable changed range merging*/, true /*map unchanged*/, mapLineRemarks, false /*allow non-block line*/, false /*disable range change*/)
	base.Set(model.AnnotationType_LineRemark)
}

type mapLineCallback func(baseLineData *model.LineAnnotation, oldLineData *model.LineAnnotation)

// depends on:
//
//	`changesForBase.Lines[N].LineChanges``
//	`oldData.Lines[N].ExecLabels`
func lineChangesMergeExecLabels(base *model.ProjectAnnotation, changesForBase *model.ProjectAnnotation, oldData *model.ProjectAnnotation, includeChanged bool, includeUnchanged bool) {
	mapLineExecLabels := func(baseLineData *model.LineAnnotation, oldLineData *model.LineAnnotation) {
		if baseLineData.ExecLabels == nil {
			baseLineData.ExecLabels = make(map[string]bool)
		}

		oldLabels := oldLineData.ExecLabels
		for label, v := range oldLabels {
			if v {
				baseLineData.ExecLabels[label] = true
			}
		}
	}
	mapLineData(base, changesForBase, oldData, includeChanged, includeUnchanged, mapLineExecLabels, true, false)
}

// mapLineData: map old data of each line to new line
// `includedUnchanged`: map unchanged data
// `includeChanged`: map changed data
// `excludeNonBlockLine`: if a line does not have syntax block, then do not map it
// `includeChangedAsMapping`
func mapLineData(base *model.ProjectAnnotation, changesForBase *model.ProjectAnnotation, oldData *model.ProjectAnnotation, includeChanged bool, includeUnchanged bool, mapLine mapLineCallback, excludeNonBlockLine bool, includeChangedAsMapping bool) {
	if base == nil {
		panic(fmt.Errorf("base cannot be nil"))
	}
	if base.Files == nil {
		base.Files = make(model.FileAnnotationMapping)
	}
	for file, changeFile := range changesForBase.Files {
		changeDetail := changeFile.ChangeDetail
		if changeDetail == nil || changeDetail.IsNew || changeDetail.Deleted {
			continue
		}
		oldFile := file
		if changeDetail.RenamedFrom != "" {
			oldFile = model.RelativeFile(changeDetail.RenamedFrom)
		}

		oldFileData := oldData.Files[oldFile]
		if oldFileData == nil {
			continue
		}

		baseFile := base.Files[file]
		if baseFile == nil {
			baseFile = &model.FileAnnotation{
				Lines: make(model.LineAnnotationMapping),
			}
			base.Files[file] = baseFile
		}
		if !changeDetail.ContentChanged {
			if !includeUnchanged {
				continue
			}
			if baseFile.Lines == nil {
				baseFile.Lines = make(model.LineAnnotationMapping, len(oldFileData.Lines))
			}

			// unchanged, just copy
			for line, lineData := range oldFileData.Lines {
				baseLineData := baseFile.Lines[line]
				if baseLineData == nil {
					baseLineData = &model.LineAnnotation{}
					baseFile.Lines[line] = baseLineData
				}
				mapLine(baseLineData, lineData)
			}
			continue
		}
		lineChanges := changeFile.LineChanges
		if lineChanges == nil {
			continue
		}

		diff.ForeachLineMapping(lineChanges.Changes, int(lineChanges.OldLineCount), int(lineChanges.NewLineCount), MergeLabelsDiffCallback(&MergeOptions{
			IncludeChanged:          includeChanged,
			IncludeUnchanged:        includeUnchanged,
			IncludeChangedAsMapping: includeChangedAsMapping,
			ShouldIncludeLine: func(line int) bool {
				if baseFile == nil || baseFile.Lines == nil {
					return true
				}
				lineData := baseFile.Lines[model.LineNum(line)]
				if lineData == nil {
					// we don't risk filter unknown lines
					return true
				}
				// exclude lines that does not belong to any block
				return !excludeNonBlockLine || lineData.BlockID != ""
			},
			MergeLine: func(newLine, oldLineStart, oldLineEnd int64) {
				for i := oldLineStart; i < oldLineEnd; i++ {
					oldLineData := oldFileData.Lines[model.LineNum(i)]
					if oldLineData == nil {
						continue
					}
					baseLineData := baseFile.Lines[model.LineNum(newLine)]
					if baseLineData == nil {
						baseLineData = &model.LineAnnotation{}
						if baseFile.Lines == nil {
							baseFile.Lines = make(model.LineAnnotationMapping)
						}
						baseFile.Lines[model.LineNum(newLine)] = baseLineData
					}
					mapLine(baseLineData, oldLineData)
				}
			},
		}))
	}
	base.Set(model.AnnotationType_LineExecLabels)
}

type DiffCallback = func(oldLineStart, oldLineEnd, newLineStart, newLineEnd int, changeType diff.ChangeType)

type MergeOptions struct {
	IncludeUnchanged bool
	// IncludeChanged vs IncludeChangedAsMapping
	// range: oldA~oldB newA~newB
	// IncludeChanged:          each ln in newA~newB, map ln -> oldA~oldB
	// IncludeChangedAsMapping:  map common prefix of newA~newB and oldA~oldB
	IncludeChanged          bool
	IncludeChangedAsMapping bool
	ShouldIncludeLine       func(line int) bool

	// oldLineStart inclsuive,oldLineEnd exclusive
	MergeLine func(newLine int64, oldLineStart int64, oldLineEnd int64)
}

func MergeLabelsDiffCallback(opts *MergeOptions) DiffCallback {
	var includeChanged bool
	var includeUnchanged bool
	var includeChangedAsMapping bool
	var shouldIncludeLine func(line int) bool
	var mergeLine func(newLine int64, oldLineStart int64, oldLineEnd int64)
	if opts != nil {
		includeChanged = opts.IncludeChanged
		includeUnchanged = opts.IncludeUnchanged
		includeChangedAsMapping = opts.IncludeChangedAsMapping
		shouldIncludeLine = opts.ShouldIncludeLine
		mergeLine = opts.MergeLine
	}
	return func(oldLineStart, oldLineEnd, newLineStart, newLineEnd int, changeType diff.ChangeType) {
		if changeType == diff.ChangeTypeUnchange {
			if !includeUnchanged {
				return
			}
			// line-to-line merge
			n := newLineEnd - newLineStart
			for i := 0; i < n; i++ {
				line := newLineStart + i
				oldLine := oldLineStart + i
				if shouldIncludeLine != nil && !shouldIncludeLine(line) {
					// exclude non-block lines
					continue
				}
				mergeLine(int64(line), int64(oldLine), int64(oldLine+1))
			}
		} else if changeType == diff.ChangeTypeUpdate {
			if includeChanged {
				// line-range-to-line-range merge
				// assign labels
				for line := newLineStart; line < newLineEnd; line++ {
					if shouldIncludeLine != nil && !shouldIncludeLine(line) {
						// exclude non-block lines
						continue
					}
					mergeLine(int64(line), int64(oldLineStart), int64(oldLineEnd))
				}
				return
			}
			if includeChangedAsMapping {
				// as mapping
				// choose min
				n := newLineEnd - newLineStart
				m := oldLineEnd - oldLineStart
				if n > m {
					n = m
				}
				for i := 0; i < n; i++ {
					if shouldIncludeLine != nil && !shouldIncludeLine(newLineStart+i) {
						// exclude non-block lines
						continue
					}
					mergeLine(int64(newLineStart+i), int64(oldLineStart+i), int64(oldLineStart+i+1))
				}
				return
			}
		}
	}
}
