package revision

import (
	"path/filepath"

	"github.com/xhd2015/xgo/support/git"
)

func IncrementXgoVersion(rootDir string, amend bool, autoIncrementNumber bool) error {
	// first list all changed files, if nothing changes at all, do nothing.
	changedFiles, err := git.ListFileUpdates(rootDir, git.COMMIT_WORKING, "HEAD", nil)
	if err != nil {
		return err
	}
	if len(changedFiles) == 0 {
		return nil
	}
	refLast := "HEAD"
	if amend {
		refLast = "HEAD~1"
	}

	commitHash, err := GetCommitHash(rootDir, refLast)
	if err != nil {
		return err
	}

	// due to the nature of git, we cannot
	// know the commit hash of current commit
	// which has not yet happened, so we add
	// suffix "+1" to indicate this
	rev := commitHash + "+1"

	relVersionFiles := []string{GetXgoVersionFile("")}
	for _, relFile := range relVersionFiles {
		file := filepath.Join(rootDir, relFile)
		content, err := GetFileContent(rootDir, commitHash, relFile)
		if err != nil {
			return err
		}
		version, err := GetVersionNumber(content)
		if err != nil {
			return err
		}
		err = PatchVersionFile(file, "", rev, autoIncrementNumber, version+1)
		if err != nil {
			return err
		}
	}
	return nil
}
