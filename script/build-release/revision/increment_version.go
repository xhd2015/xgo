package revision

import "path/filepath"

func IncrementXgoVersion(rootDir string, amend bool, autoIncrementNumber bool) error {
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
