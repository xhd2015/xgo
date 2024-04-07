package main

import (
	"fmt"

	"github.com/xhd2015/xgo/support/goinfo"
)

type FilePatch struct {
	FilePath _FilePath
	Patches  []*Patch
}

type Patch struct {
	Mark string

	InsertIndex  int // insert before which anchor, 0: insert at head, -1: tail
	InsertBefore bool
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
	file := c.FilePath.Join(goroot)

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

	return editFile(file, func(content string) (string, error) {
		for _, patch := range c.Patches {
			if patch.CheckGoVersion != nil && !patch.CheckGoVersion(goVersion) {
				continue
			}
			beginMark := fmt.Sprintf("/*<begin %s>*/", patch.Mark)
			endMark := fmt.Sprintf("/*<end %s>*/", patch.Mark)
			content = addContentAtIndex(content, beginMark, endMark, patch.Anchors, patch.InsertIndex, patch.InsertBefore, patch.Content)
		}
		return content, nil
	})
}
