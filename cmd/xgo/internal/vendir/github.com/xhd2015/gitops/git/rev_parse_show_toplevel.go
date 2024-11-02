package git

import "github.com/xhd2015/xgo/support/cmd"

func ShowToplevel(dir string) (string, error) {
	return cmd.Dir(dir).Output("git", "rev-parse", "--show-toplevel")
}
