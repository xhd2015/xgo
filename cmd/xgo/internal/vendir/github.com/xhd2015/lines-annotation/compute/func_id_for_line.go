package compute

import (
	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/model"
)

func EnsureFuncID_ForLine(project *model.ProjectAnnotation) {
	if project.Has(model.AnnotationType_LineFuncID) {
		return
	}
	FuncID_ForLine(project)
}

func FuncID_ForLine(project *model.ProjectAnnotation) {
	// requirement
	project.MustHave(string(model.AnnotationType_LineFuncID), model.AnnotationType_FileFuncs)

	for _, file := range project.Files {
		// we try to allocate a line to later block
		// if there is a nesting
		lineMapping := make(map[model.LineNum]model.FuncID)
		lineToBlock := make(map[model.LineNum]*model.Block)

		for fnID, fn := range file.Funcs {
			block := fn.Block
			for lineNum := block.StartLine; lineNum <= block.EndLine; lineNum++ {
				prevBlock := lineToBlock[model.LineNum(lineNum)]
				// prevBlock is greater that current block
				if prevBlock != nil && prevBlock.After(block) {
					continue
				}
				lineToBlock[model.LineNum(lineNum)] = block
				lineMapping[model.LineNum(lineNum)] = fnID
			}
		}

		lines := file.Lines
		if lines == nil {
			lines = make(model.LineAnnotationMapping, len(lineMapping))
			file.Lines = lines
		}

		for lineNum, funcID := range lineMapping {
			ann := lines[lineNum]
			if ann == nil {
				ann = &model.LineAnnotation{
					FuncID: funcID,
				}
				lines[lineNum] = ann
			} else {
				ann.FuncID = funcID
			}
		}
	}

	project.Set(model.AnnotationType_LineFuncID)
}
