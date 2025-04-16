package build

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// $GOROOT/pkg/tool/linux_amd64
func GetToolPath(goroot string) (string, error) {
	runtimeOS := runtime.GOOS
	arch := runtime.GOARCH
	if runtimeOS == "" {
		return "", errors.New("cannot get runtime.GOOS")
	}
	if arch == "" {
		return "", errors.New("cannot get runtime.GOARCH")
	}
	dir := filepath.Join(goroot, "pkg", "tool", runtimeOS+"_"+arch)
	stat, err := os.Stat(dir)
	if err != nil {
		return "", err
	}
	if !stat.IsDir() {
		return "", fmt.Errorf("cover tool path is not a directory: %s", dir)
	}
	return dir, nil
}
