package git

import (
	"fmt"

	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/go-coverage/sh"

	git "github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/gitops/model"
)

// method 1:
// 1. find merge points of master
//    M=git rev-list --merges --first-parent master
// 2. find commits of dev   (NOTE: dev must have a start point from master)
//    D=git rev-list --first-parent dev
// 3.find first common commit of M and D,
//    X = (M^D)[0]
//  X is M's start point
// 4. find all commits in M after X(not inclusive), check each commit if it is reachable from dev
//
// tech details:
//    git merge-base --all
//

// method 2:
// 1. find all commits uniquely belonging to dev
//    D=git rev-list --first-parent --exclude-first-parent-only master..dev
//    let x=last commit of D, which is the first commit of dev, thus y=x^1, which is the start point
//    validate that D must be non-empty, and y must exists,and belongs to master
// 2.find all merge points of master
//    M=git rev-list --first-parent --merges y..master
// 3.for each commit m in M, check if m^2 is in D,if so, add m to results

// FindMergePoints `ref` is the target branch, `base` is the branch where `ref` comes from, usually 'master'
func FindMergePoints(dir string, ref string, base string) ([]*git.MergePoint, error) {
	mergeCommits, _, err := findMergePoints(dir, ref, base)
	if err != nil {
		return nil, err
	}
	// batch resolve
	var mergePoints []*git.MergePoint
	for _, commit := range mergeCommits {
		parentCommit, err := RevParseVerified(dir, commit+"^1")
		if err != nil {
			return nil, err
		}

		mergePoints = append(mergePoints, &git.MergePoint{
			CommitHash:            commit,
			FirstParentCommitHash: parentCommit,
		})
	}
	return mergePoints, nil
}

func FindDiffPoints(dir string, ref string, base string, mergedAnchorBase string) (exists bool, merged bool, refCommit string, baseCommit string, actualMergedAnchorBase string, refCommits []string, err error) {
	refCommit, err = RevParseVerified(dir, ref)
	if err != nil {
		return
	}
	baseCommit, err = RevParseVerified(dir, base)
	if err != nil {
		return
	}

	mergeCommits, refCommits, err := doFindMergePoints(dir, refCommit, baseCommit, ref, base, false)
	if err != nil {
		return
	}
	exists = true
	merged = len(mergeCommits) > 0
	if merged {
		actualMergedAnchorBase = mergedAnchorBase
		if actualMergedAnchorBase != "" {
			// TODO: warn the given anchor base does not exist any more
			found := false
			for _, mergeCommit := range mergeCommits {
				if mergeCommit == mergedAnchorBase {
					found = true
					break
				}
			}
			if !found {
				actualMergedAnchorBase = ""
			}
		}

		if actualMergedAnchorBase == "" {
			// choose the first one by default
			actualMergedAnchorBase = mergeCommits[len(mergeCommits)-1]
		}

		baseCommit, err = RevParseVerified(dir, actualMergedAnchorBase+"^1")
		if err != nil {
			return
		}
	}

	return
}

func FindBranchCommits(dir string, ref string, base string) ([]string, error) {
	return runCommdSplitLines(dir,
		fmt.Sprintf("git rev-list --first-parent --exclude-first-parent-only %s..%s", sh.Quote(base), sh.Quote(ref)),
	)
}

func findMergePoints(dir string, ref string, base string) (mergeCommits []string, refCommits []string, err error) {
	return doFindMergePoints(dir, ref, base, ref, base, true)
}

func doFindMergePoints(dir string, ref string, base string, refName string, baseName string, needParseRef bool) (mergeCommits []string, refCommits []string, err error) {
	if ref == "" || base == "" {
		err = fmt.Errorf("requires ref,base")
		return
	}
	refHeadCommit := ref
	if needParseRef {
		refHeadCommit, err = RevParseVerified(dir, ref)
		if err != nil {
			return
		}
	}

	baseHeadCommit := base
	if needParseRef {
		baseHeadCommit, err = RevParseVerified(dir, base)
		if err != nil {
			return
		}
	}

	refCommits, err = FindBranchCommits(dir, refHeadCommit, baseHeadCommit)
	if err != nil {
		return
	}
	// validate
	if len(refCommits) == 0 {
		// they are just the same commit
		if refHeadCommit == baseHeadCommit {
			return
		}
		var baseHasNewerCommit bool
		// or baseHeadCommit is newer than refHeadCommit, which means
		// branch has no commits yet
		baseHasNewerCommit, err = IsAncesorOf(dir, refHeadCommit, baseHeadCommit)
		if err != nil {
			return
		}
		if baseHasNewerCommit {
			return
		}
		err = fmt.Errorf("%s has no commits compared with %s", trimRef(refName), trimRef(base))
		return
	}
	firstRefCommit := refCommits[len(refCommits)-1]
	var firstRefCommitParent string
	firstRefCommitParent, err = revParseVerified(dir, firstRefCommit+"^1")
	if err != nil {
		if err == ErrNotExists {
			// see https://github.com/xhd2015/gitops/issues/3
			// this error means ref has no intersection with base,
			// which means merge commit list is empty
			err = nil
			// TODO: return a dedicated error
			return
		}
		return
	}
	var firstIsAncestor bool
	firstIsAncestor, err = IsAncesorOf(dir, firstRefCommitParent, baseHeadCommit)
	if err != nil {
		return
	}
	if !firstIsAncestor {
		// TODO: return a dedicated error
		err = fmt.Errorf("%s seems does not branch from %s, the starting point is %s", trimRef(refName), trimRef(base), firstRefCommitParent)
		return
	}

	var baseCommits []string
	baseCommits, err = runCommdSplitLines(dir,
		fmt.Sprintf("git rev-list --first-parent --merges %s..%s", sh.Quote(firstRefCommitParent), sh.Quote(baseHeadCommit)),
	)
	if err != nil {
		return
	}

	refCommitsMapping := make(map[string]bool, len(refCommits))
	for _, refCommit := range refCommits {
		refCommitsMapping[refCommit] = true
	}

	for _, baseCommit := range baseCommits {
		var baseSecParent string
		baseSecParent, err = RevParseVerified(dir, baseCommit+"^2")
		if err != nil {
			return
		}
		if !refCommitsMapping[baseSecParent] {
			continue
		}
		mergeCommits = append(mergeCommits, baseCommit)
	}
	return
}

func runCommdSplitLines(dir string, cmds ...string) ([]string, error) {
	res, err := RunCommands(dir, cmds...)
	if err != nil {
		return nil, err
	}
	return splitLinesFilterEmpty(res), nil
}
