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
		fmt.Fprintln(os.Stderr, err)
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

// NOTE: even when your commit is already
// based on master, choosing 'Merge'
// will always generate a 'Merge Commit'.
//
// So we add an enforcement:
//
//	only one commit at a time, and the version must
//	match master+1.
//	With this limitation, we can always use rebase
//	to merge new commits without breaking the version file
func checkMerge(args []string) error {
	gitDir, err := git.ShowTopLevel("")
	if err != nil {
		return err
	}

	const targetBranch = "master"

	err = git.FetchRef(gitDir, "origin", targetBranch)
	if err != nil {
		return err
	}

	masterCommitHash, err := revision.GetCommitHash(gitDir, "origin/"+targetBranch)
	if err != nil {
		return err
	}

	branch, err := git.GetCurrentBranch(gitDir)
	if err != nil {
		return err
	}

	var found bool
	n := 20
	i := 0
	for ; i < n; i++ {
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
		return fmt.Errorf("%s is not based on newest %s, please rebase with %s", branch, targetBranch, targetBranch)
	}
	if i == 0 {
		return fmt.Errorf("%s is same with %s, nothing to merge", branch, targetBranch)
	}
	if i != 1 {
		return fmt.Errorf("%s has %d commits to be merged into %s, please squash them into 1(using 'git reset --soft origin/%s')", branch, i, targetBranch, targetBranch)
	}
	return nil
}
