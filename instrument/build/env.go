package build

import (
	"fmt"
	"path/filepath"
	"strings"
)

// this ensure we do not build cross-platform
// while building native binary
func AppendNativeBuildEnv(env []string) []string {
	return append(env, "GOOS=", "GOARCH=")
}

// see https://github.com/xhd2015/xgo/issues/320
// the GOEXPERIMENT,GOOS and GOARCH could affect
// building process. we make a fresh env
func EnvForNative(env []string, goroot string) []string {
	cleanEnv := MakeGorootEnv(env, goroot)
	cleanEnv = AppendNativeBuildEnv(cleanEnv)
	cleanEnv = append(cleanEnv, "GOEXPERIMENT=")
	return cleanEnv
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
