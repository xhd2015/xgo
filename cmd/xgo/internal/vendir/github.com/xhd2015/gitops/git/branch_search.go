package git

import (
	"fmt"

	"github.com/xhd2015/xgo/support/cmd"
)

func SearchBranchesContainingRef(dir string, ref string) ([]string, error) {
	refs, err := SearchRefsContainingRef(dir, ref)
	if err != nil {
		return nil, err
	}
	return TrimRefsAsBranches(refs), nil
}

// git branch -l --all --sort=-committerdate --format='%(refname)' --contains xxx
// refs/heads/dev-master
// refs/remotes/origin/dev-master
// refs/remotes/origin/fix-my-tun
func SearchRefsContainingRef(dir string, ref string) ([]string, error) {
	if ref == "" {
		return nil, fmt.Errorf("requires ref")
	}
	// --exact-match
	output, err := cmd.Dir(dir).Output("git", "branch", "-l", "--all", "--sort=-committerdate", "--format=%(refname)", "--contains", ref)
	if err != nil {
		return nil, err
	}
	return splitLinesFilterEmpty(output), nil
}

func GetBranchesHoldingRef(dir string, ref string) ([]string, error) {
	if ref == "" {
		return nil, fmt.Errorf("requires ref")
	}

	// resolve possible refs
	refs, err := SearchRefsContainingRef(dir, ref)
	if err != nil {
		return nil, err
	}
	var possibleBranches []string
	for _, branch := range refs {
		ok, err := IsFirstParentReachable(dir, ref, branch)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		possibleBranches = append(possibleBranches, branch)
	}
	return TrimRefsAsBranches(possibleBranches), nil
}
