package model

type LineNum int64

type LineAnnotation struct {
	OldLine int64   `json:"oldLine,omitempty"` // if 0 or undefined, no oldLine. only effective when not changed
	Changed bool    `json:"changed,omitempty"` // is this line changed
	BlockID BlockID `json:"blockID,omitempty"` // related blockID, can not be zero
	Empty   bool    `json:"empty,omitempty"`   // is this an empty line

	// Labels tells what code sets this line matches
	// a map to make merge easier
	Labels StringSet `json:"labels,omitempty"`

	// ExecLabels is the set of labels collected by runtime execution
	ExecLabels StringSet `json:"execLabels,omitempty"`

	// CoverageLabels true->has cover, false->has no cover
	// examples: ALL:true, RC:true|false
	// NOTE: it's not a StringSet, its just a map derived from Labels & ExecLabels and other options
	CoverageLabels map[string]bool `json:"coverageLabels,omitempty"`

	FuncID FuncID `json:"funcID,omitempty"`

	// Uncoverable empty or non-block or just a simple "}", or it has been marked unreachable
	Uncoverable bool `json:"uncoverable,omitempty"`

	Code *CodeAnnotation `json:"code,omitempty"`

	Remark *Remark `json:"remark,omitempty"`

	LineDataType string      `json:"lineDataType,omitempty"`
	LineData     interface{} `json:"lineData,omitempty"` // extra line data
}
