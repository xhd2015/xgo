package git

import (
	"fmt"
	"sync"

	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/go-coverage/git/gitops"
	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/go-coverage/model"
)

type GitDiff struct {
	dir       string
	oldCommit string
	newCommit string

	oldGit *GitSnapshot
	newGit *GitSnapshot

	mergeOnce   sync.Once
	newToOld    map[string]string // merged updated and renamed
	newToOldErr error

	renameOnce sync.Once
	renames    map[string]string // renamed, new to old
	renameErr  error

	updateOnce  sync.Once
	updateFiles []string // file list
	updateErr   error

	addedOnce  sync.Once
	addedFiles []string // file list
	addedErr   error

	fileDetailsOnce sync.Once
	fileDetails     map[string]*FileDetail
	fileDetailsErr  error

	fileDetailsV2Once sync.Once
	fileDetailsV2     map[string]*model.FileDetail
	fileDetailsV2Err  error
}

type ChangeType string

const (
	ChangeTypeUnchanged ChangeType = "unchanged"
	ChangeTypeUpdated   ChangeType = "updated"
	ChangeTypeAdded     ChangeType = "added"
)

type FileDetail = model.FileDetail

func NewGitDiff(dir string, oldCommit string, newCommit string) *GitDiff {
	return &GitDiff{
		dir:       dir,
		oldCommit: oldCommit,
		newCommit: newCommit,
		oldGit:    NewSnapshot(dir, oldCommit),
		newGit:    NewSnapshot(dir, newCommit),
	}
}

func (c *GitDiff) AllFiles() ([]string, error) {
	return c.newGit.ListFiles()
}

// AllFilesDetails deprecated, it has bug.
// use AllFilesDetailsV2 instead
func (c *GitDiff) AllFilesDetails() (map[string]*FileDetail, error) {
	c.fileDetailsOnce.Do(func() {
		allFiles, err := c.AllFiles()
		if err != nil {
			c.fileDetailsErr = fmt.Errorf("get all files err:%v", err)
			return
		}
		renames, err := c.GetRenames()
		if err != nil {
			c.fileDetailsErr = fmt.Errorf("get rename files err:%v", err)
			return
		}
		updates, err := c.GetUpdates()
		if err != nil {
			c.fileDetailsErr = fmt.Errorf("get updates err:%v", err)
			return
		}
		added, err := c.GetNew()
		if err != nil {
			c.fileDetailsErr = fmt.Errorf("get added err:%v", err)
			return
		}
		updateMap := make(map[string]bool, len(updates))
		for _, file := range updates {
			updateMap[file] = true
		}

		addedMap := make(map[string]bool, len(added))
		for _, file := range added {
			addedMap[file] = true
		}

		fileDetailsMap := make(map[string]*FileDetail, len(allFiles))
		for _, file := range allFiles {
			fileDetailsMap[file] = &FileDetail{
				IsNew:          addedMap[file],
				RenamedFrom:    renames[file],
				ContentChanged: updateMap[file],
			}
		}
		c.fileDetails = fileDetailsMap
	})
	return c.fileDetails, c.fileDetailsErr
}

func (c *GitDiff) AllFilesDetailsV2() (map[string]*model.FileDetail, error) {
	c.fileDetailsV2Once.Do(func() {
		c.fileDetailsV2, c.fileDetailsV2Err = gitops.DiffCommit(c.dir, c.newCommit, c.oldCommit, nil)
	})
	return c.fileDetailsV2, c.fileDetailsV2Err
}
func (c *GitDiff) GetUpdateAndRenames() (newToOld map[string]string, err error) {
	c.mergeOnce.Do(func() {
		renames, err := c.GetRenames()
		if err != nil {
			c.newToOldErr = err
			return
		}
		updates, err := c.GetUpdates()
		if err != nil {
			c.newToOldErr = err
			return
		}
		newToOld := make(map[string]string, len(updates)+len(renames))
		for k, v := range renames {
			newToOld[k] = v
		}
		for _, u := range updates {
			if _, ok := renames[u]; ok {
				c.newToOldErr = fmt.Errorf("invalid file: %s found both renamed and updated", u)
				return
			}
			newToOld[u] = u
		}
		c.newToOld = newToOld
	})
	return c.newToOld, c.newToOldErr
}

func (c *GitDiff) GetRenames() (newToOld map[string]string, err error) {
	c.renameOnce.Do(func() {
		c.renames, c.renameErr = FindRenames(c.dir, c.oldCommit, c.newCommit)
	})
	return c.renames, c.renameErr
}
func (c *GitDiff) GetUpdates() ([]string, error) {
	c.updateOnce.Do(func() {
		repo := &GitRepo{Dir: c.dir}
		updates, err := repo.FindUpdate(c.oldCommit, c.newCommit)
		if err != nil {
			c.updateErr = err
			return
		}
		c.updateFiles = updates
	})
	return c.updateFiles, c.updateErr
}
func (c *GitDiff) GetNew() ([]string, error) {
	c.addedOnce.Do(func() {
		repo := &GitRepo{Dir: c.dir}
		added, err := repo.FindAdded(c.oldCommit, c.newCommit)
		if err != nil {
			c.addedErr = err
			return
		}
		c.addedFiles = added
	})
	return c.addedFiles, c.addedErr
}
func (c *GitDiff) GetNewContent(newFile string) (string, error) {
	return c.newGit.GetContent(newFile)
}

func (c *GitDiff) GetOldContent(oldFile string) (string, error) {
	return c.oldGit.GetContent(oldFile)
}
func (c *GitDiff) GetOldContentNewFile(newFile string) (string, error) {
	oldFile, ok, err := c.GetOldFile(newFile)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", fmt.Errorf("file does not exist in old:%s", newFile)
	}
	return c.oldGit.GetContent(oldFile)
}

func (c *GitDiff) GetOldFile(newFile string) (oldFile string, ok bool, err error) {
	newToOld, err := c.GetUpdateAndRenames()
	if err != nil {
		return "", false, err
	}
	oldFile, ok = newToOld[newFile]
	return
}
