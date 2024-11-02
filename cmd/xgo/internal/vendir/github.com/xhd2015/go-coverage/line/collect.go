package line

import (
	"fmt"
	"strings"

	diff "github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/go-coverage/diff/myers"
	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/go-coverage/git"
	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/go-coverage/model"
)

// type ChangeType string

// const (
// 	ChangeTypeUnchanged ChangeType = "unchanged"
// 	ChangeTypeNew       ChangeType = "new"
// 	ChangeTypeUpdated   ChangeType = "updated"
// )

// type FileLines struct {
// 	ChangeType  ChangeType
// 	LineMapping map[int]int // effective when ChangeTypeUpdated==updated
// }

type LineMapping map[int]int
type DeletedLineMapping map[int]bool

// CollectUnchangedLinesMapping
func CollectUnchangedLinesMapping(dir string, oldCommit string, newCommit string) (map[string]LineMapping, map[string]DeletedLineMapping, error) {
	gitDiff := git.NewGitDiff(dir, oldCommit, newCommit)
	return CollectUnchangedLinesMappingWithDetails(gitDiff, nil)
}
func MakeGitDiffRanger(gitDiff *git.GitDiff, filterFile func(file string) bool) (func(fn func(file string, oldFile string, newContent string, oldContent string) bool) error, error) {
	fileDetails, err := gitDiff.AllFilesDetailsV2()
	if err != nil {
		return nil, err
	}
	return func(fn func(file string, oldFile string, newContent string, oldContent string) bool) error {
		for file, fd := range fileDetails {
			if filterFile != nil && !filterFile(file) {
				continue
			}
			if fd.IsNew || !fd.ContentChanged {
				continue
			}
			oldFile := file
			if fd.RenamedFrom != "" {
				oldFile = fd.RenamedFrom
			}

			// get content
			newContent, err := gitDiff.GetNewContent(file)
			if err != nil {
				return err
			}
			oldContent, err := gitDiff.GetOldContent(oldFile)
			if err != nil {
				return err
			}
			if !fn(file, oldFile, newContent, oldContent) {
				break
			}
		}
		return nil
	}, nil
}

func CollectUnchangedLinesMappingWithDetails(gitDiff *git.GitDiff, filterFile func(file string) bool) (map[string]LineMapping, map[string]DeletedLineMapping, error) {
	ranger, err := MakeGitDiffRanger(gitDiff, filterFile)
	if err != nil {
		return nil, nil, err
	}

	mapping := make(map[model.PkgFile]LineMapping)
	deleteMapping := make(map[model.PkgFile]DeletedLineMapping)
	err = ranger(func(file, oldFile, newContent, oldContent string) bool {
		newLines := strings.Split(newContent, "\n")
		oldLines := strings.Split(oldContent, "\n")

		lineMapping, deletedLines := diff.ComputeBlockMapping(oldLines, newLines)

		trimFile := strings.TrimPrefix(file, "/")
		mapping[trimFile] = lineMapping
		deleteMapping[trimFile] = deletedLines
		return true
	})
	if err != nil {
		return nil, nil, err
	}
	return mapping, deleteMapping, nil
}

func CollectChanges(forChangedFiles func(fn func(file string, oldFile string, newContent string, oldContent string) bool) error) (model.FileChanges, error) {
	var err error
	changesMapping := make(model.FileChanges)
	forErr := forChangedFiles(func(file string, oldFile string, newContent string, oldContent string) bool {
		changes, diffErr := Diff(newContent, oldContent)
		if diffErr != nil {
			err = fmt.Errorf("%s: %w", file, diffErr)
			return false
		}

		changesMapping[strings.TrimPrefix(file, "/")] = changes
		return true
	})
	if forErr != nil {
		return nil, forErr
	}
	if err != nil {
		return nil, err
	}
	return changesMapping, nil
}

func Diff(newContent string, oldContent string) (*model.LineChanges, error) {
	newLines := strings.Split(newContent, "\n")
	oldLines := strings.Split(oldContent, "\n")

	changes, err := diff.DiffChanges(oldLines, newLines)
	if err != nil {
		return nil, fmt.Errorf("diff changes of: %w", err)
	}

	return &model.LineChanges{
		NewLineCount: int64(len(newLines)),
		OldLineCount: int64(len(oldLines)),
		Changes:      changes,
	}, nil
}
