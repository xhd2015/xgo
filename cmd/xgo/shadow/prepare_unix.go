// for go <= 1.17 darwin || linux
// for go > 1.18  unix is enough
//
//go:build darwin || linux || unix
// +build darwin linux unix

package main

import (
	"os"
	"path/filepath"
	"strings"
)

func exclude(dir string) (func(), error) {
	path := os.Getenv("PATH")
	list := filepath.SplitList(path)

	newList, err := excludeInList(dir, path, list)
	if err != nil {
		return nil, err
	}
	os.Setenv("PATH", strings.Join(newList, string(filepath.ListSeparator)))
	return func() {
		os.Setenv("PATH", path)
	}, nil
}
