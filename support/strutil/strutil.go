package strutil

import "strings"

func IndexSequenceAt(s string, sequence []string, begin bool) int {
	return indexSequence(s, sequence, begin)
}

func IndexSequence(s string, sequence []string) int {
	return indexSequence(s, sequence, false)
}
func indexSequence(s string, sequence []string, begin bool) int {
	if len(sequence) == 0 {
		return 0
	}
	firstIdx := -1
	base := 0
	for _, seq := range sequence {
		idx := strings.Index(s, seq)
		if idx < 0 {
			return -1
		}
		if firstIdx < 0 {
			firstIdx = idx
		}
		s = s[idx+len(seq):]
		base += idx + len(seq)
	}
	if begin {
		return firstIdx
	}
	return base
}
