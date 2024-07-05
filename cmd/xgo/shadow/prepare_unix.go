//go:build unix
// +build unix

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func exclude(dir string) (func(), error) {
	path := os.Getenv("PATH")

	list := filepath.SplitList(path)
	n := len(list)
	j := 0
	for i := 0; i < n; i++ {
		d := list[i]
		if d == dir {
			continue
		}
		list[j] = d
		j++
	}
	if j == n {
		return nil, fmt.Errorf("xgo shadow not found in PATH: %s", dir)
	}
	os.Setenv("PATH", strings.Join(list[:j], string(filepath.ListSeparator)))
	return func() {
		os.Setenv("PATH", path)
	}, nil
}
