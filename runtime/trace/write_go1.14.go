//go:build !go1.16
// +build !go1.16

package trace

import (
	"io/ioutil"
	"os"
)

func WriteFile(name string, data []byte, perm os.FileMode) error {
	return ioutil.WriteFile(name, data, perm)
}
