package model

type FileDetail struct {
	// if none of following fields set, this file is unchanged.
	// Unchanged      bool   `json:",omitempty"`
	IsNew          bool   `json:",omitempty"`
	Deleted        bool   `json:",omitempty"`
	RenamedFrom    string `json:",omitempty"` // empty if not renamed
	ContentChanged bool   `json:",omitempty"` // content change type
}

func (c *FileDetail) Unchanged() bool {
	return !c.IsNew && !c.Deleted && c.RenamedFrom == "" && !c.ContentChanged
}
