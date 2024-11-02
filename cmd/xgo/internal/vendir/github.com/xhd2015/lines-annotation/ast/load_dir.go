package ast

import (
	"go/token"
	"io/fs"
	"strings"

	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/path/filter"
	"github.com/xhd2015/xgo/support/fileutil"
)

// LoadDir scans a given directory
// to specify extra options, one can use:
//
//	LoadDirOptions().Exclude("vendor").Load("some/go_project")
func LoadDir(dir string) (LoadInfo, error) {
	return LoadDirOptions().Load(dir)
}

type dirLoadInfo struct {
	dir   string
	fset  *token.FileSet
	files []*file
}

var _ LoadInfo = (*dirLoadInfo)(nil)

func (c *dirLoadInfo) FileSet() *token.FileSet {
	return c.fset
}

func (c *dirLoadInfo) RangeFiles(handler func(f File) bool) {
	for _, file := range c.files {
		if !handler(file) {
			return
		}
	}
}

type loadDirOptions struct {
	includes []string
	excludes []string
}

func LoadDirOptions() *loadDirOptions {
	return &loadDirOptions{}
}

func (c *loadDirOptions) Include(includes ...string) *loadDirOptions {
	c.includes = append(c.includes, includes...)
	return c
}

func (c *loadDirOptions) Exclude(excludes ...string) *loadDirOptions {
	c.excludes = append(c.excludes, excludes...)
	return c
}

func (c *loadDirOptions) Load(dir string) (LoadInfo, error) {
	opts := c
	if opts == nil {
		opts = &loadDirOptions{}
	}
	includes := opts.includes
	excludes := opts.excludes
	filter := filter.NewFileFilter(includes, excludes)

	fset := token.NewFileSet()
	var files []*file

	err := fileutil.WalkRelative(dir, func(path, relPath string, d fs.DirEntry) error {
		if !isGoFile(d) {
			return nil
		}
		// is go file, so check if file matches
		if !filter.MatchFile(relPath) {
			return nil
		}
		file, err := loadFile(fset, path, relPath)
		if err != nil {
			return err
		}
		if file != nil {
			files = append(files, file)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &dirLoadInfo{
		dir:   dir,
		fset:  fset,
		files: files,
	}, nil
}

func hasGoSuffix(s string) bool {
	return strings.HasSuffix(s, ".go")
}
func isGoFile(d fs.DirEntry) bool {
	return !d.IsDir() && hasGoSuffix(d.Name())
}
