package git

import (
	"fmt"

	"github.com/xhd2015/xgo/support/cmd"
)

// worktree root
func ShowTopLevel(dir string) (string, error) {
	return cmd.Dir(dir).Output("git", "rev-parse", "--show-toplevel")
}

// when multiple worktree exists, this denotes to
// the only one .git
func GetGitDir(dir string) (string, error) {
	return cmd.Dir(dir).Output("git", "rev-parse", "--git-dir")
}

func GetCurrentBranch(dir string) (string, error) {
	return cmd.Dir(dir).Output("git", "branch", "--show-current")
}

func FetchRef(dir string, origin string, ref string) error {
	if origin == "" {
		return fmt.Errorf("requires origin")
	}
	if ref == "" {
		return fmt.Errorf("requires ref")
	}
	return cmd.Dir(dir).Run("git", "fetch", origin, ref)
}
