package model

type BlockAnnotation struct {
	Block  *Block `json:"block,omitempty"`
	FuncID FuncID `json:"funcID,omitempty"`
	// ExecLabels is the set of labels collected by runtime execution
	ExecLabels    StringSet   `json:"execLabels,omitempty"`
	BlockDataType string      `json:"blockDataType,omitempty"`
	BlockData     interface{} `json:"blockData,omitempty"` // extra block data
}

func (c *BlockAnnotation) Simplify() {
}
