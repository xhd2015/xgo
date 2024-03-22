package main

import (
	"log"
	"path/filepath"
	"testing"
)

func TestExampleGitListWorkingTree(t *testing.T) {
	gitDir, err := getGitDir()
	if err != nil {
		log.Fatal(err)
	}
	rootDir := filepath.Dir(gitDir)
	files, err := gitListWorkingTreeChangedFiles(rootDir)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("files: %v\n", files)
}
