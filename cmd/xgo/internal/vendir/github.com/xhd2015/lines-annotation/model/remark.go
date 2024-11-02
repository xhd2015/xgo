package model

type ExcludeReason string
type CodeSuggestion string

const (
	ExcludeReason_Other ExcludeReason = "other"
)

const (
	Suggestion_Other CodeSuggestion = "other"
)

type Remark struct {
	Excluded   bool           `json:"excluded,omitempty"`
	Reason     ExcludeReason  `json:"reason,omitempty"`
	Suggestion CodeSuggestion `json:"suggestion,omitempty"`
	Comments   []*Comment     `json:"comments,omitempty"`
	CreateTime string         `json:"createTime,omitempty"`
	Creator    string         `json:"creator,omitempty"`
	UpdateTime string         `json:"updateTime,omitempty"`
	Updater    string         `json:"updater,omitempty"`
}

type Comment struct {
	Author string `json:"author,omitempty"`
	// UTC time
	CreateTime string `json:"createTime,omitempty"`
	Content    string `json:"content,omitempty"`
}

type RemapRequest struct {
	Include       []string           `json:"include"`
	Exclude       []string           `json:"exclude"`
	GitURL        string             `json:"gitURL"`
	CommitHash    string             `json:"commitHash"`
	OldCommitHash string             `json:"oldCommitHash"`
	Annotation    *ProjectAnnotation `json:"annotation"`
}
type RemapResponse struct {
	Annotation *ProjectAnnotation `json:"annotation"`
}
