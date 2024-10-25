package model

type CommentLabel string

const (
	CommentLabels_Labels      CommentLabel = "labels"
	CommentLabels_Unreachable CommentLabel = "unreachable"
	CommentLabels_NoCov       CommentLabel = "nocov"
	CommentLabels_Deprecated  CommentLabel = "deprecated"
)

type CodeAnnotation struct {
	// Labels     StringSet           `json:"labels,omitempty"`
	// ExecLabels StringSet           `json:"execLabels,omitempty"`
	Comments map[CommentLabel]*CodeComment `json:"comments,omitempty"` // example:  {labels: ["a","b","c"], "unreachable":true}

	Excluded bool `json:"excluded,omitempty"`
}

type CodeComment struct {
	Author string   `json:"author,omitempty"`
	Values []string `json:"values,omitempty"`
}

func (c *CodeComment) Clone() *CodeComment {
	if c == nil {
		return nil
	}
	return &CodeComment{
		Author: c.Author,
		Values: cloneSlice(c.Values),
	}
}

func cloneSlice(vals []string) []string {
	if vals == nil {
		return nil
	}
	copied := make([]string, 0, len(vals))
	return append(copied, vals...)
}
