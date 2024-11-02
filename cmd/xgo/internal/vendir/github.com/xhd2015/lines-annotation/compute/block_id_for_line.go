package compute

import "github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/model"

func EnsureBlockID_ForLine(project *model.ProjectAnnotation) {
	if project.Has(model.AnnotationType_LineBlockID) {
		return
	}
	BlockID_ForLine(project)
}

// convert block labels to line labels
// the map: <line,label> -> true|false
// the map will not be very huge, but it will certainly repeat
// how to resolve that? id is just a trick to reduce repeation
func BlockID_ForLine(project *model.ProjectAnnotation) {
	project.MustHave(string(model.AnnotationType_LineBlockID), model.AnnotationType_Blocks)

	for _, file := range project.Files {
		// former block will win a line if that line
		// has overlapping

		// we try to allocate a line to later block
		// if there is a nesting
		lineMapping := make(map[model.LineNum]model.BlockID)
		lineToBlock := make(map[model.LineNum]*model.Block)
		for blockID, blockAnnotation := range file.Blocks {
			block := blockAnnotation.Block
			for lineNum := block.StartLine; lineNum <= block.EndLine; lineNum++ {
				prevBlock := lineToBlock[model.LineNum(lineNum)]
				// prevBlock is greater that current block
				if prevBlock != nil && prevBlock.Before(block) {
					continue
				}
				lineToBlock[model.LineNum(lineNum)] = block
				lineMapping[model.LineNum(lineNum)] = blockID
			}
		}

		lines := file.Lines
		if lines == nil {
			lines = make(model.LineAnnotationMapping, len(lineMapping))
			file.Lines = lines
		}

		for lineNum, blockID := range lineMapping {
			ann := lines[lineNum]
			if ann == nil {
				ann = &model.LineAnnotation{
					BlockID: blockID,
				}
				lines[lineNum] = ann
			} else {
				ann.BlockID = blockID
			}
		}
	}
	project.Set(model.AnnotationType_LineBlockID)
}
