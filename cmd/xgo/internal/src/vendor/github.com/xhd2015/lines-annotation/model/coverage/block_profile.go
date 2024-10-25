package coverage

import "github.com/xhd2015/lines-annotation/model"

// BlockProfile a general way to organize block
type BlockProfile map[PkgFile][]*BlockData

func (c BlockProfile) ForeachBlock(fn func(pkgFile string, data *BlockData) bool) {
	for pkgFile, blocks := range c {
		for _, data := range blocks {
			if !fn(pkgFile, data) {
				return
			}
		}
	}
}
func (c BlockProfile) Append(pkgFile string, data *BlockData) {
	c[pkgFile] = append(c[pkgFile], data)
}

func (c BlockProfile) SortAll() {
	for _, blocks := range c {
		model.SortIBlocks(blocks)
	}
}

func (c BlockProfile) MakeBlockMapping() map[PkgFile]map[model.Block]*BlockData {
	m := make(map[PkgFile]map[model.Block]*BlockData, len(c))
	for pkgFile, blocks := range c {
		bm := make(map[model.Block]*BlockData, len(blocks))
		for _, data := range blocks {
			bm[*data.Block] = data
		}

		m[pkgFile] = bm
	}
	return m
}

// LeftJoin perform join operation on two profiles, identifying blocks are based on a.
// joint: if a,b not nil, they are joined together.
func (c BlockProfile) LeftJoin(b BlockProfile, joint func(a interface{}, b interface{}) (r interface{}, ok bool)) BlockProfile {
	res := make(BlockProfile)
	bm := b.MakeBlockMapping()
	for pkgFile, blocks := range c {
		jm := make([]*BlockData, 0, len(blocks))
		for _, x := range blocks {
			y := bm[pkgFile][*x.Block]
			if y == nil {
				continue
			}
			r, ok := joint(x.Data, y.Data)
			if !ok {
				continue
			}
			jm = append(jm, &BlockData{
				Block: x.Block,
				Data:  r,
			})
		}
		res[pkgFile] = jm
	}
	return res
}
