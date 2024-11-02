package merge

import (
	"fmt"
	"strings"

	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/go-coverage/git"
	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/go-coverage/sh"
)

func MergeGit(old Profile, new Profile, modPrefix string, dir string, oldCommit string, newCommit string) (Profile, error) {
	_, merged, _, err := MergeGitDiff(old, new, modPrefix, dir, oldCommit, newCommit)
	return merged, err
}
func MergeGitDiff(old Profile, new Profile, modPrefix string, dir string, oldCommit string, newCommit string) (actualModPrefix string, merged Profile, gitDiff *git.GitDiff, err error) {
	actualModPrefix = modPrefix
	if modPrefix == "" || modPrefix == "auto" {
		actualModPrefix, err = GetModPath(dir)
		if err != nil {
			err = fmt.Errorf("get mod path err:%v", err)
			return
		}
	} else if strings.HasPrefix(actualModPrefix, "/") || strings.HasSuffix(actualModPrefix, "/") {
		err = fmt.Errorf("modPrefix must not start or end with '/':%s", actualModPrefix)
		return
	}

	gitDiff = git.NewGitDiff(dir, oldCommit, newCommit)

	fileDetails, err := gitDiff.AllFilesDetailsV2()
	if err != nil {
		return
	}
	getFileUpdateDetail := func(pkgFile string) (isNew bool, hasUpdate bool, oldFile string) {
		// trim modPrefix
		relativeFile := strings.TrimPrefix(pkgFile, actualModPrefix)
		relativeFile = strings.TrimPrefix(relativeFile, "/")
		relativeFile = strings.TrimPrefix(relativeFile, ".")
		fd := fileDetails[relativeFile]
		if fd == nil {
			panic(fmt.Errorf("file not found: %v", relativeFile))
		}
		isNew = fd.IsNew
		if isNew {
			return
		}
		oldFile = relativeFile
		if fd.RenamedFrom != "" {
			oldFile = fd.RenamedFrom
		}
		oldFile = actualModPrefix + "/" + oldFile
		hasUpdate = fd.ContentChanged
		return
	}

	getOldContent := func(file string) (string, error) {
		return gitDiff.GetOldContent(strings.TrimPrefix(file, actualModPrefix+"/"))
	}
	getNewContent := func(file string) (string, error) {
		return gitDiff.GetNewContent(strings.TrimPrefix(file, actualModPrefix+"/"))
	}

	merged, err = Merge(old, getOldContent, new, getNewContent, MergeOptions{
		GetFileUpdateDetail: getFileUpdateDetail,
	})
	return
}

type MergeOptions struct {
	// GetFileUpdateDetail only return file that have changed
	// GetFileUpdateDetail() returns files that have content updates,including: content changed files, new files.
	// if isNew == true, then the given file is completely new
	// otherwise, if oldFile == "", then the file is not updated, which means all blocks should be merged
	//
	GetFileUpdateDetail func(newFile string) (isNew bool, hasUpdate bool, oldFile string)
}

// Merge merge 2 profiles with their code diffs
// for files there are 2 independent conditions:
//  1.name changed  2.content changed
func Merge(old Profile, oldCodeGetter func(f string) (string, error), newProfile Profile, newCodeGetter func(f string) (string, error), opts MergeOptions) (Profile, error) {
	var err error
	res := newProfile.Clone()
	newProfile.RangeCounters(func(pkgFile string, newCounters Counters) bool {
		var oldCounters Counters

		var isNewFile bool
		var contentUpdated bool
		var oldFile string
		isNewFile, contentUpdated, oldFile = opts.GetFileUpdateDetail(pkgFile)
		if isNewFile {
			res.SetCounters(pkgFile, newCounters)
			return true
		}
		if !contentUpdated {
			oldCounters = old.GetCounters(oldFile)
			var oldCountersLen int
			if oldCounters != nil {
				oldCountersLen = oldCounters.Len()
			}
			if !isNewFile && oldCountersLen == 0 {
				// if this is not a new file, and its content not updated, and it has zero counters,
				// then it is the case that `go list -deps ./src` does show the package, that is,
				// the package is excluded effectively from the build.
				res.SetCounters(pkgFile, newCounters)
				return true
			}
			if newCounters.Len() != oldCountersLen {
				err = fmt.Errorf("unchanged file found different lenght of counters: file=%s, old=%d, new=%d", pkgFile, oldCountersLen, newCounters.Len())
				return false
			}
			// plain merge
			addedCounters := newCounters.New(newCounters.Len())
			for i := 0; i < newCounters.Len(); i++ {
				addedCounters.Set(i, newCounters.Get(i).Add(oldCounters.Get(i)))
			}
			res.SetCounters(pkgFile, addedCounters)
			return true
		}

		// file updated, merge unchanged counters
		var oldMustExist bool // TODO: for debug
		oldCounters = old.GetCounters(oldFile)
		if oldCounters == nil {
			if oldMustExist {
				err = fmt.Errorf("counters not found for old file %s", oldFile)
				return false
			}
			res.SetCounters(pkgFile, newCounters)
			return true
		}
		var oldCode string
		var newCode string
		oldCode, err = oldCodeGetter(oldFile)
		if err != nil {
			return false
		}
		newCode, err = newCodeGetter(pkgFile)
		if err != nil {
			return false
		}
		var mergedCounter Counters
		mergedCounter, err = MergeFileNameCounters(oldCounters, oldFile, oldCode, newCounters, pkgFile, newCode)
		if err != nil {
			return false
		}
		res.SetCounters(pkgFile, mergedCounter)
		return true
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

// MergeFileCounter merge counters of the same file between two commits,
// the algorithm takes semantic update into consideration, making the
// merge more accurate while strict.
func MergeFileCounter(oldCounter Counters, oldCode string, newCounter Counters, newCode string) (mergedCounters Counters, err error) {
	return MergeFileNameCounters(oldCounter, "old.go", oldCode, newCounter, "new.go", newCode)
}

func MergeFileNameCounters(oldCounter Counters, oldFileName string, oldCode string, newCounter Counters, newFileName string, newCode string) (mergedCounters Counters, err error) {
	newToOld, err := ComputeFileBlockMapping(oldFileName, oldCode, newFileName, newCode)
	mergedCounters = newCounter.New(newCounter.Len())
	for i := 0; i < newCounter.Len(); i++ {
		c := newCounter.Get(i)
		if oldIdx, ok := newToOld[i]; ok {
			c = c.Add(oldCounter.Get(oldIdx))
		}
		mergedCounters.Set(i, c)
	}
	return
}

func GetModPath(dir string) (modPath string, err error) {
	// try to read mod from dir
	var mod struct {
		Module struct {
			Path string
		}
	}
	_, _, err = sh.RunBashCmdOpts(fmt.Sprintf(`cd %s && go mod edit -json`, sh.Quote(dir)), sh.RunBashOptions{
		StdoutToJSON: &mod,
	})
	if err != nil {
		err = fmt.Errorf("get module path: %v", err)
		return
	}
	modPath = mod.Module.Path
	return
}
