package overlay

import (
	"encoding/json"
	"os"
)

type Replace map[AbsFile]AbsFile

type GoOverlay struct {
	Replace Replace
}

func (c *GoOverlay) Write(file string) error {
	overlayData, err := json.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(file, overlayData, 0755)
}

func ReadGoOverlay(file string) (*GoOverlay, error) {
	content, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var overlay GoOverlay
	err = json.Unmarshal(content, &overlay)
	if err != nil {
		return nil, err
	}
	return &overlay, nil
}
