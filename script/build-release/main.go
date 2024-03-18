package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/xgo/script/build-release/revision"
	"github.com/xhd2015/xgo/support/cmd"
)

// TODO: apply build tag for development and release mode
func main() {
	err := buildRelease("xgo-release", []*osArch{
		{"darwin", "amd64"},
		{"darwin", "arm64"},
		{"linux", "amd64"},
		{"linux", "arm64"},
		{"windows", "amd64"},
		{"windows", "arm64"},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

type osArch struct {
	goos   string
	goarch string
}

func buildRelease(releaseDir string, osArches []*osArch) error {
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
	err = cmd.Run("git", "worktree", "add", tmpSrcDir)
	if err != nil {
		return err
	}
	defer cmd.Run("git", "worktree", "remove", "--force", tmpSrcDir)

	// update the version
	rev, err := revision.GetCommitHash("", "HEAD")
	if err != nil {
		return err
	}

	err = updateRevisions(tmpSrcDir, false, rev)
	if err != nil {
		return err
	}

	for _, osArch := range osArches {
		err := buildBinaryRelease(dir, tmpSrcDir, version, osArch.goos, osArch.goarch)
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

func updateRevisions(targetDir string, unlink bool, rev string) error {
	// unlink files because all files are symlink
	files := revision.GetVersionFiles(targetDir)
	if unlink {
		for _, file := range files {
			err := unlinkFile(file)
			if err != nil {
				return err
			}
		}
	}

	for _, file := range files {
		err := revision.PatchVersionFile(file, rev)
		if err != nil {
			return err
		}
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

func buildBinaryRelease(dir string, srcDir string, version string, goos string, goarch string) error {
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

	bin := filepath.Join(tmpDir, "xgo")
	archive := filepath.Join(tmpDir, "archive")

	// build xgo
	err = cmd.Env([]string{"GOOS=" + goos, "GOARCH=" + goarch}).
		Dir(srcDir).
		Run("go", "build", "-o", bin, "./cmd/xgo")
	if err != nil {
		return err
	}

	// package it as a tar.gz
	err = cmd.Dir(tmpDir).Run("tar", "-czf", archive, "./xgo")
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
