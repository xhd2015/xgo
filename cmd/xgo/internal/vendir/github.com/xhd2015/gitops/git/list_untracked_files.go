package git

import (
	"fmt"

	"github.com/xhd2015/xgo/support/cmd"
)

func ListUntrackedFiles(dir string, compareRef string, patterns []string) ([]string, error) {
	if compareRef == "" {
		return nil, fmt.Errorf("requires compareRef")
	}
	if compareRef == COMMIT_WORKING {
		return nil, fmt.Errorf("compareRef cannot be WORKING commit")
	}

	args := []string{"ls-files", "--no-empty-directory", "--exclude-standard", "--others", "--full-name"}
	if len(patterns) > 0 {
		args = append(args, "--")
		args = append(args, patterns...)
	}

	output, err := cmd.Dir(dir).Output("git", args...)

	if err != nil {
		return nil, err
	}
	return splitLinesFilterEmpty(output), nil
}
