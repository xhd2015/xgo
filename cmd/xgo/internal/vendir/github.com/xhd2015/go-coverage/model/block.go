package model

// Block can be used as a unique key across one file
type Block struct {
	StartLine int `json:"startLine"`
	StartCol  int `json:"startCol"`
	EndLine   int `json:"endLine"`
	EndCol    int `json:"endCol"`
}

// BlockProfile a general way to organize block
type BlockProfile map[PkgFile][]*BlockData

type ShortFilename = string // xxx.go
type PkgPath = string
type PkgFile = string // <PkgPath>/<ShortFilename>

// BlockData provide a general way to process block stats.
type BlockData struct {
	Block *Block
	Data  interface{}
}
