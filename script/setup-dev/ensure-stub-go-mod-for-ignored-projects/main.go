package main

import (
	"fmt"
	"os"
	"path/filepath"
)

var stubs = []string{
	"go-release-git-versioned",
	"go-release",
	"go-patch-diff",
	"tmp",
}

var stubContent = []byte(`// stub file to ensure it is not scanned by ` + "`go list`" + "\n")

func main() {
	for _, dir := range stubs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}
		modFile := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(modFile); err == nil {
			continue
		}
		if err := os.WriteFile(modFile, stubContent, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "ensure stub go.mod for %s: %v\n", dir, err)
			os.Exit(1)
		}
		fmt.Printf("created stub go.mod in %s\n", dir)
	}
}
