package git

import (
	"errors"
	"fmt"
	"os/exec"

	"github.com/xhd2015/xgo/support/cmd"
)

func ListFile(dir string, ref string) ([]string, error) {
	return ListFilePatterns(dir, ref, nil)
}

func ListFilePatterns(dir string, ref string, patterns []string) ([]string, error) {
	if ref == "" {
		return nil, fmt.Errorf("requires revision")
	}
	args := []string{"ls-files"}
	if ref != COMMIT_WORKING {
		args = append(args, "--with-tree", ref)
	} else {
		// for working
		args = append(args, "--exclude-standard", "-co")
	}
	if len(patterns) > 0 {
		args = append(args, "--")
		args = append(args, patterns...)
	}
	res, err := cmd.Dir(dir).Output("git", args...)
	if err != nil {
		return nil, err
	}
	return splitLinesFilterEmpty(res), nil
}

// extraFlags: --diff-filter=A
func diffFiles(dir string, ref string, compareRef string, extraFlags []string, patterns []string) (string, error) {
	if ref == "" {
		return "", fmt.Errorf("requires ref")
	}
	if compareRef == "" {
		return "", fmt.Errorf("requires compareRef")
	}

	// git diff --diff-filter=A --name-only --ignore-submodules "$compareRef" "$ref" -- ${patterns} || true
	// --relative: always respect current dir. without this, when dir is a sub dir from toplevle, diff still returns full path
	args := []string{"-c", "core.fileMode=false", "diff", "--relative"}
	args = append(args, extraFlags...)
	args = append(args, "--name-only", "--ignore-submodules", compareRef)
	if ref != COMMIT_WORKING {
		args = append(args, ref)
	}

	if len(patterns) > 0 {
		args = append(args, "--")
		args = append(args, patterns...)
	}

	output, err := cmd.Dir(dir).Output("git", args...)
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return "", nil
		}
		return "", err
	}
	return output, nil
}
