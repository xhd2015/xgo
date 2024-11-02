package worktree

import (
	"os/exec"

	"github.com/xhd2015/xgo/support/cmd"
)

func IsClean(dir string) (ok bool, err error) {
	return workTreeClean(dir, false)
}

func IndexClean(dir string) (ok bool, err error) {
	return workTreeClean(dir, true)
}

func workTreeClean(dir string, index bool) (bool, error) {
	// 0: clean, 1: not clean
	args := []string{"diff", "--exit-code", "--quiet"}
	if index {
		args = append(args, "--cached")
	} else {
		// compare HEAD with working tree
		args = append(args, "HEAD")
	}

	// exit code: 0=clean, 1=not clean
	err := cmd.Dir(dir).Run("git", args...)
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
