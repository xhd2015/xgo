package model

type DiffCommitOptions struct {
	PathPatterns []string `json:"pathPatterns"` // only include these patterns, example: src/**/*.go
}
