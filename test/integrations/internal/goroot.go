package internal

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/xhd2015/xgo/support/gitgoroot"
	"github.com/xhd2015/xgo/support/goinfo"
)

func FindRepoRoot() string {
	if _, err := os.Stat(filepath.Join("cmd", "xgo", "main.go")); err == nil {
		if cwd, err := os.Getwd(); err == nil {
			return cwd
		}
	}
	root, err := Output("", "git", "rev-parse", "--show-toplevel")
	if err != nil {
		Fatalf("find repo root: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "cmd", "xgo", "main.go")); err != nil {
		Fatalf("not xgo repo root: %s", root)
	}
	return root
}

func NormalizeGoVersion(version string) string {
	v := strings.TrimSpace(version)
	v = strings.TrimPrefix(v, "go")
	if !strings.Contains(v, ".") {
		Fatalf("invalid version: %q", version)
	}
	parts := strings.Split(v, ".")
	if len(parts) == 2 {
		v += ".0"
	}
	return "go" + v
}

func IsGitRepo(dir string) bool {
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--git-dir")
	return cmd.Run() == nil
}

func CpGoroot(src, dst string) error {
	return exec.Command("cp", "-R", src, dst).Run()
}

func EnsureGoroot(rootDir, goVersionFlag, gorootFlag string) (string, *goinfo.GoVersion) {
	var goroot string
	var err error
	if gorootFlag != "" {
		goroot = gorootFlag
	} else {
		gitVersionedDir := filepath.Join(rootDir, "go-release-git-versioned")
		goroot, err = gitgoroot.EnsureGitGoroot(gitVersionedDir, goVersionFlag, "")
		if err != nil {
			Fatalf("%v", err)
		}
	}

	goVersion, err := goinfo.GetGorootVersion(goroot)
	if err != nil {
		Fatalf("get goroot version: %v", err)
	}
	return goroot, goVersion
}

func SpawnWorktree(goroot string, keepTemp bool) (string, func()) {
	isGit := IsGitRepo(goroot)
	if isGit {
		Logf("spawning worktree...")
		wt, cleanup, err := gitgoroot.SpawnWorktree(goroot)
		if err != nil {
			Fatalf("spawn worktree: %v", err)
		}
		if keepTemp {
			Logf("keeping worktree: %s", wt)
			return wt, func() {}
		}
		return wt, cleanup
	}

	tmpDir, err := os.MkdirTemp("", "xgo-test-repeat-*")
	if err != nil {
		Fatalf("create temp dir: %v", err)
	}
	gorootBase := filepath.Base(goroot)
	workGoroot := filepath.Join(tmpDir, gorootBase)
	Logf("syncing goroot to temp dir...")
	if err := CpGoroot(goroot, workGoroot); err != nil {
		Fatalf("sync goroot: %v", err)
	}
	if keepTemp {
		Logf("keeping temp dir: %s", tmpDir)
		return workGoroot, func() {}
	}
	return workGoroot, func() { os.RemoveAll(tmpDir) }
}
