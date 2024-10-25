package model

import (
	"fmt"
)

type FileAnnotationMapping map[RelativeFile]*FileAnnotation
type LineAnnotationMapping map[LineNum]*LineAnnotation
type BlockAnnotationMapping map[BlockID]*BlockAnnotation
type FuncAnnotationMapping map[FuncID]*FuncAnnotation

type ProjectAnnotation struct {
	// short file -> annotations
	Files           FileAnnotationMapping   `json:"files,omitempty"`
	Types           map[AnnotationType]bool `json:"types,omitempty"` // types indicator
	CommitHash      string                  `json:"commitHash,omitempty"`
	ProjectDataType string                  `json:"projectDataType,omitempty"`
	ProjectData     interface{}             `json:"projectData,omitempty"` // extra project data
}

type AnnotationType string

const (
	AnnotationType_Blocks       AnnotationType = "blocks"        // the File.Blocks
	AnnotationType_ChangeDetail AnnotationType = "change_detail" // the File.ChangeDetail

	// lines
	AnnotationType_LineEmpty          AnnotationType = "line_empty"           // the Line.Empty
	AnnotationType_LineBlockID        AnnotationType = "line_block_id"        // the Line.BlockID
	AnnotationType_LineFuncID         AnnotationType = "line_func_id"         // the Line.BlockID
	AnnotationType_LineChanges        AnnotationType = "line_changes"         // the Line.LineChanges
	AnnotationType_LineChanged        AnnotationType = "line_changed"         // the Line.Changed
	AnnotationType_LineOldLine        AnnotationType = "line_old_line"        // the Line.OldLine
	AnnotationType_LineLabels         AnnotationType = "line_labels"          // the Line.Labels
	AnnotationType_LineExecLabels     AnnotationType = "line_exec_labels"     // the Line.ExecLabels
	AnnotationType_LineCoverageLabels AnnotationType = "line_coverage_labels" // the Line.CoverageLabels
	AnnotationType_LineUncoverable    AnnotationType = "line_uncoverable"     // the Line.Uncoverable
	AnnotationType_LineCodeExcluded   AnnotationType = "line_code_excluded"   // the Line.Code.Excluded
	AnnotationType_LineRemark         AnnotationType = "line_remark"          // the Line.Remark

	// funcs
	AnnotationType_FileFuncs          AnnotationType = "funcs"                // the File.Funcs
	AnnotationType_FuncLabels         AnnotationType = "func_labels"          // the Line.ExecLabels
	AnnotationType_FuncExecLabels     AnnotationType = "func_exec_labels"     // the Line.ExecLabels
	AnnotationType_FuncCoverageLabels AnnotationType = "func_coverage_labels" // the Func.CoverageLabels
	AnnotationType_FuncChanged        AnnotationType = "func_changed"         // the Func.Changed
	AnnotationType_FuncCodeComments   AnnotationType = "func_code_comments"   // the Func.Code.Comments

	// Func.Code.Excluded is the exclude-flag by comment, there are some other
	// flags that can indicate whether a func is excluded:
	// func excluded = headLineExcluded || allLineExcluded || excludedByComment
	AnnotationType_FuncCodeExcluded AnnotationType = "func_code_excluded" // the Func.Code.Excluded

	// first line excluded
	AnnotationType_FirstLineExcluded AnnotationType = "func_first_line_excluded"
)

func (c *ProjectAnnotation) Has(annotationType AnnotationType) bool {
	return c.Types[annotationType]
}

func (c *ProjectAnnotation) MustHave(reason string, annotationTypes ...AnnotationType) {
	if err := c.ShouldHave(reason, annotationTypes...); err != nil {
		panic(err)
	}
}

func (c *ProjectAnnotation) ShouldHave(reason string, annotationTypes ...AnnotationType) error {
	for _, annType := range annotationTypes {
		if !c.Types[annType] {
			return fmt.Errorf("%s requires %s", reason, annType)
		}
	}
	return nil
}

func (c *ProjectAnnotation) Set(annotationType AnnotationType) {
	if c.Types == nil {
		c.Types = make(map[AnnotationType]bool)
	}
	c.Types[annotationType] = true
}

// clone impl
var MergeAnnotationsInto func(res *ProjectAnnotation, annotations ...*ProjectAnnotation)

func (c *ProjectAnnotation) Clone() *ProjectAnnotation {
	b := &ProjectAnnotation{
		CommitHash:      c.CommitHash,
		ProjectDataType: c.ProjectDataType,
		ProjectData:     c.ProjectData,
	}
	MergeAnnotationsInto(b, c)
	return b
}
func (c *ProjectAnnotation) Simplified() *ProjectAnnotation {
	c.Simplify()
	return c
}

func (c *ProjectAnnotation) Simplify() {
	if c == nil {
		return
	}
	if len(c.Types) == 0 {
		c.Types = nil
	}
	if len(c.Files) == 0 {
		c.Files = nil
	} else {
		for _, an := range c.Files {
			an.Simplify()
		}
	}
}

func (c *FileAnnotation) Simplify() {
	if c == nil {
		return
	}

	if len(c.Lines) == 0 {
		c.Lines = nil
	} else {
		for _, line := range c.Lines {
			line.Simplify()
		}
	}

	if len(c.Blocks) == 0 {
		c.Blocks = nil
	} else {
		for _, block := range c.Blocks {
			block.Simplify()
		}
	}

	if len(c.Funcs) == 0 {
		c.Funcs = nil
	} else {
		for _, fn := range c.Funcs {
			fn.Simplify()
		}
	}
}

func (c *LineAnnotation) Simplify() {
	if c == nil {
		return
	}
	if len(c.Labels) == 0 {
		c.Labels = nil
	}
	if len(c.ExecLabels) == 0 {
		c.ExecLabels = nil
	}
	if len(c.CoverageLabels) == 0 {
		c.CoverageLabels = nil
	}
}

func (c *FuncAnnotation) Simplify() {
	if c == nil {
		return
	}
	if len(c.Labels) == 0 {
		c.Labels = nil
	}
	if len(c.ExecLabels) == 0 {
		c.ExecLabels = nil
	}
	if len(c.CoverageLabels) == 0 {
		c.CoverageLabels = nil
	}
}
