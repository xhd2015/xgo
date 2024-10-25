package merge

import (
	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/go-coverage/code"
	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/go-coverage/cover"

	// diff "github.com/xhd2015/go-coverage/diff/myers"
	diff "github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/go-coverage/diff/vscode"
)

func ComputeFileBlockMapping(oldFileName string, oldCode string, newFileName string, newCode string) (newToOldMapping map[int]int, err error) {
	oldFset, oldAst, err := code.ParseCodeString(oldFileName, oldCode)
	if err != nil {
		return nil, err
	}
	newFset, newAst, err := code.ParseCodeString(newFileName, newCode)
	if err != nil {
		return nil, err
	}
	oldBlocks := cover.CollectStmts(oldFset, oldAst, []byte(oldCode))
	newBlocks := cover.CollectStmts(newFset, newAst, []byte(newCode))
	return diff.ComputeBlockMapping(oldBlocks, newBlocks)
}
