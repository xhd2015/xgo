package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func RunXgoSetup(xgoHome string, useFilePatches bool, goroot string, inPlace bool, skipRebuildCompilerAndGo bool) (string, error) {
	// Find repo root for "go run ./cmd/xgo"
	repoRoot := FindRepoRoot()

	args := []string{"run", "-tags=dev", filepath.Join(repoRoot, "cmd", "xgo"), "setup",
		"--xgo-home", xgoHome,
		"--with-goroot", goroot,
	}
	if useFilePatches {
		args = append(args, "--use-file-patches=true")
	} else {
		args = append(args, "--use-file-patches=false")
	}
	if inPlace {
		args = append(args, "--patch-goroot-in-place")
	}
	if skipRebuildCompilerAndGo {
		args = append(args, "--skip-rebuild-compiler-and-go")
	}

	// Ensure clean xgo-home each run
	os.RemoveAll(xgoHome)

	out, err := Output("", "go", args...)
	if err != nil {
		return "", fmt.Errorf("xgo setup: %w", err)
	}
	instrumentGoroot := strings.TrimSpace(out)
	if instrumentGoroot == "" {
		return "", fmt.Errorf("xgo setup: empty output")
	}
	return instrumentGoroot, nil
}
