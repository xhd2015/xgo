package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/xhd2015/xgo/script/build-release/revision"
	"github.com/xhd2015/xgo/support/cmd"
	"github.com/xhd2015/xgo/support/filecopy"
	"github.com/xhd2015/xgo/support/git"
	"github.com/xhd2015/xgo/support/goinfo"
	"github.com/xhd2015/xgo/support/osinfo"
)

// usage:
//  go run ./script/build-release
//  go run ./script/build-release --local --local-name xgo_dev
//  go run ./script/build-release --local --local-name xgo_dev --debug
//  go run ./script/build-release --include-install-src --include-local

func main() {
	args := os.Args[1:]
	n := len(args)
	var installLocal bool
	var localName string
	var debug bool
	var includeInstallSrc bool
	var includeLocal bool
	for i := 0; i < n; i++ {
		arg := args[i]
		if arg == "--local" {
			installLocal = true
			continue
		}
		if arg == "--debug" {
			debug = true
			continue
		}
		if arg == "--local-name" {
			localName = args[i+1]
			i++
			continue
		}
		if arg == "--include-install-src" {
			includeInstallSrc = true
			continue
		}
		if arg == "--include-local" {
			includeLocal = true
			continue
		}
		fmt.Fprintf(os.Stderr, "unrecognized option: %s\n", arg)
		os.Exit(1)
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
		debug = false
	} else {
		archs = []*osArch{
			{runtime.GOOS, runtime.GOARCH},
		}
	}

	err := buildRelease("xgo-release", installLocal, localName, debug, archs, includeInstallSrc, includeLocal)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

type osArch struct {
	goos   string
	goarch string
}

func buildRelease(releaseDirName string, installLocal bool, localName string, debug bool, osArches []*osArch, includeInstallSrc bool, includeLocal bool) error {
	if installLocal && len(osArches) != 1 {
		return fmt.Errorf("--install-local requires only one target")
	}
	goVersionStr, err := goinfo.GetGoVersionOutput("go")
	if err != nil {
		return err
	}
	goVersion, err := goinfo.ParseGoVersion(goVersionStr)
	if err != nil {
		return fmt.Errorf("parsing go version: %s %w", goVersionStr, err)
	}
	var extraBuildFlags []string
	// requires at least go1.18 to support -trimpath
	// see: https://github.com/golang/go/issues/50402
	if goVersion.Major == 1 && goVersion.Minor < 18 {
		return fmt.Errorf("build-release relies on go1.18 or later to use 'go build -trimpath' option,current: %s", goVersionStr)
	} else {
		extraBuildFlags = append(extraBuildFlags, "-trimpath")
	}

	projectRoot, err := git.ShowTopLevel("")
	if err != nil {
		return err
	}
	xgoVersionStr, err := cmd.Dir(projectRoot).Output("go", "run", "./cmd/xgo", "version")
	if err != nil {
		return err
	}
	dir := filepath.Join(filepath.Join(projectRoot, releaseDirName), xgoVersionStr)
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	tmpDir, err := os.MkdirTemp("", "xgo")
	if err != nil {
		return err
	}
	if !debug {
		defer os.RemoveAll(tmpDir)
	} else {
		fmt.Printf("%s\n", tmpDir)
	}

	tmpSrcDir := projectRoot
	if false {
		tmpSrcDir := filepath.Join(tmpDir, "src")
		// use git worktree to prepare the directory for building
		// add a worktree detached at HEAD
		err = cmd.Dir(projectRoot).Run("git", "worktree", "add", "--detach", tmpSrcDir, "HEAD")
		if err != nil {
			return err
		}
		// --force: delete files even there is untracked content
		if !debug {
			defer cmd.Dir(projectRoot).Run("git", "worktree", "remove", "--force", tmpSrcDir)
		} else {
			fmt.Printf("git worktree remove --force %s\n", tmpSrcDir)
		}
	}

	// copy modified files
	modifiedFiles, err := gitListWorkingTreeChangedFiles(projectRoot)
	if err != nil {
		return err
	}
	if false {
		for _, file := range modifiedFiles {
			err := filecopy.CopyFileAll(filepath.Join(projectRoot, file), filepath.Join(tmpSrcDir, file))
			if err != nil {
				return fmt.Errorf("copying file %s: %w", file, err)
			}
		}
	}

	// update the version
	rev, err := revision.GetCommitHash("", "HEAD")
	if err != nil {
		return err
	}
	if len(modifiedFiles) > 0 {
		rev += fmt.Sprintf("_DEV_%s", time.Now().UTC().Format("2006-01-02T15:04:05Z"))
	}

	restore, err := fixupSrcDir(tmpSrcDir, rev)
	if restore != nil {
		defer restore()
	}
	if err != nil {
		return err
	}

	if debug {
		extraBuildFlags = append(extraBuildFlags, "-gcflags=all=-N -l")
	}

	for _, osArch := range osArches {
		err := buildBinaryRelease(dir, tmpSrcDir, xgoVersionStr, osArch.goos, osArch.goarch, installLocal, includeLocal, localName, extraBuildFlags)
		if err != nil {
			return fmt.Errorf("GOOS=%s GOARCH=%s:%w", osArch.goos, osArch.goarch, err)
		}
	}
	if !installLocal {
		err = generateSums(dir, filepath.Join(dir, "SHASUMS256.txt"))
		if err != nil {
			return fmt.Errorf("sum sha256: %w", err)
		}
	}
	if includeInstallSrc {
		err := createInstallSRC(filepath.Join(dir, "install-src.zip"), projectRoot)
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

func buildBinaryRelease(dir string, srcDir string, version string, goos string, goarch string, installLocal bool, includeLocal bool, localName string, extraBuildFlags []string) error {
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

	exeSuffix := osinfo.EXE_SUFFIX

	archive := filepath.Join(tmpDir, "archive")

	bins := [][2]string{
		{"xgo", "./cmd/xgo"},
		// {"exec_tool", "./cmd/exec_tool"},
		// {"trace", "./cmd/trace"},
	}

	var archiveFilesWithExe []string
	// build xgo, exec_tool and trace
	for _, bin := range bins {
		binName, binSrc := bin[0], bin[1]
		archiveFilesWithExe = append(archiveFilesWithExe, "./"+binName+exeSuffix)
		buildArgs := []string{"build", "-o", filepath.Join(tmpDir, binName) + exeSuffix}
		buildArgs = append(buildArgs, extraBuildFlags...)
		buildArgs = append(buildArgs, binSrc)
		// fmt.Printf("flags: %v\n", buildArgs)
		err = cmd.Env([]string{"GOOS=" + goos, "GOARCH=" + goarch}).
			Dir(srcDir).
			Run("go", buildArgs...)
		if err != nil {
			return err
		}
	}
	var needInstall bool
	var installCopy bool
	if installLocal {
		needInstall = true
	} else if includeLocal {
		if runtime.GOOS == goos && runtime.GOARCH == goarch {
			needInstall = true
			installCopy = true
		}
	}

	if needInstall {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		binDir := filepath.Join(homeDir, ".xgo", "bin")
		err = os.MkdirAll(binDir, 0755)
		if err != nil {
			return err
		}
		xgoBaseNameExe := "xgo" + exeSuffix
		for _, file := range archiveFilesWithExe {
			baseNameExe := filepath.Base(file)
			toBaseNameExe := baseNameExe
			if toBaseNameExe == "xgo"+exeSuffix && localName != "" {
				localNameExe := localName
				if !strings.HasSuffix(localNameExe, exeSuffix) {
					localNameExe += exeSuffix
				}
				toBaseNameExe = localNameExe
				xgoBaseNameExe = localNameExe
			}
			var err error
			if installCopy {
				err = filecopy.CopyFile(filepath.Join(tmpDir, baseNameExe), filepath.Join(binDir, toBaseNameExe))
			} else {
				err = os.Rename(filepath.Join(tmpDir, baseNameExe), filepath.Join(binDir, toBaseNameExe))
			}
			if err != nil {
				return err
			}
		}

		xgoExeName := xgoBaseNameExe
		_, lookPathErr := exec.LookPath(xgoExeName)
		if lookPathErr != nil {
			fmt.Printf("%s built successfully, you may need to add %s to your PATH variables\n", xgoExeName, binDir)
		}
		if installLocal {
			return nil
		}
	}

	// package it as a tar.gz
	err = cmd.Dir(tmpDir).Run("tar", append([]string{"-czf", archive}, archiveFilesWithExe...)...)
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
