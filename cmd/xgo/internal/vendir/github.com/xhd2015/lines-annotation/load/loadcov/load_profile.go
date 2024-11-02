package loadcov

import (
	_ "embed"
	"errors"
	"fmt"
	"strings"

	"os"
	"regexp"
	"strconv"

	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/load"
	ann_model "github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/model"
	coverage_model "github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/model/coverage"
	"github.com/xhd2015/xgo/support/coverage"
)

func LoadCoverageProfileFiles(modPath string, files []string, excludePrefix []string) (*ann_model.ProjectAnnotation, error) {
	return loadProfiles(modPath, files, excludePrefix, false)
}

func LoadOptionalCoverageProfileFiles(modPath string, files []string, excludePrefix []string) (*ann_model.ProjectAnnotation, error) {
	return loadProfiles(modPath, files, excludePrefix, true)
}

func loadProfiles(modPath string, files []string, excludePrefix []string, allowMissing bool) (*ann_model.ProjectAnnotation, error) {
	res, err := parseProfiles(files, excludePrefix, allowMissing)
	if err != nil {
		return nil, err
	}
	profile, err := ConvertToBinaryProfile(res)
	if err != nil {
		return nil, err
	}
	return load.BinaryProfileToAnnotation(modPath, profile), nil
}

func ParseProfiles(files []string, excludePrefix []string) ([]*coverage.CovLine, error) {
	return parseProfiles(files, excludePrefix, false)
}

func parseProfiles(files []string, excludePrefix []string, allowMissing bool) ([]*coverage.CovLine, error) {
	covs := make([][]*coverage.CovLine, 0, len(files))
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			if allowMissing && os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		_, cov := coverage.Parse(string(content))
		covs = append(covs, cov)
	}
	res := coverage.Merge(covs...)
	res = coverage.Filter(res, func(line *coverage.CovLine) bool {
		return !hasAnyPrefix(line.Prefix, excludePrefix)
	})
	return res, nil
}

func hasAnyPrefix(s string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(s, prefix) {
			return true
		}
	}
	return false
}

func ConvertToBinaryProfile(res []*coverage.CovLine) (coverage_model.BinaryProfile, error) {
	profile := make(coverage_model.BinaryProfile)
	for _, covLine := range res {
		line := covLine.Prefix + " " + strconv.FormatInt(covLine.Count, 10)
		block, err := parseBlock(line)
		if err != nil {
			return nil, err
		}
		profileBlock := &coverage_model.BlockStats{
			Block: &ann_model.Block{
				StartLine: block.Start.Line,
				StartCol:  block.Start.Col,
				EndLine:   block.End.Line,
				EndCol:    block.End.Col,
			},
			Count: map[string]int64{
				"": int64(block.Count),
			},
		}
		profile[block.FileName] = append(profile[block.FileName], profileBlock)
	}
	return profile, nil
}

type CoverageBlock struct {
	FileName string // format: <pkg>/<file>, NOTE: ends with .go
	Block
	NumStmts int
	Count    int
}

// <fileName>:<line0>.<col0>,<line1>.<col1> <num_stmts> <count>
var re = regexp.MustCompile(`^([^:]+):(\d+)\.(\d+),(\d+)\.(\d+) (\d+) (\d+)`)

type Pos struct {
	Line int
	Col  int
}
type Block struct {
	Start Pos
	End   Pos
}

func parseBlock(line string) (*CoverageBlock, error) {
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
