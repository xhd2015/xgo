package model

import "github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/go-coverage/diff/vscode"

type FileDetail struct {
	// if none of following fields set, this file is unchanged.
	// Unchanged      bool   `json:",omitempty"`
	IsNew          bool   `json:",omitempty"`
	Deleted        bool   `json:",omitempty"`
	RenamedFrom    string `json:",omitempty"` // empty if not renamed
	ContentChanged bool   `json:",omitempty"` // content change type
}
type FileChanges map[string]*LineChanges

type LineChanges struct {
	OldLineCount int64                `json:"oldLineCount"`
	NewLineCount int64                `json:"newLineCount"`
	Changes      []*vscode.LineChange `json:"changes"`
}
