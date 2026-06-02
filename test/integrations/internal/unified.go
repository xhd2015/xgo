package internal

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/xhd2015/xgo/support/gitgoroot"
)

func CreateUnifiedWorktree(baseGoroot, progGoroot, fbGoroot string) (wt string, commitProg string, commitFB string, cleanup func(), err error) {
	// Get the initial commit from the base git repo.
	initialCommit, err := gitOutput(baseGoroot, "rev-parse", "HEAD")
	if err != nil {
		return "", "", "", nil, fmt.Errorf("get initial commit: %w", err)
	}

	// Spawn a worktree from the base goroot (always kept).
	wtObj, wtCleanup, err := gitgoroot.SpawnWorktree(baseGoroot)
	if err != nil {
		return "", "", "", nil, fmt.Errorf("spawn unified worktree: %w", err)
	}
	cleanup = func() {
		wtCleanup()
		os.RemoveAll(wtObj)
	}

	// Apply programmatic patch diff → commit.
	applyDir(wtObj, filepath.Join(progGoroot, "src"), filepath.Join(wtObj, "src"))
	if err := gitRun(wtObj, "add", "-A", "src/"); err != nil {
		return "", "", "", nil, fmt.Errorf("stage prog src: %w", err)
	}
	if err := gitRun(wtObj, "-c", "user.name=xgo-by-xhd2015", "-c", "user.email=xhd2015@gmail.com", "commit", "--no-verify", "-m", "programmatic patch [xgo]"); err != nil {
		return "", "", "", nil, fmt.Errorf("commit prog: %w", err)
	}
	commitProg, err = gitOutput(wtObj, "rev-parse", "HEAD")
	if err != nil {
		return "", "", "", nil, fmt.Errorf("get prog commit: %w", err)
	}

	// Reset to initial, apply file-based patch diff → commit.
	if err := gitRun(wtObj, "reset", "--hard", initialCommit); err != nil {
		return "", "", "", nil, fmt.Errorf("reset to initial: %w", err)
	}
	applyDir(wtObj, filepath.Join(fbGoroot, "src"), filepath.Join(wtObj, "src"))
	if err := gitRun(wtObj, "add", "-A", "src/"); err != nil {
		return "", "", "", nil, fmt.Errorf("stage fb src: %w", err)
	}
	if err := gitRun(wtObj, "-c", "user.name=xgo-by-xhd2015", "-c", "user.email=xhd2015@gmail.com", "commit", "--no-verify", "-m", "file-based patch [xgo]"); err != nil {
		return "", "", "", nil, fmt.Errorf("commit fb: %w", err)
	}
	commitFB, err = gitOutput(wtObj, "rev-parse", "HEAD")
	if err != nil {
		return "", "", "", nil, fmt.Errorf("get fb commit: %w", err)
	}

	// Leave the worktree at the programmatic commit for inspection.
	if err := gitRun(wtObj, "checkout", commitProg); err != nil {
		return "", "", "", nil, fmt.Errorf("checkout prog commit: %w", err)
	}

	return wtObj, commitProg, commitFB, cleanup, nil
}

// applyDir copies srcDir to dstDir with cp -R.
// First removes dstDir so the copy is a clean replacement.
func applyDir(workDir, srcDir, dstDir string) {
	os.RemoveAll(dstDir)
	if err := RunLogged(workDir, nil, "cp", "-R", srcDir, dstDir); err != nil {
		Fatalf("copy src: %v", err)
	}
}

func gitRun(dir string, args ...string) error {
	return RunLogged(dir, nil, "git", args...)
}

func gitOutput(dir string, args ...string) (string, error) {
	return Output(dir, "git", args...)
}
