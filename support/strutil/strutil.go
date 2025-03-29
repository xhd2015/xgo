package strutil

import (
	"fmt"
	"strings"
)

func IndexSequenceAt(s string, sequence []string, begin bool) int {
	idx := 0
	if !begin {
		idx = -1
	}
	off, _, _ := indexSequence(s, sequence, idx, begin)
	return off
}

// - `anchorOffset` is the offset of the `i`'s anchor, if `begin` is false, it adds the length of the anchor
// - `anchorLen` is the length of the `i`'s anchor
// - `endOffset` is the end offset, which is the last anchor's starting offset plus its length
func SequenceOffset(s string, sequence []string, i int, begin bool) (anchorOffset int, anchorLen int, endOffset int) {
	return indexSequence(s, sequence, i, begin)
}

func IndexSequence(s string, sequence []string) int {
	off, _, _ := indexSequence(s, sequence, -1, false)
	return off
}

// [a,b,c]
// before ->
func indexSequence(s string, sequence []string, anchorIdx int, begin bool) (offset int, anchorLen int, endOffset int) {
	if len(sequence) == 0 {
		return 0, 0, 0
	}
	if anchorIdx == -1 {
		anchorIdx = len(sequence) - 1
	} else if anchorIdx < 0 || anchorIdx >= len(sequence) {
		return -1, -1, -1
	}
	var recordOff int
	cursor := 0
	for i, seq := range sequence {
		idx := strings.Index(s, seq)
		if idx < 0 {
			return -1, -1, -1
		}
		s = s[idx+len(seq):]
		cursor += idx + len(seq)
		if i == anchorIdx {
			recordOff = cursor
			if begin {
				recordOff -= len(seq)
			}
			anchorLen = len(seq)
		}
	}
	return recordOff, anchorLen, cursor
}

func CheckSequence(output string, sequence []string) error {
	idx, _, _ := indexSequence(output, sequence, -1, false)
	if idx < 0 {
		return fmt.Errorf("sequence %q missing from %q", sequence, output)
	}
	return nil
}
