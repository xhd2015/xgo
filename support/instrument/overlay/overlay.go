package overlay

import (
	"os"
	"path/filepath"
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

func (o Overlay) Override(absFile AbsFile, targetFile AbsFile) {
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

func (o Overlay) MakeGoOverlay(overlayDir string) (*GoOverlay, error) {
	absOverlayDir, err := filepath.Abs(overlayDir)
	if err != nil {
		return nil, err
	}
	replace := make(Replace, len(o))
	for absFile, fo := range o {
		if fo.hasOverriddenContent {
			writeFile := filepath.Join(absOverlayDir, string(absFile))
			err := os.MkdirAll(filepath.Dir(writeFile), 0755)
			if err != nil {
				return nil, err
			}
			err = os.WriteFile(writeFile, []byte(fo.Content), 0644)
			if err != nil {
				return nil, err
			}
			replace[absFile] = AbsFile(writeFile)
			continue
		}
		if fo.AbsFile == "" {
			continue
		}
		replace[absFile] = fo.AbsFile
	}
	return &GoOverlay{Replace: replace}, nil
}
