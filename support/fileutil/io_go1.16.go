//go:build go1.16
// +build go1.16

package fileutil

import "os"

func ReadFile(file string) ([]byte, error) {
	return os.ReadFile(file)
}

func WriteFile(file string, data []byte) error {
	return os.WriteFile(file, data, 0755)
}
