package strutil

import (
	"fmt"
	"strings"
)

func IndexSequenceAt(s string, sequence []string, begin bool) int {
	_, idx := indexSequence(s, sequence, begin)
	return idx
}

func IndexSequence(s string, sequence []string) int {
	_, idx := indexSequence(s, sequence, false)
	return idx
}
func indexSequence(s string, sequence []string, begin bool) (int, int) {
	if len(sequence) == 0 {
		return 0, 0
	}
	firstIdx := -1
	base := 0
	for i, seq := range sequence {
		idx := strings.Index(s, seq)
		if idx < 0 {
			return i, -1
		}
		if firstIdx < 0 {
			firstIdx = idx
		}
		s = s[idx+len(seq):]
		base += idx + len(seq)
	}
	if begin {
		return -1, firstIdx
	}
	return -1, base
}

func CheckSequence(output string, sequence []string) error {
	missing, idx := indexSequence(output, sequence, false)
	if idx < 0 {
		return fmt.Errorf("sequence at %d: missing %q", missing, sequence[missing])
	}
	return nil
}
