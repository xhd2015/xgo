package git

import (
	"fmt"

	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/gitops/model"
	"github.com/xhd2015/xgo/support/cmd"
)

// RevListAll git rev-list --all --since=date --format=...
func RevListAll(dir string, sinceDate string) ([]*model.Commit, error) {
	if dir == "" {
		return nil, fmt.Errorf("requires dir")
	}

	// --no-commit-header: don't show header "commit xxxx" before each format. by default on
	args := []string{"rev-list", "--all", "--no-commit-header", "--format=" + makeCommitFormat(true)}
	if sinceDate != "" {
		args = append(args, "--since="+sinceDate)
	}
	res, err := cmd.Dir(dir).Output("git", args...)
	if err != nil {
		return nil, err
	}

	return parseCommits(res), nil
}
