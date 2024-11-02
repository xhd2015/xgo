package git

import (
	"fmt"
	"strings"

	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/gitops/model"
)

type ListBranchRefOptions struct {
	Count int64
}

// $ git for-each-ref --format='%(objectname) %(refname:lstrip=3)' refs/remotes/origin
// ba3e2ab996a202c99f814c58384bcd65792c9015 1.11.0-cr-xf
// d45c5e4ddfb708ab66e90da1a1f95592c222647d 1.14.0-csc-mr-use
// lstrip=3: remove 3 prefix path, that is ["refs","remotes","origin"]
func ListBranchRef(dir string, opts *ListBranchRefOptions) ([]*model.BranchRef, error) {
	var countOpts string
	if opts != nil {
		if opts.Count > 0 {
			countOpts = fmt.Sprintf("--count=%d", opts.Count)
		}
	}
	res, err := RunCommand(dir, func(commands []string) []string {
		return append(commands, []string{
			fmt.Sprintf("git for-each-ref --sort=-committerdate %s --format='%%(objectname) %%(refname:lstrip=3)' refs/remotes/origin", countOpts),
		}...)
	})
	if err != nil {
		return nil, err
	}

	lines := splitLinesFilterEmpty(res)
	branches := make([]*model.BranchRef, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		splits := strings.SplitN(line, " ", 2)
		if len(splits) != 2 {
			continue
		}
		commitHash := strings.TrimSpace(splits[0])
		branch := strings.TrimSpace(splits[1])

		if commitHash == "" || branch == "" || branch == "HEAD" {
			// exclude HEAD
			continue
		}

		branches = append(branches, &model.BranchRef{
			Branch:     branch,
			CommitHash: commitHash,
		})
	}

	return branches, nil
}
