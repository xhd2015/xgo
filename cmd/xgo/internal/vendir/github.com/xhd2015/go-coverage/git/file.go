package git

import "context"

func ForEachDiff(ctx context.Context, dir string, oldRef string, newRef string, fn func(ctx context.Context, gitDiff *GitDiff, file string, detail *FileDetail) error) error {
	gitDiff := NewGitDiff(dir, oldRef, newRef)

	fileDetails, err := gitDiff.AllFilesDetailsV2()
	if err != nil {
		return err
	}
	for file, detail := range fileDetails {
		err := fn(ctx, gitDiff, file, detail)
		if err != nil {
			return err
		}
	}
	return nil
}
