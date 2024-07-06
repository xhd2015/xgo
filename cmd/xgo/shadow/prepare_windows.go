//go:build windows
// +build windows

package main

import (
	"os"
	"path/filepath"
	"strings"
)

// test on powershell:
//  $env:path="C:\Users\xhd2015\.xgo\shadow;"+$env:path

func exclude(dir string) (func(), error) {
	path := os.Getenv("path")
	list := filepath.SplitList(path)

	newList, err := excludeInList(dir, path, list)
	if err != nil {
		return nil, err
	}
	os.Setenv("path", strings.Join(newList, string(filepath.ListSeparator)))
	return func() {
		os.Setenv("path", path)
	}, nil
}
