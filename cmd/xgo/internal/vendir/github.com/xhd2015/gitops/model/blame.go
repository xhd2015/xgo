package model

type BlameInfo struct {
	Line       int64  `json:"line"`
	CommitHash string `json:"commitHash"`
}

type PlainBlameInfo struct {
	Line        int64  `json:"line"`
	CommitHash  string `json:"commitHash"` // the commit hash is short
	AuthorEmail string `json:"authorEmail"`
	Boundary    bool   `json:"bounary,omitempty"` // blame output starts with ^
	Timestamp   int64  `json:"timestamp"`
}
