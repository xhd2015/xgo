package model

type BlockLine struct {
	Type string `json:"type"`
	// when two blocks intersect, the first block wins the overlapping line.
	Lines []int `json:"lines"`
}
