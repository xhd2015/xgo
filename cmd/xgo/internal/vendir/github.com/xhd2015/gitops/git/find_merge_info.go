package git

import (
	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/gitops/model"
)

func FindMergeInfo(dir string, ref string, base string) (*model.MergeInfo, error) {
	refCommit, err := RevParseVerified(dir, ref)
	if err != nil {
		return nil, err
	}
	baseCommit, err := RevParseVerified(dir, base)
	if err != nil {
		return nil, err
	}

	mergeCommits, _, err := doFindMergePoints(dir, refCommit, baseCommit, ref, base, false)
	if err != nil {
		return nil, err
	}
	var unmergedCommit *model.Commit
	var mergePoints []*model.Commit
	if len(mergeCommits) > 0 {
		commitsInfo, err := GetCommits(dir, mergeCommits)
		if err != nil {
			return nil, err
		}
		mergePoints = make([]*model.Commit, 0, len(mergeCommits))
		for _, mergeCommit := range mergeCommits {
			commitInfo := commitsInfo[mergeCommit]
			if commitInfo != nil {
				mergePoints = append(mergePoints, commitInfo)
			}
		}
	} else {
		// if len(mergeCommits) == 0, must be unmerged
		//   if ref is just the same as base, then ref is created without any modification
		if refCommit != baseCommit {
			refContainsBase, err := Contains(dir, refCommit, baseCommit)
			if err != nil {
				return nil, err
			}
			if !refContainsBase {
				commitInfo, err := GetCommit(dir, base)
				if err != nil {
					return nil, err
				}
				unmergedCommit = commitInfo
			}
		}
	}
	return &model.MergeInfo{
		UnmergedCommit: unmergedCommit,
		MergePoints:    mergePoints,
	}, nil
}

// req:
// {
//     appName
//     branch
//     base
// }

// resp:
// {
//     unmergedInfo:{
//           "commitHash":"xx",
//            "authorTime":"2023-05-30 18:02:00"
//     },
//     mergePoints:[{
//         commitHash:"",
//         "authorTime":"2023-05-30 18:02:00"

//     }]
// }
