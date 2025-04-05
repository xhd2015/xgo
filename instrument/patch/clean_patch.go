package patch

import (
	"strings"
)

// CleanPatch removes all content between /*<begin and /*<end markers
func CleanPatch(content string) string {
	return CleanPatchMarkers(content, PatchMarker{Begin: "/*<begin", End: ">*/"}, PatchMarker{Begin: "/*<end", End: ">*/"})
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
		beginIdx, beginEndIdx, endIdx, endEndIdx := findStartEnd(content, i, start, end)
		if beginIdx == -1 {
			result.WriteString(content[i:])
			break
		}
		result.WriteString(content[i:beginIdx])
		oldContent := extractOldContent(content[beginEndIdx:endIdx])
		if oldContent != "" {
			result.WriteString(oldContent)
		}
		i = endEndIdx
	}
	return result.String()
}

func extractOldContent(content string) string {
	idxOld := strings.Index(content, "/*old:")
	if idxOld == -1 {
		return ""
	}
	base := idxOld + len("/*old:")
	idxEnd := strings.Index(content[base:], "*/")
	if idxEnd == -1 {
		return ""
	}
	idxEnd += base

	return content[base:idxEnd]
}

func findStartEnd(content string, i int, start PatchMarker, end PatchMarker) (beginIdx int, beginEndIdx int, endIdx int, endEndIdx int) {
	beginIdx, beginEndIdx = findMarker(content, i, start, "")
	if beginIdx == -1 {
		return -1, -1, -1, -1
	}

	// label is the content after marker.Begin, before marker.End
	// e.g. /*<begin label>*//*<end label>*/
	label := content[beginIdx+len(start.Begin) : beginEndIdx-len(start.End)]

	endIdx, endEndIdx = findMarker(content, beginEndIdx, end, label)
	if endIdx == -1 {
		return -1, -1, -1, -1
	}

	return beginIdx, beginEndIdx, endIdx, endEndIdx
}

// findMarker finds the first occurrence of marker.Begin
// and marker.End in content starting from i.
// It returns the start and end indices of the marker.
// If the marker is not found, it returns -1, -1.
func findMarker(content string, i int, marker PatchMarker, label string) (int, int) {
	start := strings.Index(content[i:], marker.Begin+label)
	if start == -1 {
		return -1, -1
	}
	start += i
	i = start + len(marker.Begin) + len(label)
	end := strings.Index(content[i:], marker.End)
	if end == -1 {
		return -1, -1
	}
	end += i + len(marker.End)
	return start, end
}
