package model

import (
	"github.com/xhd2015/go-coverage/git"
	"github.com/xhd2015/go-coverage/model"
)

type RelativeFile string

type FileAnnotation struct {
	ChangeDetail *git.FileDetail        `json:"changeDetail,omitempty"`
	Lines        LineAnnotationMapping  `json:"lines,omitempty"`       // 1-based mapping
	LineChanges  *model.LineChanges     `json:"lineChanges,omitempty"` // optional
	DeletedLines map[int64]bool         `json:"deletedLines,omitempty"`
	Blocks       BlockAnnotationMapping `json:"blocks,omitempty"`
	Funcs        FuncAnnotationMapping  `json:"funcs,omitempty"` // by func id
	FileDataType string                 `json:"fileDataType,omitempty"`
	FileData     interface{}            `json:"fileData,omitempty"` // extra file data
}
