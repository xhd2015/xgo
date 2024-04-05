package git

import "github.com/xhd2015/xgo/support/cmd"

func ShowTopLevel(dir string) (string, error) {
	return cmd.Dir(dir).Output("git", "rev-parse", "--show-toplevel")
}

func GetGitDir(dir string) (string, error) {
	return cmd.Dir(dir).Output("git", "rev-parse", "--git-dir")
}
