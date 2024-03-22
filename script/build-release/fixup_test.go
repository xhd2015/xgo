package main

import (
	"fmt"
	"log"
	"path/filepath"
)

func ExampleGitListWorkingTreeChangedFiles() {
	gitDir, err := getGitDir()
	if err != nil {
		log.Fatal(err)
	}
	rootDir := filepath.Dir(gitDir)
	files, err := gitListWorkingTreeChangedFiles(rootDir)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("files: %v\n", files)
	// Output [ignored]:
	//   files: [script/build-release/fixup_test.go]
}
