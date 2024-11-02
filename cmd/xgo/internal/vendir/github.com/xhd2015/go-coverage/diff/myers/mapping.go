package myers

import (
	"fmt"

	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/go-coverage/diff/vscode"
	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/go-coverage/diff/vscode/goja"
)

// useGojaDiff was used to control an experimental debugging process
// there was an issue that massive memory is allocated by myers diff,
// so we use GojaDiff instead
// but the actual problem is that there were too many too many files
// to compare that caused 12G memory being used, adding a filter
// solved that problem.
// so we laterly switched to the other diff algorithm, and this flag
// is hereby unuseful.
var useGojaDiff bool = true

func UseGojaDiff() {
	useGojaDiff = true
}

// the myers diff is based on line mapping
// but if you do not care about line-wise change, only block change,
// you can group all blocks together to make them act like lines

// ComputeBlockMapping
// a block is represented in a string
// the function computes
// the result is a mapping from each index in new to its
// counterpart in old, -1 if new.
// the result is 0-based, which is a historical design.
// we may optimize to 1-based in the future.xs
// NOTE: `newToOld` is 0-based, `deletedLines` is 1-based
func ComputeBlockMapping(oldBlocks []string, newBlocks []string) (newToOld map[int]int, deletedLines map[int]bool) {
	if useGojaDiff {
		newToOld, deletedLines = ComputeBlockMappingUsingVscodeDiff(oldBlocks, newBlocks)
		return
	}
	if true {
		panic(fmt.Errorf("ComputeBlockMapping only supports vscode"))
	}
	newToOld = make(map[int]int, len(newBlocks))
	operationsComplex(oldBlocks, newBlocks, func(oldLine, newLine int) {
		newToOld[newLine] = oldLine
	}, nil)
	return
}

func ComputeBlockMappingV2(oldBlocks []string, newBlocks []string) (newToOld map[int]int) {
	newToOld = make(map[int]int, len(newBlocks))
	// newToOldUpdate = make(map[int]int,len(newBlocks))
	operationsComplex(oldBlocks, newBlocks, func(oldLine, newLine int) { // on same
		newToOld[newLine] = oldLine
		// fmt.Printf("M: %d %d\n", oldLine+1, newLine+1)
	}, func(oldLineStart, oldLineEnd, newLineStart, newLineEnd int) { // on update
		// newToOldUpdate[]
		// fmt.Printf("U: %d %d; %d %d\n", oldLineStart+1, oldLineEnd+1, newLineStart+1, newLineEnd+1)
	})
	return
}

func ComputeBlockMappingUsingVscodeDiff(oldBlocks []string, newBlocks []string) (newToOld map[int]int, deletedLines map[int]bool) {
	changes, err := DiffChanges(oldBlocks, newBlocks)
	if err != nil {
		panic(fmt.Errorf("compute block error:%v", err))
	}
	newToOld = make(map[int]int, len(newBlocks))
	deletedLines = make(map[int]bool)
	vscode.ForeachLineMapping(changes, len(oldBlocks), len(newBlocks), func(oldLineStart, oldLineEnd, newLineStart, newLineEnd int, changeType vscode.ChangeType) {
		if changeType == vscode.ChangeTypeUnchange {
			for i, j := oldLineStart, newLineStart; i < oldLineEnd; i++ {
				// NOTE: the mapping is from new line to old line
				newToOld[j-1] = i - 1
				j++
			}
		} else if changeType == vscode.ChangeTypeDelete {
			for i := oldLineStart; i < oldLineEnd; i++ {
				// NOTE: deleted line is 1-based
				deletedLines[i] = true
			}
		}
	})
	return
}

func DiffChanges(oldBlocks []string, newBlocks []string) (changes []*vscode.LineChange, err error) {
	res, err := goja.Diff(&vscode.Request{
		OldLines: oldBlocks,
		NewLines: newBlocks,
	})
	if err != nil {
		return nil, fmt.Errorf("compute block error:%v", err)
	}
	return res.Changes, nil
}

// TODO: currently not used, maybe the only important thing is finding sames, not updates or deletions.
// deleted of old
// func ComputeMapping(oldBlocks []string, newBlocks []string) (sameNewToOld map[int]int, updatedNewToOld map[int]int, deletedOld []int) {
// 	sameNewToOld = make(map[int]int, len(newBlocks))
// 	operations := operationsComplex(oldBlocks, newBlocks, func(oldLine, newLine int) {
// 		sameNewToOld[newLine] = oldLine
// 	}, func(oldLineStart, oldLineEnd, newLineStart, newLineEnd int) {

// 	})
// 	for _, op := range operations {
// 		switch op.Kind {
// 		case Delete:
// 			for i := op.I1; i <= op.I2; i++ {
// 				deletedOld = append(deletedOld, i)
// 			}
// 		case Insert:

// 		}
// 	}
// 	return
// }
