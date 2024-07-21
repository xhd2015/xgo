package line

import (
	"fmt"
	"sort"
)

type Edit struct {
	edits []*edit
}

type edit struct {
	line    int
	prepend []string
	append  []string
	replace []string
}

func (c *Edit) Prepend(lineNum int, lines []string) {
	c.edits = append(c.edits, &edit{
		line:    lineNum,
		prepend: lines,
	})
}

func (c *Edit) Append(lineNum int, lines []string) {
	c.edits = append(c.edits, &edit{
		line:   lineNum,
		append: lines,
	})
}

func (c *Edit) Replace(lineNum int, lines []string) {
	c.edits = append(c.edits, &edit{
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
			newLines = append(newLines, c.edits[k].prepend...)
		}
		var hasReplaced bool
		for k := i; k < j; k++ {
			if len(c.edits[k].replace) > 0 {
				newLines = append(newLines, c.edits[k].replace...)
				hasReplaced = true
			}
		}
		if !hasReplaced {
			newLines = append(newLines, lines[cursor-1])
		}
		cursor++
		for k := i; k < j; k++ {
			newLines = append(newLines, c.edits[k].append...)
		}
		i = j - 1
	}
	newLines = append(newLines, lines[cursor-1:]...)
	return newLines, nil
}
