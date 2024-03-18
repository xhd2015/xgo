package strutil

import "strings"

func IndexSequence(s string, sequence []string) int {
	base := 0
	for _, seq := range sequence {
		idx := strings.Index(s, seq)
		if idx < 0 {
			return -1
		}
		s = s[idx+len(seq):]
		base += idx + len(seq)
	}
	return base
}
