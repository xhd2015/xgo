//go:build windows
// +build windows

package main

import (
	"fmt"
)

var xgo_shadow int = "not support windows yet"

func exclude(dir string) (func(), error) {
	return nil, fmt.Errorf("windows not supported")
}
