package model

type GrepLineOptions struct {
	// OnlyMatch  bool // implied
	IgnoreCase bool     `json:"ignoreCase"` // -i
	WordMatch  bool     `json:"wordMatch"`  // -w
	Posix      bool     `json:"posix"`      // -E
	Patterns   []string `json:"patterns"`
	Files      []string `json:"files"`
}
