package filter

import (
	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/model/filter"
	path_filter "github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/path/filter"
)

type checker struct {
	opts       *filter.Options
	pathFilter *path_filter.FileFilter
}

func NewChecker(opts *filter.Options) *checker {
	if opts == nil {
		return nil
	}
	noFileFilter := len(opts.Include) == 0 && len(opts.Exclude) == 0
	if len(opts.Suffix) == 0 && len(opts.ExcludeSuffix) == 0 && noFileFilter {
		return nil
	}
	var pathFilter *path_filter.FileFilter
	if !noFileFilter {
		pathFilter = path_filter.NewFileFilter(opts.Include, opts.Exclude)
	}

	return &checker{
		opts:       opts,
		pathFilter: pathFilter,
	}
}

func (c *checker) MatchFile(file string) bool {
	if c == nil {
		return true
	}
	if !c.opts.MatchSuffix(file) {
		return false
	}
	if c.pathFilter != nil && !c.pathFilter.MatchFile(file) {
		return false
	}
	return true
}
