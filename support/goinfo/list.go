package goinfo

import (
	"strings"

	"github.com/xhd2015/xgo/support/cmd"
)

// go list ./pkg
func ListPackages(dir string, mod string, args []string) ([]string, error) {
	flags := []string{"list"}
	if mod != "" {
		flags = append(flags, "-mod="+mod)
	}
	flags = append(flags, args...)
	output, err := cmd.Dir(dir).Output("go", flags...)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(output, "\n")
	return lines, nil
}
