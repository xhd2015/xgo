package main

import (
	"strings"

	"github.com/xhd2015/xgo/script/build-release/revision"
	"github.com/xhd2015/xgo/support/cmd"
)

// fixup src dir to prepare for release build
func fixupSrcDir(targetDir string, rev string) error {
	err := updateRevisions(targetDir, false, rev)
	if err != nil {
		return err
	}
	return nil
}

func updateRevisions(targetDir string, unlink bool, rev string) error {
	// unlink files because all files are symlink
	files := revision.GetVersionFiles(targetDir)
	if unlink {
		for _, file := range files {
			err := unlinkFile(file)
			if err != nil {
				return err
			}
		}
	}

	for _, file := range files {
		err := revision.PatchVersionFile(file, rev, false)
		if err != nil {
			return err
		}
	}
	return nil
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
