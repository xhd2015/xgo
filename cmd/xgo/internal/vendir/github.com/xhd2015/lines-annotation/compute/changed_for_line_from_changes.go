package compute

import (
	"fmt"

	diff "github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/go-coverage/diff/vscode"

	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/model"
)

func Changed_ForLineFromChanges(project *model.ProjectAnnotation) {
	if !project.Has(model.AnnotationType_LineChanges) {
		panic(fmt.Errorf("computing line changed, missing line changes"))
	}
	for _, file := range project.Files {
		lines := file.Lines
		if lines == nil {
			lines = make(model.LineAnnotationMapping)
			file.Lines = lines
		} else {
			// clear previous result
			for _, lineData := range lines {
				lineData.Changed = false
			}
		}
		if file.LineChanges == nil {
			continue
		}
		diff.ForeachLineMapping(file.LineChanges.Changes, int(file.LineChanges.OldLineCount), int(file.LineChanges.NewLineCount), func(oldLineStart, oldLineEnd, newLineStart, newLineEnd int, changeType diff.ChangeType) {
			if changeType == diff.ChangeTypeUpdate || changeType == diff.ChangeTypeInsert {
				for i := newLineStart; i < newLineEnd; i++ {
					lineNum := model.LineNum(i)
					lineData := lines[lineNum]
					if lineData == nil {
						lineData = &model.LineAnnotation{}
						lines[lineNum] = lineData
					}
					lineData.Changed = true
				}
			} else if changeType == diff.ChangeTypeUnchange {
				// newLineEnd - newLineStart == oldLineEnd - oldLineStart
				j := oldLineStart
				for i := newLineStart; i < newLineEnd; i++ {
					lineNum := model.LineNum(i)
					lineData := lines[lineNum]
					if lineData == nil {
						lineData = &model.LineAnnotation{}
						lines[lineNum] = lineData
					}
					lineData.OldLine = int64(j)
					j++
				}
			}
		})
	}

	project.Set(model.AnnotationType_LineChanged)
	project.Set(model.AnnotationType_LineOldLine)
}
