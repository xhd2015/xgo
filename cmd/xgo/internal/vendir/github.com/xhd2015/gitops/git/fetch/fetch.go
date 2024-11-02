package fetch

import (
	"fmt"
	"strings"
)

type Options struct {
	Branch            string // optional branch
	Unshawllow        bool
	Depth             int  // --depth
	RequireSubmodules bool // default --recurse-submodules=no
	AllTags           bool // --tags
}

// `remote` like origin
func FormatFetch(remote string, opts *Options) []string {
	if opts == nil {
		opts = &Options{}
	}
	depth := opts.Depth
	branch := opts.Branch
	unshawllow := opts.Unshawllow
	requireSubmodules := opts.RequireSubmodules
	allTags := opts.AllTags

	args := []string{"fetch"}
	if depth > 0 {
		args = append(args, fmt.Sprintf("--depth=%d", depth))
	}
	if unshawllow {
		args = append(args, "--unshallow")
	}
	if !requireSubmodules {
		args = append(args, "--recurse-submodules=no")
	}
	if allTags {
		args = append(args, "--tags")
	}
	if remote != "" {
		args = append(args, remote)
		if branch != "" {
			args = append(args, strings.TrimPrefix(branch, "origin/"))
		}
	}
	return args
}
