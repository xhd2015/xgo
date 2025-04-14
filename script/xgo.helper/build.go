package main

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/xhd2015/xgo/support/cmd"
	"github.com/xhd2015/xgo/support/osinfo"
)

func buildGo(goroot string) (string, error) {
	absGoroot, err := filepath.Abs(goroot)
	if err != nil {
		return "", err
	}
	src := filepath.Join(absGoroot, "src")
	bin := filepath.Join(absGoroot, "bin")
	output := filepath.Join(bin, "go.debug"+osinfo.EXE_SUFFIX)
	err = cmd.Debug().Dir(src).Env([]string{"GOROOT=" + absGoroot, "GO_BYPASS_XGO=true"}).Run(filepath.Join(bin, "go"+osinfo.EXE_SUFFIX), "build", "-gcflags=all=-N -l", "-o", output, "./cmd/go")
	if err != nil {
		return "", fmt.Errorf("build go: %w", err)
	}
	return output, nil
}

func buildCompiler(goroot string) (string, error) {
	absGoroot, err := filepath.Abs(goroot)
	if err != nil {
		return "", err
	}
	src := filepath.Join(absGoroot, "src")
	bin := filepath.Join(absGoroot, "bin")

	// /Users/xhd2015/installed/go1.23.6-copy/pkg/tool/darwin_arm64
	toolDir := filepath.Join(absGoroot, "pkg", "tool", runtime.GOOS+"_"+runtime.GOARCH)

	output := filepath.Join(toolDir, "compile.debug"+osinfo.EXE_SUFFIX)
	err = cmd.Debug().Dir(src).Env([]string{"GOROOT=" + absGoroot, "GO_BYPASS_XGO=true"}).Run(filepath.Join(bin, "go"+osinfo.EXE_SUFFIX), "build", "-gcflags=all=-N -l", "-o", output, "./cmd/compile")
	if err != nil {
		return "", fmt.Errorf("build go: %w", err)
	}
	return output, nil
}
