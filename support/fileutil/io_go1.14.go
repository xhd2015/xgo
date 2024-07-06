//go:build !go1.16
// +build !go1.16

package fileutil

import (
	"io/ioutil"
)

func ReadFile(file string) ([]byte, error) {
	return ioutil.ReadFile(file)
}

func WriteFile(file string, data []byte) error {
	return ioutil.WriteFile(file, data, 0755)
}
