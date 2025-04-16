//go:build go1.16
// +build go1.16

package legacy

import (
	"io/fs"
	"os"
)

func WriteFile(name string, data []byte, perm fs.FileMode) error {
	return os.WriteFile(name, data, perm)
}
