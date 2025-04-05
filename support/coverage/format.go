package coverage

import (
	"strconv"
	"strings"
)

// the coverage format is defined at:
//
//	  https://github.com/golang/go/blob/master/src/cmd/cover/profile.go
//	>	First line is "mode: set|count|atomic".
//	>	Rest of file is in the format
//	>	  encoding/base64/base64.go:34.44,37.40 3 1
//	>	where the fields are: pkg/name.go:startLine.column,endLine.column numberOfStatements count
func Format(mode string, lines []*CovLine) string {
	strs := make([]string, 0, len(lines)+1)
	strs = append(strs, "mode: "+mode)
	for _, line := range lines {
		strs = append(strs, line.Prefix+" "+strconv.FormatInt(line.Count, 10))
	}
	return strings.Join(strs, "\n")
}
