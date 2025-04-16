package build

import (
	"fmt"
	"path/filepath"
	"strings"
)

func AppendNativeBuildEnv(env []string) []string {
	return append(env, "GOOS=", "GOARCH=")
}

// MakeGorootEnv makes a new env with GOROOT and PATH set
func MakeGorootEnv(env []string, goroot string) []string {
	goroot, err := filepath.Abs(goroot)
	if err != nil {
		panic(fmt.Errorf("failed to make goroot env: %w", err))
	}
	newEnv := make([]string, 0, len(env))
	var lastPath string
	for _, e := range env {
		if strings.HasPrefix(e, "GOROOT=") {
			continue
		}
		if strings.HasPrefix(e, "PATH=") {
			lastPath = e
			continue
		}
		newEnv = append(newEnv, e)
	}
	gorootBin := filepath.Join(goroot, "bin")
	pathEnv := gorootBin
	if lastPath != "" {
		pathEnv = pathEnv + string(filepath.ListSeparator) + strings.TrimPrefix(lastPath, "PATH=")
	}
	newEnv = append(newEnv,
		"GOROOT="+goroot,
		"PATH="+pathEnv,
	)
	return newEnv
}
