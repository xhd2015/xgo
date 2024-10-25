package model

type StringSet = map[string]bool

// FuncID and BlockID are globally unique within a file and a commit
type FuncID = BlockID

type FuncAnnotation struct {
	Block         *Block `json:"block,omitempty"`
	Name          string `json:"name,omitempty"`
	OwnerTypeName string `json:"ownerTypeName,omitempty"`
	Closure       bool   `json:"closure"` // a closure, without name
	Changed       bool   `json:"changed"` // derived from line changed
	Ptr           bool   `json:"ptr,omitempty"`

	// Labels inerited from lines
	Labels StringSet `json:"labels,omitempty"`

	// ExecLabels inherited from lines
	ExecLabels StringSet `json:"execLabels,omitempty"`

	CoverageLabels map[string]bool `json:"coverageLabels,omitempty"`
	Code           *CodeAnnotation `json:"code,omitempty"`

	FirstLineExcluded bool `json:"firstLineExcluded,omitempty"`
}
