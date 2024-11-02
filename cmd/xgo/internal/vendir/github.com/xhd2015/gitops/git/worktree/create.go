package worktree

import (
	"fmt"
	"io"
	"os"
	"path"

	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/gitops/model"
	"github.com/xhd2015/xgo/support/cmd"
	"github.com/xhd2015/xgo/support/filecopy"
)

func AcquireTemp(dir string, ref string) (tmpDir string, remove func() error, err error) {
	baseName := path.Base(dir)
	if baseName == "." {
		var wd string
		wd, err = os.Getwd()
		if err != nil {
			return
		}
		baseName = path.Base(wd)
	}
	tmpDir, err = os.MkdirTemp(os.TempDir(), baseName+"-tmp-*")
	if err != nil {
		return
	}
	remove, err = AddWorkTree(dir, ref, tmpDir)
	return
}

func AddWorkTree(srcDir string, ref string, targetDir string) (remove func() error, err error) {
	if srcDir == "" {
		return nil, fmt.Errorf("requires srcDir")
	}
	if ref == "" {
		return nil, fmt.Errorf("requires ref")
	}
	if ref == model.COMMIT_WORKING {
		// use copy -R
		err = filecopy.CopyReplaceDir(srcDir, targetDir, false)
		if err != nil {
			return
		}

		remove = func() error {
			return os.RemoveAll(targetDir)
		}
		return
	}
	if targetDir == "" {
		return nil, fmt.Errorf("requires targetDir")
	}

	// remove existing worktree
	err = ForceRemoveWorktree(srcDir, targetDir)
	if err != nil {
		return
	}

	// git worktree add --detach <path>
	// -d, --detach: don't automatically associate any branch
	// with this option, the target dir shows ((no branch))
	err = cmd.Dir(srcDir).Run("git", "worktree", "add", "-f", "--detach", targetDir, ref)
	if err != nil {
		return
	}
	remove = func() error {
		return cmd.Dir(srcDir).Stderr(io.Discard).Stdout(io.Discard).Run("git", "worktree", "remove", "-f", targetDir)
	}
	return
}
