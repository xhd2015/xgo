package vscode

import "fmt"

type ChangeType int

const (
	ChangeTypeNone     ChangeType = 0
	ChangeTypeUnchange ChangeType = 1
	ChangeTypeUpdate   ChangeType = 2
	ChangeTypeInsert   ChangeType = 3
	ChangeTypeDelete   ChangeType = 4
)

// compute unchanged line mapping
func ComputeBlockMapping(oldBlocks []string, newBlocks []string) (map[int]int, error) {
	res, err := Diff(&Request{
		OldLines: oldBlocks,
		NewLines: newBlocks,
	})
	if err != nil {
		return nil, err
	}
	return ParseUnchangedLines(res.Changes, len(oldBlocks), len(newBlocks))
}

func ParseUnchangedLines(changes []*LineChange, oldLines int, newLines int) (map[int]int, error) {
	mapping := make(map[int]int)
	err := ForeachLineMapping(changes, oldLines, newLines, func(oldLineStart, oldLineEnd, newLineStart, newLineEnd int, changeType ChangeType) {
		if changeType == ChangeTypeUnchange {
			m := newLineEnd - newLineStart
			for k := 0; k < m; k++ {
				// our mapping is 0-based
				mapping[newLineStart+k-1] = oldLineStart + k - 1
			}
		}
	})
	if err != nil {
		return nil, err
	}

	return mapping, nil
}

// ForeachLineMapping iterate each line range with associated markup
// NOTE: when fn gets called, the lines are 1-based, and start is inclusive, end is exclusive.
func ForeachLineMapping(changes []*LineChange, oldLines int, newLines int, fn func(oldLineStart int, oldLineEnd int, newLineStart int, newLineEnd int, changeType ChangeType)) error {
	// convert changes to line mapping
	oldLineStart := 1
	newLineStart := 1
	n := len(changes)
	for i := 0; i <= n; i++ {
		var oldLineEnd int
		var newLineEnd int
		var oldLineStartNext int            // to update oldLineStart
		var newLineStartNext int            // to update newLineStart
		var oldLineStartNextForCallback int // to call fn
		var newLineStartNextForCallback int // to call fn
		changeType := ChangeTypeNone
		if i < n {
			change := changes[i]

			changeType = ChangeTypeUpdate

			oldLineEnd = change.OriginalStartLineNumber
			oldLineStartNext = change.OriginalEndLineNumber + 1
			oldLineStartNextForCallback = oldLineStartNext
			// lines before change are unmodified
			if change.OriginalEndLineNumber == 0 {
				// insertion of new text, align at next line
				oldLineEnd++
				oldLineStartNext = oldLineEnd
				oldLineStartNextForCallback = oldLineEnd
				changeType = ChangeTypeInsert
			}

			newLineEnd = change.ModifiedStartLineNumber
			newLineStartNext = change.ModifiedEndLineNumber + 1
			newLineStartNextForCallback = newLineStartNext
			if change.ModifiedEndLineNumber == 0 {
				// deletion of old text
				newLineEnd++
				newLineStartNext = newLineEnd
				newLineStartNextForCallback = newLineEnd
				changeType = ChangeTypeDelete
			}
		} else {
			// lines are 1-based
			oldLineEnd = oldLines + 1
			newLineEnd = newLines + 1
		}
		// unchanged mapping must have exactluy same count of lines
		m := newLineEnd - newLineStart
		if m != oldLineEnd-oldLineStart {
			return fmt.Errorf("unchanged range not the same: new=[%v->%v](%d), old=[%v->%v](%d)", newLineStart, newLineEnd, m, oldLineStart, oldLineEnd, oldLineEnd-oldLineStart)
		}
		// unchanged lines
		fn(oldLineStart, oldLineEnd, newLineStart, newLineEnd, ChangeTypeUnchange)
		if changeType != ChangeTypeNone {
			fn(oldLineEnd, oldLineStartNextForCallback, newLineEnd, newLineStartNextForCallback, changeType)
		}

		oldLineStart = oldLineStartNext
		newLineStart = newLineStartNext
	}

	return nil
}
