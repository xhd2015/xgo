package main

import (
	"fmt"

	"github.com/xhd2015/xgo/support/goinfo"
	instrument_patch "github.com/xhd2015/xgo/support/instrument/patch"
)

type FilePatch struct {
	FilePath instrument_patch.FilePath
	Patches  []*Patch
}

type Patch struct {
	Mark string

	InsertIndex    int // insert before which anchor, 0: insert at head, -1: tail
	UpdatePosition instrument_patch.UpdatePosition
	// anchor should be unique
	// appears exactly once
	Anchors []string

	Content string

	CheckGoVersion func(goVersion *goinfo.GoVersion) bool
}

func (c *FilePatch) Apply(goroot string, goVersion *goinfo.GoVersion) error {
	if goroot == "" {
		return fmt.Errorf("requires goroot")
	}
	if len(c.FilePath) == 0 {
		return fmt.Errorf("invalid file path")
	}

	if len(c.Patches) == 0 {
		return nil
	}
	file := c.FilePath.JoinPrefix(goroot)

	// validate patch mark
	seenMark := make(map[string]bool, len(c.Patches))
	for i, patch := range c.Patches {
		if patch.Content == "" {
			return fmt.Errorf("empty content at: %d", i)
		}
		if patch.Mark == "" {
			return fmt.Errorf("empty mark at %d", i)
		}
		if len(patch.Anchors) == 0 {
			return fmt.Errorf("empty anchors at: %d", i)
		}
		if _, ok := seenMark[patch.Mark]; ok {
			return fmt.Errorf("duplicate mark: %s", patch.Mark)
		}
		seenMark[patch.Mark] = true
	}

	return instrument_patch.EditFile(file, func(content string) (string, error) {
		for _, patch := range c.Patches {
			if patch.CheckGoVersion != nil && !patch.CheckGoVersion(goVersion) {
				continue
			}
			beginMark := fmt.Sprintf("/*<begin %s>*/", patch.Mark)
			endMark := fmt.Sprintf("/*<end %s>*/", patch.Mark)
			content = instrument_patch.UpdateContentLines(content, beginMark, endMark, patch.Anchors, patch.InsertIndex, patch.UpdatePosition, patch.Content)
		}
		return content, nil
	})
}
