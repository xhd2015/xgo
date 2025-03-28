package patch

import (
	"fmt"
	"strings"
)

func RemoveBuildIgnore(content string) (string, error) {
	const buildIgnore = "//go:build ignore"

	return RemoveMarker(content, buildIgnore)
}

func RemoveMarker(content string, marker string) (string, error) {
	startIdx := strings.Index(content, marker)
	if startIdx < 0 {
		return "", fmt.Errorf("missing %s", marker)
	}
	idx := startIdx + len(marker)
	if idx < len(content) && content[idx] == '\r' {
		idx++
	}
	if idx < len(content) && content[idx] == '\n' {
		idx++
	}
	return content[:startIdx] + content[idx:], nil
}
