package main

import (
	"os"
	"strings"

	"github.com/xhd2015/xgo/script/build-release/revision"
	"github.com/xhd2015/xgo/support/cmd"
)

// fixup src dir to prepare for release build
func fixupSrcDir(targetDir string, rev string) (restore func() error, err error) {
	restore, err = updateRevisions(targetDir, false, rev)
	if err != nil {
		return restore, err
	}
	return restore, nil
}

func stageFile(file string) (restore func() error, err error) {
	content, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return func() error {
		return os.WriteFile(file, content, 0755)
	}, nil
}

// NOTE: only commit is updated, version number not touched
func updateRevisions(targetDir string, unlink bool, rev string) (restore func() error, err error) {
	// unlink files because all files are symlink
	files := revision.GetVersionFiles(targetDir)
	var restoreFiles []func() error
	for _, file := range files {
		r, err := stageFile(file)
		if err != nil {
			return nil, err
		}
		restoreFiles = append(restoreFiles, r)
	}
	restore = func() error {
		for _, r := range restoreFiles {
			r()
		}
		return nil
	}
	if unlink {
		for _, file := range files {
			err := unlinkFile(file)
			if err != nil {
				return restore, err
			}
		}
	}

	for _, file := range files {
		err := revision.PatchVersionFile(file, "", rev, false, -1)
		if err != nil {
			return restore, err
		}
	}
	return restore, nil
}

func gitListWorkingTreeChangedFiles(dir string) ([]string, error) {
	// git ls-files:
	//   -c cached
	//   -d deleted
	//   -m modified
	//   -o untracked files
	//   --exclude-standard apply ignore rules
	//
	// example:
	//   all files in HEAD:  git ls-files --exclude-standard -c
	//   modified files:  git ls-files --exclude-standard -m
	//   untracked files:  git ls-files --exclude-standard -o
	output, err := cmd.Dir(dir).Output("git", "ls-files", "--exclude-standard", "-mo")
	if err != nil {
		return nil, err
	}
	return splitLinesFilterEmpty(output), nil
}

func splitLinesFilterEmpty(s string) []string {
	list := strings.Split(s, "\n")
	idx := 0
	for _, e := range list {
		if e != "" {
			list[idx] = e
			idx++
		}
	}
	return list[:idx]
}
