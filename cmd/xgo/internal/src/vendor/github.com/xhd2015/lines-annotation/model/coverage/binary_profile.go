package coverage

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"

	"github.com/xhd2015/lines-annotation/model"
)

type ShortFilename = string // xxx.go
type PkgPath = string
type PkgFile = string // <PkgPath>/<ShortFilename>

// BinaryProfile represents results generated dynamically
// it is named binary to distinguish with the original
// string-formatted profile.
type BinaryProfile map[PkgFile][]*BlockStats

func (c BinaryProfile) Labels() []string {
	labelMap := make(map[string]bool)
	c.ForeachBlock(func(pkgFile string, data *BlockStats) bool {
		for label := range data.Count {
			labelMap[label] = true
		}
		return true
	})
	labels := make([]string, 0, len(labelMap))
	for label := range labelMap {
		labels = append(labels, label)
	}
	return labels
}

func (c BinaryProfile) Clone() BinaryProfile {
	b := make(BinaryProfile, len(c))
	for pkgFile, counters := range c {
		newCounters := make([]*BlockStats, 0, len(counters))
		for _, counter := range counters {
			newCounters = append(newCounters, counter.Clone())
		}
		b[pkgFile] = newCounters
	}
	return b
}

// before calling MergeSameLoad, you should call StaticChecksum to compare
// if two profiles have have the same static structure.
func (c BinaryProfile) MergeSameLoad(b BinaryProfile) error {
	for pkgFile, blocksA := range c {
		blocksB := b[pkgFile]
		if len(blocksB) != len(blocksA) {
			return fmt.Errorf("inconsistent blocks at file: %s, want %d blocks, actual %d", pkgFile, len(blocksA), len(blocksB))
		}

		for i, blockA := range blocksA {
			blockB := blocksB[i]
			if blockA.Count == nil {
				blockA.Count = make(map[string]int64, len(blockB.Count))
			}
			for label, count := range blockB.Count {
				blockA.Count[label] += count
			}
		}
	}
	return nil
}

// StaticChecksum return's the profile's block info checksum, discarding
// dynamic counters. It can be used to check if two profile have the same
// static checksum to safely perform merge operation.
func (c BinaryProfile) StaticChecksum() string {
	pairs := c.sortedPairs()

	h := md5.New()
	h.Write([]byte(strconv.Itoa(len(pairs))))
	for _, pair := range pairs {
		h.Write([]byte(pair.PkgFile))
		h.Write([]byte(strconv.Itoa(len(pair.Blocks))))
		for _, blockData := range pair.Blocks {
			block := blockData.Block
			h.Write([]byte(strconv.Itoa(block.StartLine)))
			h.Write([]byte(strconv.Itoa(block.StartCol)))
			h.Write([]byte(strconv.Itoa(block.EndLine)))
			h.Write([]byte(strconv.Itoa(block.EndCol)))
		}
	}
	return hex.EncodeToString(h.Sum(nil))
}

// StaticChecksum_Col8bits this is a temporary util method, for bugfix.
func (c BinaryProfile) StaticChecksum_Col8bits() string {
	pairs := c.sortedPairs()

	h := md5.New()
	h.Write([]byte(strconv.Itoa(len(pairs))))
	for _, pair := range pairs {
		h.Write([]byte(pair.PkgFile))
		h.Write([]byte(strconv.Itoa(len(pair.Blocks))))
		for _, blockData := range pair.Blocks {
			block := blockData.Block
			h.Write([]byte(strconv.Itoa(block.StartLine)))
			h.Write([]byte(strconv.Itoa(block.StartCol & 0xff)))
			h.Write([]byte(strconv.Itoa(block.EndLine)))
			h.Write([]byte(strconv.Itoa(block.EndCol)))
		}
	}
	return hex.EncodeToString(h.Sum(nil))
}

func (c BinaryProfile) Checksum() string {
	pairs := c.sortedPairs()

	h := md5.New()
	h.Write([]byte(strconv.Itoa(len(pairs))))
	for _, pair := range pairs {
		h.Write([]byte(pair.PkgFile))
		h.Write([]byte(strconv.Itoa(len(pair.Blocks))))
		for _, blockData := range pair.Blocks {
			block := blockData.Block
			h.Write([]byte(strconv.Itoa(block.StartLine)))
			h.Write([]byte(strconv.Itoa(block.StartCol)))
			h.Write([]byte(strconv.Itoa(block.EndLine)))
			h.Write([]byte(strconv.Itoa(block.EndCol)))

			// sum label count
			type LabelCount struct {
				Label string
				Count int64
			}
			countMapping := blockData.Count
			counts := make([]LabelCount, 0, len(countMapping))
			for label, count := range countMapping {
				counts = append(counts, LabelCount{
					Label: label,
					Count: count,
				})
			}
			sort.Slice(counts, func(i, j int) bool {
				return counts[i].Label < counts[j].Label
			})
			for _, c := range counts {
				h.Write([]byte(c.Label))
				h.Write([]byte(strconv.Itoa(int(c.Count))))
			}

		}
	}
	return hex.EncodeToString(h.Sum(nil))
}

// StaticFileChecksum return's the profile's block info checksum, grouping by file
func (c BinaryProfile) StaticFileChecksum() map[string]string {
	m := make(map[string]string, len(c))
	for pkgFile, blockData := range c {
		h := md5.New()
		blocks := BlockStatsSlice(blockData).SortCopy()
		for _, block := range blocks {
			h.Write([]byte(strconv.Itoa(block.StartLine)))
			h.Write([]byte(strconv.Itoa(block.StartCol)))
			h.Write([]byte(strconv.Itoa(block.EndLine)))
			h.Write([]byte(strconv.Itoa(block.EndCol)))
		}
		m[pkgFile] = hex.EncodeToString(h.Sum(nil))
	}
	return m
}

type fileBlockPair struct {
	PkgFile string
	Blocks  []*BlockStats
}

func (c BinaryProfile) sortedPairs() []*fileBlockPair {
	// sort pairs and blocks
	pairs := make([]*fileBlockPair, 0, len(c))
	for pkgFile, blocks := range c {
		pairs = append(pairs, &fileBlockPair{PkgFile: pkgFile, Blocks: BlockStatsSlice(blocks).SortCopy()})
	}
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].PkgFile < pairs[j].PkgFile
	})
	return pairs
}

func (c BinaryProfile) ToBlockProfile() BlockProfile {
	b := make(BlockProfile, len(c))
	for pkgFile, blocks := range c {
		bm := make([]*BlockData, 0, len(blocks))
		for _, data := range blocks {
			bm = append(bm, &BlockData{
				Block: data.Block,
				Data:  data.Count,
			})
		}
		b[pkgFile] = bm
	}
	return b
}

func (c BinaryProfile) ForeachBlock(fn func(pkgFile string, data *BlockStats) bool) {
	for pkgFile, blocks := range c {
		for _, data := range blocks {
			if !fn(pkgFile, data) {
				return
			}
		}
	}
}

func (c BinaryProfile) SortAll() {
	for _, blocks := range c {
		model.SortIBlocks(blocks)
	}
}
