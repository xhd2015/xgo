package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/xhd2015/xgo/script/build-release/revision"
	"github.com/xhd2015/xgo/support/cmd"
)

// usage:
//  go run ./script/build-release
//  go run ./script/build-release --local --local-name xgo_exp

// TODO: apply build tag for development and release mode
func main() {
	args := os.Args[1:]
	n := len(args)
	var installLocal bool
	var localName string
	for i := 0; i < n; i++ {
		arg := args[i]
		if arg == "--local" {
			installLocal = true
			continue
		}
		if arg == "--local-name" {
			localName = args[i+1]
			i++
			continue
		}
	}
	var archs []*osArch
	if !installLocal {
		archs = []*osArch{
			{"darwin", "amd64"},
			{"darwin", "arm64"},
			// {"darwin", "arm"}, // not supported, both arm and arm32
			{"linux", "amd64"},
			{"linux", "arm64"},
			{"linux", "arm"}, // NOTE: not arm32
			{"windows", "amd64"},
			{"windows", "arm64"},
		}
	} else {
		archs = []*osArch{
			{runtime.GOOS, runtime.GOARCH},
		}
	}

	err := buildRelease("xgo-release", installLocal, localName, archs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

type osArch struct {
	goos   string
	goarch string
}

func buildRelease(releaseDir string, installLocal bool, localName string, osArches []*osArch) error {
	if installLocal && len(osArches) != 1 {
		return fmt.Errorf("--install-local requires only one target")
	}
	version, err := cmd.Output("go", "run", "./cmd/xgo", "version")
	if err != nil {
		return err
	}
	dir := filepath.Join(releaseDir, version)
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	tmpDir, err := os.MkdirTemp("", "xgo")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	tmpSrcDir := filepath.Join(tmpDir, "src")

	// use git worktree to prepare the directory for building
	// add a worktree detached at HEAD
	err = cmd.Run("git", "worktree", "add", "--detach", tmpSrcDir, "HEAD")
	if err != nil {
		return err
	}
	// --force: delete files even there is untracked content
	defer cmd.Run("git", "worktree", "remove", "--force", tmpSrcDir)
	// update the version
	rev, err := revision.GetCommitHash("", "HEAD")
	if err != nil {
		return err
	}

	err = fixupSrcDir(tmpSrcDir, rev)
	if err != nil {
		return err
	}

	for _, osArch := range osArches {
		err := buildBinaryRelease(dir, tmpSrcDir, version, osArch.goos, osArch.goarch, installLocal, localName)
		if err != nil {
			return fmt.Errorf("GOOS=%s GOARCH=%s:%w", osArch.goos, osArch.goarch, err)
		}
	}
	err = generateSums(dir, filepath.Join(dir, "SHASUMS256.txt"))
	if err != nil {
		return fmt.Errorf("sum sha256: %w", err)
	}
	return nil
}

func unlinkFile(file string) error {
	content, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	err = os.RemoveAll(file)
	if err != nil {
		return err
	}
	return os.WriteFile(file, content, 0755)
}

func generateSums(dir string, sumFile string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	args := []string{
		"-a",
		"256",
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(name, "xgo") {
			continue
		}
		args = append(args, name)
	}
	output, err := cmd.Dir(dir).Output("shasum", args...)
	if err != nil {
		return err
	}
	if !strings.HasSuffix(output, "\n") {
		output = output + "\n"
	}
	err = os.WriteFile(sumFile, []byte(output), 0755)
	if err != nil {
		return err
	}
	return nil
}

// shasum -a 256 *.tar.gz
// SHASUMS256.txt example:
//
// c2876990b545be8396b7d13f0f9c3e23b38236de8f0c9e79afe04bcf1d03742e  xgo1.0.0-darwin-arm64.tar.gz
// 6ae476cb4c3ab2c81a94d1661070e34833e4a8bda3d95211570391fb5e6a3cc0  xgo1.0.0-darwin-amd64.tar.gz

func buildBinaryRelease(dir string, srcDir string, version string, goos string, goarch string, installLocal bool, localName string) error {
	if version == "" {
		return fmt.Errorf("requires version")
	}
	if goos == "" {
		return fmt.Errorf("requires GOOS")
	}
	if goarch == "" {
		return fmt.Errorf("requires GOARCH")
	}
	_, err := os.Stat(dir)
	if err != nil {
		return err
	}
	tmpDir, err := os.MkdirTemp("", "xgo-release")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	archive := filepath.Join(tmpDir, "archive")

	bins := [][2]string{
		{"xgo", "./cmd/xgo"},
		{"exec_tool", "./cmd/exec_tool"},
	}
	var archiveFiles []string
	// build xgo and exec_tool
	for _, bin := range bins {
		binName, binSrc := bin[0], bin[1]
		archiveFiles = append(archiveFiles, "./"+binName)
		err = cmd.Env([]string{"GOOS=" + goos, "GOARCH=" + goarch}).
			Dir(srcDir).
			Run("go", "build", "-o", filepath.Join(tmpDir, binName), binSrc)
		if err != nil {
			return err
		}
	}

	if installLocal {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		binDir := filepath.Join(homeDir, ".xgo", "bin")
		for _, file := range archiveFiles {
			baseName := filepath.Base(file)
			toBaseName := baseName
			if toBaseName == "xgo" && localName != "" {
				toBaseName = localName
			}
			err := os.Rename(filepath.Join(tmpDir, baseName), filepath.Join(binDir, toBaseName))
			if err != nil {
				return err
			}
		}
		return nil
	}

	// package it as a tar.gz
	err = cmd.Dir(tmpDir).Run("tar", append([]string{"-czf", archive}, archiveFiles...)...)
	if err != nil {
		return err
	}

	// mv the release to dir
	targetArchive := filepath.Join(dir, fmt.Sprintf("xgo%s-%s-%s.tar.gz", version, goos, goarch))
	err = os.Rename(archive, targetArchive)
	if err != nil {
		return err
	}

	return nil
}
