package git

import (
	"fmt"

	"github.com/xhd2015/xgo/support/cmd"
)

// git tag -l --sort=-committerdate --format='%(refname)' --contains xxx
//
//	refs/tags/v2.24.0.d01
func SearchTagsContainingRef(dir string, ref string) ([]string, error) {
	if ref == "" {
		return nil, fmt.Errorf("requires ref")
	}
	// --exact-match
	output, err := cmd.Dir(dir).Output("git", "tag", "-l", "--sort=-committerdate", "--format=%(refname)", "--contains", ref)
	if err != nil {
		return nil, err
	}
	return splitLinesFilterEmpty(output), nil
}
