package profile

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

type Pos struct {
	Line int
	Col  int
}

type Block struct {
	Start Pos
	End   Pos
}

type CoverageBlock struct {
	FileName string // format: <pkg>/<file>, NOTE: ends with .go
	Block
	NumStmts int
	Count    int
}

// <fileName>:<line0>.<col0>,<line1>.<col1> <num_stmts> <count>
var re = regexp.MustCompile(`^([^:]+):(\d+)\.(\d+),(\d+)\.(\d+) (\d+) (\d+)`)

func ParseBlock(line string) (*CoverageBlock, error) {
	m := re.FindStringSubmatch(line)
	if m == nil {
		return nil, errors.New("invalid line")
	}
	line0, err := strconv.ParseInt(m[2], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("line0:%v", err)
	}
	col0, err := strconv.ParseInt(m[3], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("col0:%v", err)
	}
	line1, err := strconv.ParseInt(m[4], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("line1:%v", err)
	}
	col1, err := strconv.ParseInt(m[5], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("col1:%v", err)
	}
	numStmts, err := strconv.ParseInt(m[6], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("num_stmts:%v", err)
	}
	count, err := strconv.ParseInt(m[7], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("count:%v", err)
	}
	return &CoverageBlock{
		FileName: m[1],
		Block: Block{
			Start: Pos{
				Line: int(line0),
				Col:  int(col0),
			},
			End: Pos{
				Line: int(line1),
				Col:  int(col1),
			},
		},
		NumStmts: int(numStmts),
		Count:    int(count),
	}, nil
}

func (c *Block) String() string {
	return fmt.Sprintf("%d.%d,%d.%d", c.Start.Line, c.Start.Col, c.End.Line, c.End.Col)
}

func (c *CoverageBlock) String() string {
	return c.FormatWithCount(c.Count)
}
func (c *CoverageBlock) FormatWithCount(count int) string {
	return fmt.Sprintf("%s:%s %d %d", c.FileName, c.Block.String(), c.NumStmts, count)
}

func (c *Block) Compare(b *Block) int {
	start := c.Start.Compare(&b.Start)
	if start != 0 {
		return start
	}
	return c.End.Compare(&b.End)
}

func (c *Pos) Compare(b *Pos) int {
	l := c.Line - b.Line
	if l != 0 {
		return l
	}
	return c.Col - b.Col
}
