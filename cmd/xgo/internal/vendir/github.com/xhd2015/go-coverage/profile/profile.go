package profile

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

type Mode string

const (
	ModeSet   Mode = "set"
	ModeCount Mode = "count"
)

type Profile struct {
	Mode   Mode
	Blocks []*CoverageBlock // don't sort blocks
}

// mode: <mode>
var modeRegex = regexp.MustCompile(`^mode: ([^\s]+)`)

func ParseProfileFile(file string) (*Profile, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ParseProfileReader(f)
}

func ParseProfile(content string) (*Profile, error) {
	return ParseProfileReader(strings.NewReader(content))
}

func ParseProfileReader(reader io.Reader) (*Profile, error) {
	scanner := bufio.NewScanner(reader)
	if !scanner.Scan() {
		return nil, fmt.Errorf("no data")
	}
	var mode Mode
	var blocks []*CoverageBlock

	line0 := scanner.Text()
	m := modeRegex.FindStringSubmatch(line0)
	if m == nil {
		return nil, fmt.Errorf("invalid mode line: %v", line0)
	}
	mode = Mode(m[1])

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			// ignore empty
			continue
		}
		block, err := ParseBlock(line)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, block)
	}
	return &Profile{
		Mode:   mode,
		Blocks: blocks,
	}, nil
}

func (c *Profile) Clone() *Profile {
	blocks := make([]*CoverageBlock, 0, len(c.Blocks))
	for _, block := range c.Blocks {
		cb := *block
		blocks = append(blocks, &cb)
	}
	return &Profile{
		Mode:   c.Mode,
		Blocks: blocks,
	}
}

func (c *Profile) String() string {
	var buf strings.Builder
	c.Format(&buf, false)
	return buf.String()
}

func (c *Profile) Format(w io.Writer, normalize bool) error {
	_, err := fmt.Fprintf(w, "mode: %s\n", c.Mode)
	if err != nil {
		return err
	}
	for _, block := range c.Blocks {
		count := block.Count
		if normalize && c.Mode == ModeSet && count > 1 {
			count = 1
		}
		_, err := fmt.Fprintln(w, block.FormatWithCount(count))
		if err != nil {
			return err
		}
	}
	return nil
}
func (c *Profile) Write(file string) error {
	return c.write(file, false)
}
func (c *Profile) WriteNormalized(file string) error {
	return c.write(file, true)
}
func (c *Profile) write(file string, normalize bool) error {
	f, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		return err
	}
	defer f.Close()
	return c.Format(f, normalize)
}

func (c *Profile) Counters() map[string][]int {
	counters := make(map[string][]int, len(c.Blocks))
	for _, block := range c.Blocks {
		counters[block.FileName] = append(counters[block.FileName], block.Count)
	}
	return counters
}

func (c *Profile) ResetCounters(counters map[string][]int) {
	idxMap := make(map[string]int, len(counters))
	for _, block := range c.Blocks {
		idx := idxMap[block.FileName]
		block.Count = counters[block.FileName][idx]
		idxMap[block.FileName] = idx + 1
	}

	if len(counters) != len(idxMap) {
		panic(fmt.Errorf("invalid counters,expect:%d files, given:%d files", len(c.Blocks), len(counters)))
	}

	// validate
	for file, idx := range idxMap {
		c, ok := counters[file]
		if !ok {
			panic(fmt.Errorf("file missing:%v", file))
		}

		if len(c) != idx {
			panic(fmt.Errorf("invalid counters %s,expect:%d , given:%d files", file, idx, len(c)))
		}
	}
}

func (c *Profile) Normalize() {
	// change to 1
	if c.Mode == ModeSet {
		for _, block := range c.Blocks {
			if block.Count > 1 {
				block.Count = 1
			}
		}
	}
}
