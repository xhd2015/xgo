package cover

import (
	"go/ast"
	"go/token"

	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/go-coverage/code"
)

// blockCollector collects non-empty stmts
type blockCollector struct {
	blocks []string
}

var _ Callback = ((*blockCollector)(nil))

// OnBlock implements Callback
func (c *blockCollector) OnBlock(insertPos token.Pos, pos token.Pos, end token.Pos, numStmts int, basicStmts []ast.Stmt) {
	stmtsCode := code.CleanList(func(f func(n ast.Node, last bool)) {
		for i, s := range basicStmts {
			f(s, i >= len(basicStmts)-1)
		}
	}, ";", code.CleanOpts{
		ShouldFormat: func(n ast.Node) bool {
			return n.Pos() < end
		},
	})
	if stmtsCode == "" {
		return
	}
	c.blocks = append(c.blocks, stmtsCode)
}

// OnWrapElse implements Callback
func (c *blockCollector) OnWrapElse(lbrace int, rbrace int) {
	// do nothing
}

func CollectStmts(fset *token.FileSet, f *ast.File, content []byte) []string {
	bc := &blockCollector{}
	fvisitor := &File{
		fset:     fset,
		content:  content,
		callback: bc,
	}
	ast.Walk(fvisitor, f)
	return bc.blocks
}
