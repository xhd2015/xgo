package git

import (
	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/gitops/git/worktree"
)

func AcquireTempWorkTree(dir string, ref string) (tmpDir string, remove func() error, err error) {
	return worktree.AcquireTemp(dir, ref)
}

func AddWorkTree(srcDir string, ref string, targetDir string) (remove func() error, err error) {
	return worktree.AddWorkTree(srcDir, ref, targetDir)
}
