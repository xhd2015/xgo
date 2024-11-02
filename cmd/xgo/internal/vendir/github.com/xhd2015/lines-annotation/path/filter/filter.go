package filter

import (
	"github.com/xhd2015/xgo/support/pattern"
)

// FileFilter
type FileFilter struct {
	Include []string `json:"include"`
	Exclude []string `json:"exclude"`

	includePatterns pattern.Patterns
	excludePatterns pattern.Patterns
}

func NewFileFilter(include []string, exclude []string) *FileFilter {
	return &FileFilter{
		Include:         include,
		Exclude:         exclude,
		includePatterns: CompilePatterns(include),
		excludePatterns: CompilePatterns(exclude),
	}
}

func CompilePatterns(patterns []string) pattern.Patterns {
	list := make([]*pattern.Pattern, 0, len(patterns))
	for _, p := range patterns {
		ptn := pattern.CompilePattern(p)
		list = append(list, ptn)
	}
	return list
}

// MatchFile checks whether patterns of this filter
// match given *file*.
// NOTE: the target `file` must be a file, not a
// directory
func (c *FileFilter) MatchFile(file string) bool {
	hasInclude := len(c.Include) > 0
	if hasInclude {
		if !c.includePatterns.MatchAnyPrefix(file) {
			return false
		}
	}
	hasExclude := len(c.Exclude) > 0
	if hasExclude {
		if c.excludePatterns.MatchAnyPrefix(file) {
			return false
		}
	}
	return true
}
