package overlay

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/xgo/instrument/patch"
	"github.com/xhd2015/xgo/support/fileutil"
)

type AbsFile string

// serves as an in-memory FS
type Overlay map[AbsFile]*FileOverlay

type FileOverlay struct {
	AbsFile AbsFile
	Content string

	hasOverriddenContent bool
}

func MakeOverlay() Overlay {
	return make(Overlay)
}

func (o Overlay) OverrideFile(absFile AbsFile, targetFile AbsFile) {
	o[absFile] = &FileOverlay{
		AbsFile: targetFile,
	}
}

func (o Overlay) OverrideContent(absFile AbsFile, content string) {
	o[absFile] = &FileOverlay{
		Content:              content,
		hasOverriddenContent: true,
	}
}

func (o Overlay) Get(absFile AbsFile) *FileOverlay {
	return o[absFile]
}

func (o Overlay) Size(absFile AbsFile) (size int64, err error) {
	overlayFile := o.Get(absFile)

	readOSFile := absFile
	if overlayFile != nil {
		if overlayFile.hasOverriddenContent {
			return int64(len(overlayFile.Content)), nil
		}
		if overlayFile.AbsFile != "" {
			readOSFile = overlayFile.AbsFile
		}
	}
	info, err := os.Stat(string(readOSFile))
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

func (o Overlay) Read(absFile AbsFile) (hitContent bool, content string, err error) {
	overlayFile := o.Get(absFile)

	readOSFile := absFile
	if overlayFile != nil {
		if overlayFile.hasOverriddenContent {
			return true, overlayFile.Content, nil
		}
		if overlayFile.AbsFile != "" {
			readOSFile = overlayFile.AbsFile
		}
	}
	data, err := os.ReadFile(string(readOSFile))
	if err != nil {
		return false, "", err
	}
	return false, string(data), nil
}

func (o Overlay) ReadBytes(absFile AbsFile) (hitContent bool, content []byte, err error) {
	overlayFile := o.Get(absFile)

	readOSFile := absFile
	if overlayFile != nil {
		if overlayFile.hasOverriddenContent {
			return true, []byte(overlayFile.Content), nil
		}
		if overlayFile.AbsFile != "" {
			readOSFile = overlayFile.AbsFile
		}
	}
	data, err := os.ReadFile(string(readOSFile))
	if err != nil {
		return false, nil, err
	}
	return false, data, nil
}

type Options struct {
	NoLineDirective bool
	PathMappings    []PathMapping
}

// PROJECT
// GOROOT
// GOPATH
type PathMapping struct {
	From string
	To   string
}

func getPathMapping(path string, mappings []PathMapping) string {
	for _, mapping := range mappings {
		// normalize to forward slashes for cross-platform matching
		normPath := filepath.ToSlash(path)
		normFrom := filepath.ToSlash(mapping.From)
		if !strings.HasPrefix(normPath, normFrom) {
			continue
		}
		if len(normPath) == len(normFrom) {
			return mapping.To
		}
		if normPath[len(normFrom)] != '/' {
			continue
		}
		return mapping.To + normPath[len(normFrom):]
	}
	return path
}

func (o Overlay) MakeGoOverlay(overlayDir string, opts Options) (*GoOverlay, error) {
	noLineDirective := opts.NoLineDirective
	pathMappings := opts.PathMappings

	absOverlayDir, err := filepath.Abs(overlayDir)
	if err != nil {
		return nil, err
	}
	replace := make(Replace, len(o))
	for absFile, fileOverlay := range o {
		if fileOverlay.hasOverriddenContent {
			absFilePath := getPathMapping(string(absFile), pathMappings)
			writeFile := fileutil.RebaseAbsPath(absOverlayDir, absFilePath)
			err := os.MkdirAll(filepath.Dir(writeFile), 0755)
			if err != nil {
				return nil, err
			}
			content := fileOverlay.Content
			if !noLineDirective {
				content = patch.FmtLineDirective(string(absFile), 1) + "\n" + content
			}
			err = os.WriteFile(writeFile, []byte(content), 0644)
			if err != nil {
				return nil, err
			}
			replace[AbsFile(filepath.ToSlash(string(absFile)))] = AbsFile(filepath.ToSlash(writeFile))
			continue
		}
		if fileOverlay.AbsFile == "" {
			continue
		}
		replace[AbsFile(filepath.ToSlash(string(absFile)))] = AbsFile(filepath.ToSlash(string(fileOverlay.AbsFile)))
	}
	return &GoOverlay{Replace: replace}, nil
}
