package overlay

import (
	"os"
	"path/filepath"

	"github.com/xhd2015/xgo/support/instrument/patch"
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
	fo := o.Get(absFile)

	readOSFile := absFile
	if fo != nil {
		if fo.hasOverriddenContent {
			return int64(len(fo.Content)), nil
		}
		if fo.AbsFile != "" {
			readOSFile = fo.AbsFile
		}
	}
	info, err := os.Stat(string(readOSFile))
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

func (o Overlay) Read(absFile AbsFile) (hitContent bool, content string, err error) {
	fo := o.Get(absFile)

	readOSFile := absFile
	if fo != nil {
		if fo.hasOverriddenContent {
			return true, fo.Content, nil
		}
		if fo.AbsFile != "" {
			readOSFile = fo.AbsFile
		}
	}
	data, err := os.ReadFile(string(readOSFile))
	if err != nil {
		return false, "", err
	}
	return false, string(data), nil
}

func (o Overlay) MakeGoOverlay(overlayDir string, addLineDirective bool) (*GoOverlay, error) {
	absOverlayDir, err := filepath.Abs(overlayDir)
	if err != nil {
		return nil, err
	}
	replace := make(Replace, len(o))
	for absFile, fileOverlay := range o {
		if fileOverlay.hasOverriddenContent {
			writeFile := filepath.Join(absOverlayDir, string(absFile))
			err := os.MkdirAll(filepath.Dir(writeFile), 0755)
			if err != nil {
				return nil, err
			}
			content := fileOverlay.Content
			if addLineDirective {
				content = patch.FmtLineDirective(string(absFile), 1) + "\n" + content
			}
			err = os.WriteFile(writeFile, []byte(content), 0644)
			if err != nil {
				return nil, err
			}
			replace[absFile] = AbsFile(writeFile)
			continue
		}
		if fileOverlay.AbsFile == "" {
			continue
		}
		replace[absFile] = fileOverlay.AbsFile
	}
	return &GoOverlay{Replace: replace}, nil
}
