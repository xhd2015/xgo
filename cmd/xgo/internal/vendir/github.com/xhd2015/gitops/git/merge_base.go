package git

import (
	"os/exec"

	"github.com/xhd2015/xgo/support/cmd"
)

func Contains(dir string, head string, prev string) (bool, error) {
	return IsAncesorOf(dir, prev, head)
}

func IsAncesorOf(dir string, early string, head string) (bool, error) {
	err := cmd.Dir(dir).Run("git", "merge-base", "--is-ancestor", early, head)
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
