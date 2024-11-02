package git

import (
	"fmt"

	"github.com/xhd2015/xgo/support/cmd"
)

func GetOriginURL(dir string) (string, error) {
	if dir == "" {
		return "", fmt.Errorf("requires dir")
	}
	return cmd.Dir(dir).Output("git", "remote", "get-url", "origin")
}
