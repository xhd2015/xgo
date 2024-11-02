package model

type MergePoint struct {
	CommitHash            string `json:"commitHash"`
	FirstParentCommitHash string `json:"firstParentCommitHash"` // useful to get a diff with
	// SecondParentCommitHash string `json:"secondParentCommitHash"` //
}

type MergeInfo struct {
	UnmergedCommit *Commit   `json:"unmergedCommit"`
	MergePoints    []*Commit `json:"mergePoints"`
}
