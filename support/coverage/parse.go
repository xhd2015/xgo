package coverage

import (
	"strconv"
	"strings"
)

type CovLine struct {
	Prefix string
	Count  int64
}

const modePrefix = "mode:"

func Parse(content string) (mode string, covLines []*CovLine) {
	lines := strings.Split(content, "\n")
	if len(lines) > 0 && strings.HasPrefix(lines[0], "mode:") {
		mode = strings.TrimSpace(lines[0][len(modePrefix):])
		lines = lines[1:]
	}
	covLines = make([]*CovLine, 0, len(lines))
	for _, line := range lines {
		covLine := ParseCovLine(line)
		if covLine == nil {
			continue
		}
		covLines = append(covLines, covLine)
	}
	return mode, covLines
}

func ParseCovLine(line string) *CovLine {
	idx := strings.LastIndex(line, " ")
	if idx < 0 {
		return nil
	}
	cnt, err := strconv.ParseInt(line[idx+1:], 10, 64)
	if err != nil {
		cnt = 0
	}
	return &CovLine{
		Prefix: line[:idx],
		Count:  cnt,
	}
}

// same file can have multiple lines of count
// see: https://go-review.googlesource.com/c/go/+/76875
// go will append coverage profiles of multiple runnings
func Compact(lines []*CovLine) []*CovLine {
	return compact(make(map[string]*CovLine, len(lines)), lines)
}

func compact(lineByPrefix map[string]*CovLine, lines []*CovLine) []*CovLine {
	n := len(lines)

	i := 0
	for j := 0; j < n; j++ {
		line := lines[j]
		prefix := lines[j].Prefix
		prevLine, ok := lineByPrefix[prefix]
		if ok {
			prevLine.Count += line.Count
			continue
		}
		lineByPrefix[prefix] = line
		lines[i] = lines[j]
		i++
	}
	return lines[:i]
}
