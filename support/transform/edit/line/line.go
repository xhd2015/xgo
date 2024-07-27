package line

import (
	"fmt"
	"sort"

	"github.com/xhd2015/xgo/support/transform/patch/format"
)

type Edit struct {
	edits []*edit
}

type edit struct {
	id      string
	line    int
	prepend []string
	append  []string
	replace []string
}

func (c *Edit) Prepend(lineNum int, id string, lines []string) {
	c.edits = append(c.edits, &edit{
		id:      id,
		line:    lineNum,
		prepend: lines,
	})
}

func (c *Edit) Append(lineNum int, id string, lines []string) {
	c.edits = append(c.edits, &edit{
		id:     id,
		line:   lineNum,
		append: lines,
	})
}

func (c *Edit) Replace(lineNum int, id string, lines []string) {
	c.edits = append(c.edits, &edit{
		id:      id,
		line:    lineNum,
		replace: lines,
	})
}

func (c *Edit) Apply(lines []string) ([]string, error) {
	// sort edits to apply sequentially
	sort.Slice(c.edits, func(i, j int) bool {
		d := c.edits[i].line - c.edits[j].line
		if d == 0 {
			return i < j
		}
		return d < 0
	})
	n := len(lines)
	newLines := make([]string, 0, n)

	cursor := 1

	m := len(c.edits)
	for i := 0; i < m; i++ {
		e := c.edits[i]
		line := e.line
		if line > n {
			return nil, fmt.Errorf("bad line: %d", line)
		}
		if cursor < line {
			newLines = append(newLines, lines[cursor-1:line-1]...)
			cursor = line
		}
		// cursor == line
		j := i + 1
		for ; j < m && c.edits[j].line == line; j++ {
		}
		// i -> j are same line
		for k := i; k < j; k++ {
			newLines = appendLines(newLines, c.edits[k].id, c.edits[k].prepend)
		}
		var hasReplaced bool
		var hasReplacedID string
		for k := i; k < j; k++ {
			if len(c.edits[k].replace) > 0 {
				newLines = appendLines(newLines, c.edits[k].id, c.edits[k].replace)
				hasReplaced = true
				hasReplacedID = c.edits[k].id
			}
		}
		if !hasReplaced {
			newLines = append(newLines, lines[cursor-1])
		} else if hasReplacedID != "" {
			newLines = appendLines(newLines, hasReplacedID, []string{
				format.REPLACED_BEGIN,
				lines[cursor-1],
				format.REPLACED_END,
			})
		}
		cursor++
		for k := i; k < j; k++ {
			newLines = appendLines(newLines, c.edits[k].id, c.edits[k].append)
		}
		i = j - 1
	}
	newLines = append(newLines, lines[cursor-1:]...)
	return newLines, nil
}

func appendLines(orig []string, id string, lines []string) []string {
	if len(lines) == 0 {
		return orig
	}
	if id == "" {
		return append(orig, lines...)
	}
	orig = append(orig, format.Begin(id))
	orig = append(orig, lines...)
	orig = append(orig, format.End(id))
	return orig
}
