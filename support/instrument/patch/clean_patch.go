package patch

import (
	"strings"
)

// CleanPatch removes all content between /*<begin and /*<end markers
func CleanPatch(content string) string {
	return CleanPatchMarkers(content, PatchMarker{Begin: "/*<begin", End: "*/"}, PatchMarker{Begin: "/*<end", End: "*/"})
}

type PatchMarker struct {
	Begin string
	End   string
}

func CleanPatchMarkers(content string, start PatchMarker, end PatchMarker) string {
	if start.Begin == "" || start.End == "" || end.Begin == "" || end.End == "" {
		panic("start or end marker is empty")
	}

	var result strings.Builder

	n := len(content)
	result.Grow(n)
	for i := 0; i < n; {
		startIdx, endIdx := findStartEnd(content, i, start, end)
		if startIdx == -1 {
			result.WriteString(content[i:])
			break
		}
		result.WriteString(content[i:startIdx])
		i = endIdx
	}
	return result.String()
}

func findStartEnd(content string, i int, start PatchMarker, end PatchMarker) (int, int) {
	startIdx, startEndIdx := findMarker(content, i, start)
	if startIdx == -1 {
		return -1, -1
	}

	endIdx, endEndIdx := findMarker(content, startEndIdx, end)
	if endIdx == -1 {
		return -1, -1
	}

	return startIdx, endEndIdx
}

func findMarker(content string, i int, marker PatchMarker) (int, int) {
	start := strings.Index(content[i:], marker.Begin)
	if start == -1 {
		return -1, -1
	}
	start += i
	i = start + len(marker.Begin)
	end := strings.Index(content[i:], marker.End)
	if end == -1 {
		return -1, -1
	}
	end += i + len(marker.End)
	return start, end
}
