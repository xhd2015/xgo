package git

import "github.com/xhd2015/xgo/support/cmd"

// worktree root
func ShowTopLevel(dir string) (string, error) {
	return cmd.Dir(dir).Output("git", "rev-parse", "--show-toplevel")
}

// when multiple worktree exists, this denotes to
// the only one .git
func GetGitDir(dir string) (string, error) {
	return cmd.Dir(dir).Output("git", "rev-parse", "--git-dir")
}
