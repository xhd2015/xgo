package worktree

import (
	"fmt"
	"io"
	"os/exec"

	"github.com/xhd2015/xgo/support/cmd"
)

func RemoveWorktree(dir string, worktreeDir string) error {
	return removeWorktree(dir, worktreeDir, false)
}

func ForceRemoveWorktree(dir string, worktreeDir string) error {
	return removeWorktree(dir, worktreeDir, true)
}

func removeWorktree(dir string, worktreeDir string, force bool) error {
	if dir == "" {
		return fmt.Errorf("requires dir")
	}
	if worktreeDir == "" {
		return fmt.Errorf("requires worktreeDir")
	}
	// NOTE: remove -f can fail
	err := cmd.Dir(dir).Stderr(io.Discard).Stdout(io.Discard).Run("git", "worktree", "remove", "-f", worktreeDir)
	if err != nil {
		if force {
			// ignore exit code
			if _, ok := err.(*exec.ExitError); ok {
				return nil
			}
		}
		return err
	}
	return nil
}
