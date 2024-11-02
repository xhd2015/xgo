package git

import (
	"fmt"

	"github.com/xhd2015/xgo/support/cmd"
)

// check if head is ref's ancestor,  can be found following head's first parent
// just like git merge-base --is-ancestor head ref, but excluding merge points
//
// see help for 'git rev-list': https://git-scm.com/docs/git-rev-list
// algorithm:
//
//	find first parent set of head, and exclude first parent set of ref^1
//	    S=git rev-list --first-parent head ^ref^1
//	if ref in S, then head is a direct parent of ref
//
// return a list of commits from `from` to `to`, excluding `from` and `to`
func FindFirstParentCommitsInBetween(dir string, from string, to string) (bool, []string, error) {
	if dir == "" {
		return false, nil, fmt.Errorf("requires dir")
	}
	fromCommit, err := revParseVerified(dir, from)
	if err != nil {
		return false, nil, err
	}

	toCommit, err := RevParseVerified(dir, to)
	if err != nil {
		return false, nil, err
	}

	if fromCommit == toCommit {
		return true, nil, nil
	}

	fromCommitParent, err := RevParseVerified(dir, fromCommit+"^1")
	if err != nil {
		return false, nil, err
	}

	// --author-date-order, --reverse: from old to new
	output, err := cmd.Dir(dir).Output("git", "rev-list", "--first-parent", "--author-date-order", "--reverse", "^"+fromCommitParent, toCommit)
	if err != nil {
		return false, nil, err
	}
	commits := splitLinesFilterEmpty(output)
	if len(commits) == 0 {
		return false, nil, nil
	}
	if commits[0] != fromCommit {
		return false, nil, nil
	}
	if commits[len(commits)-1] != toCommit {
		return false, nil, nil
	}
	if len(commits) == 1 {
		return true, nil, nil
	}
	return true, commits[1 : len(commits)-1], nil
}
