package model

type Commit struct {
	Hash            string `json:"hash"`
	Msg             string `json:"msg"`
	AuthorName      string `json:"authorName"`
	AuthorEmail     string `json:"authorEmail"`
	AuthorTimestamp int64  `json:"authorTimestamp"`
	CommitTimestamp int64  `json:"commitTimestamp"`
	Tag             string `json:"tag"`

	// NotFound indicates the commit is missing
	NotFound bool `json:"notFound"`

	// optional
	FirstParent  string `json:"firstParent"`
	SecondParent string `json:"secondParent"`
}

type BlameCommit struct {
	*Commit
	Boundary bool `json:"boundary"`
}
