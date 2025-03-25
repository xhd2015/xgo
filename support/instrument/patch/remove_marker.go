package patch

import (
	"bytes"
	"fmt"
)

func RemoveBuildIgnore(content []byte) ([]byte, error) {
	const buildIgnore = "//go:build ignore"

	return RemoveMarker(content, buildIgnore)
}

func RemoveMarker(content []byte, marker string) ([]byte, error) {
	startIdx := bytes.Index(content, []byte(marker))
	if startIdx < 0 {
		return nil, fmt.Errorf("missing %s", marker)
	}
	idx := startIdx + len(marker)
	if idx < len(content) && content[idx] == '\r' {
		idx++
	}
	if idx < len(content) && content[idx] == '\n' {
		idx++
	}
	return append(content[:startIdx], content[idx:]...), nil
}
