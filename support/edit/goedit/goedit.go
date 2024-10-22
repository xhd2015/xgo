package goedit

import (
	"go/token"

	"github.com/xhd2015/xgo/support/edit"
)

type Edit struct {
	buf  *edit.Buffer
	fset *token.FileSet
}

func New(fset *token.FileSet, content string) *Edit {
	return &Edit{
		fset: fset,
		buf:  edit.NewBuffer([]byte(content)),
	}
}

func (c *Edit) Delete(start token.Pos, end token.Pos) {
	c.buf.Delete(c.offsetOf(start), c.offsetOf(end))
}

func (c *Edit) Insert(start token.Pos, content string) {
	c.buf.Insert(c.offsetOf(start), content)
}

func (c *Edit) Replace(start token.Pos, end token.Pos, content string) {
	c.buf.Replace(c.offsetOf(start), c.offsetOf(end), content)
}

func (c *Edit) String() string {
	return c.buf.String()
}

func (c *Edit) offsetOf(pos token.Pos) int {
	if pos == token.NoPos {
		return -1
	}
	return c.fset.Position(pos).Offset
}
