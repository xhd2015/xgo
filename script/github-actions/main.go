package main

import (
	"fmt"
	"os"

	"github.com/xhd2015/xgo/script/build-release/revision"
	"github.com/xhd2015/xgo/support/git"
)

func main() {
	err := handle(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func handle(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("requires cmd")
	}
	cmd := args[0]
	args = args[1:]

	if cmd == "check-merge" {
		return checkMerge(args)
	} else {
		return fmt.Errorf("unrecognized command: %s", cmd)
	}
}

// check if merge to master
// will not result in merge commit
func checkMerge(args []string) error {
	gitDir, err := git.ShowTopLevel("")
	if err != nil {
		return err
	}

	err = git.FetchRef(gitDir, "origin", "master")
	if err != nil {
		return err
	}

	masterCommitHash, err := revision.GetCommitHash(gitDir, "origin/master")
	if err != nil {
		return err
	}

	var found bool
	n := 20
	for i := 0; i < n; i++ {
		ref := "HEAD"
		if i > 0 {
			ref = fmt.Sprintf("HEAD~%d", i)
		}
		commitHash, err := revision.GetCommitHash(gitDir, ref)
		if err != nil {
			return err
		}
		if commitHash == masterCommitHash {
			found = true
			break
		}
	}
	if !found {
		branch, err := git.GetCurrentBranch(gitDir)
		if err != nil {
			return err
		}
		return fmt.Errorf("%s cannot be merged into master without creating intermediate commit", branch)
	}
	return nil
}
