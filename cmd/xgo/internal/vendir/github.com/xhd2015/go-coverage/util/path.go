package util

import (
	"fmt"
	"os"
	"path"
)

// if pathName is "", cwd is returned
func ToAbsPath(pathName string) (string, error) {
	// if pathName == "" {
	// 	return "", fmt.Errorf("dir should not be empty")
	// }
	if path.IsAbs(pathName) {
		return pathName, nil
	}
	// _, err := os.Stat(pathName)
	// if err != nil {
	// 	return "", fmt.Errorf("%s not exists:%v", pathName, err)
	// }
	// if !f.IsDir() {
	// 	return "", fmt.Errorf("%s is not a dir", pathName)
	// }
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get cwd error:%v", err)
	}
	return path.Join(cwd, pathName), nil
}
