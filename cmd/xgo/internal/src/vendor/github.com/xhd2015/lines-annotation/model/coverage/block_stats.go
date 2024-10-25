package coverage

import (
	"github.com/xhd2015/go-coverage/merge"
	"github.com/xhd2015/lines-annotation/model"
)

type BlockStats struct {
	*model.Block

	// Count times execution reached
	Count map[string]int64 `json:"count"` // including empty label
	// 	Stmts int // by agent
}

type BlockStatsSlice []*BlockStats

func (c *BlockStats) GetBlock() *model.Block {
	return c.Block
}

func (c *BlockStats) Clone() *BlockStats {
	m := make(map[string]int64, len(c.Count))
	for label, value := range c.Count {
		m[label] = value
	}
	return &BlockStats{
		Block: c.Block.Clone(),
		Count: m,
	}
}

// Add implements merge.Counter
func (c *BlockStats) Add(b merge.Counter) merge.Counter {
	return c.MergeBlockStats(b.(*BlockStats))
}

// MergeBlockStats add labels
func (c *BlockStats) MergeBlockStats(b *BlockStats) *BlockStats {
	res := c.Clone()
	for label, val := range b.Count {
		res.Count[label] += val
	}
	return res
}

func (c BlockStatsSlice) SortCopy() BlockStatsSlice {
	cpBlocks := make([]*BlockStats, 0, len(c))
	cpBlocks = append(cpBlocks, c...)
	model.SortIBlocks(cpBlocks)
	return cpBlocks
}
