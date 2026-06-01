package build

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
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

// GetHostGOOS returns the OS of the machine running the current process
// (the compiler host). This is always runtime.GOOS.
//
// Example: on an Apple Silicon Mac, this returns "darwin".
func GetHostGOOS() string {
	return runtime.GOOS
}

// GetHostGOARCH returns the architecture of the machine running the current
// process. This is always runtime.GOARCH.
//
// Example: on an Apple Silicon Mac, this returns "arm64".
func GetHostGOARCH() string {
	return runtime.GOARCH
}

// GetTargetGOOS returns the OS that the compiled binary will run on.
// If the GOOS env var is set, that value is returned;
// otherwise falls back to GetHostGOOS().
//
// Host vs target matters for cross-compilation. For example, when compiling
// on a macOS host for a Linux target (GOOS=linux), the macOS system linker
// (clang/ld) cannot produce Linux binaries — so build flags like
// -linkmode=external must only be used for darwin targets.
//
// See: https://pkg.go.dev/cmd/go#hdr-Environment_variables
func GetTargetGOOS() string {
	if goos := os.Getenv("GOOS"); goos != "" {
		return goos
	}
	return GetHostGOOS()
}

// GetTargetGOARCH returns the architecture that the compiled binary will run
// on. If the GOARCH env var is set, that value is returned;
// otherwise falls back to GetHostGOARCH().
//
// Unlike GOOS, cross-compiling GOARCH within the same OS is usually
// supported (e.g. darwin/arm64 host → darwin/amd64 target using the macOS
// linker). This distinction is important when choosing linker flags.
//
// See: https://pkg.go.dev/cmd/go#hdr-Environment_variables
func GetTargetGOARCH() string {
	if goarch := os.Getenv("GOARCH"); goarch != "" {
		return goarch
	}
	return GetHostGOARCH()
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
