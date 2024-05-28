package strutil

import "strings"

func IndentLines(content string, prefix string) string {
	lines := strings.Split(content, "\n")
	n := len(lines)
	for i := 0; i < n; i++ {
		lines[i] = prefix + lines[i]
	}
	return strings.Join(lines, "\n")
}
