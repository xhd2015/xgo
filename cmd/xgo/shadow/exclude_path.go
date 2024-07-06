package main

import (
	"fmt"
)

func excludeInList(dir string, path string, list []string) ([]string, error) {
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
		return nil, fmt.Errorf("xgo shadow not found in PATH: %s", path)
	}
	return list[:j], nil
}
