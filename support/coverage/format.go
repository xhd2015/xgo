package coverage

import (
	"strconv"
	"strings"
)

// mode can be: set
func Format(mode string, lines []*CovLine) string {
	strs := make([]string, 0, len(lines)+1)
	strs = append(strs, "mode: "+mode)
	for _, line := range lines {
		strs = append(strs, line.Prefix+" "+strconv.FormatInt(line.Count, 10))
	}
	return strings.Join(strs, "\n")
}
