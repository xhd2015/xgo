package model

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strconv"
)

type BlockID string

// Block can be used as a unique key across one file
type Block struct {
	StartLine int `json:"startLine"`
	StartCol  int `json:"startCol"`
	EndLine   int `json:"endLine"`
	EndCol    int `json:"endCol"`
}

type IBlockData interface {
	GetBlock() *Block
}

func (c *Block) SameBlock(b *Block) bool {
	return c.StartLine == b.StartLine && c.StartCol == b.StartCol && c.EndLine == b.EndLine && c.EndCol == b.EndCol
}

func (c *Block) ID() BlockID {
	id := fmt.Sprintf("%d:%d-%d:%d", c.StartLine, c.StartCol, c.EndLine, c.EndCol)
	return BlockID(id)
}

func (c *Block) String() string {
	return string(c.ID())
}
func (c *Block) After(b *Block) bool {
	return c.Compare(b) > 0
}
func (c *Block) Before(b *Block) bool {
	return c.Compare(b) < 0
}

func (c *Block) Clone() *Block {
	if c == nil {
		return nil
	}
	return &Block{
		StartLine: c.StartLine,
		StartCol:  c.StartCol,
		EndLine:   c.EndLine,
		EndCol:    c.EndCol,
	}
}
func (a *Block) Compare(b *Block) int {
	lineD := a.StartLine - b.StartLine
	if lineD != 0 {
		return lineD
	}
	colD := a.StartCol - b.StartCol
	if colD != 0 {
		return colD
	}
	endLineD := a.EndLine - b.EndLine
	if endLineD != 0 {
		return endLineD
	}
	return a.EndCol - b.EndCol
}

func (c *BlockID) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		*c = ""
		return nil
	}
	if data[0] == '"' {
		var s string
		err := json.Unmarshal(data, &s)
		if err != nil {
			return err
		}
		*c = BlockID(s)
		return nil
	}
	// try as number
	_, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return json.Unmarshal(data, &c)
	}
	s := string(data)
	*c = BlockID(s)
	return nil
}

// blocks should be a slice of IBlockData
func SortIBlocks(blocks interface{}) {
	v := reflect.ValueOf(blocks)
	if v.Kind() != reflect.Slice {
		panic(fmt.Errorf("SortIBlocks requires slice, found: %T", blocks))
	}
	getBlock := func(i int) *Block {
		return v.Index(i).Interface().(IBlockData).GetBlock()
	}
	doSortBlocks(blocks, getBlock)
}

func doSortBlocks(blocks interface{}, getBlock func(i int) *Block) {
	sort.Slice(blocks, func(i, j int) bool {
		return getBlock(i).Compare(getBlock(j)) < 0
	})
}
