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
	off, _ := indexSequence(s, sequence, idx, begin)
	return off
}

// - `offset` is the begin offset of the `i`'s anchor
// - `endOffset` is the end offset, which is the last anchor's starting offset plus its length
func SequenceOffset(s string, sequence []string, i int, begin bool) (offset int, endOffset int) {
	return indexSequence(s, sequence, i, begin)
}

func IndexSequence(s string, sequence []string) int {
	off, _ := indexSequence(s, sequence, -1, false)
	return off
}

// [a,b,c]
// before ->
func indexSequence(s string, sequence []string, seqIdx int, begin bool) (offset int, endOffset int) {
	if len(sequence) == 0 {
		return 0, 0
	}
	if seqIdx == -1 {
		seqIdx = len(sequence) - 1
	} else if seqIdx < 0 || seqIdx >= len(sequence) {
		return -1, -1
	}
	var recordOff int
	cursor := 0
	for i, seq := range sequence {
		idx := strings.Index(s, seq)
		if idx < 0 {
			return -1, -1
		}
		s = s[idx+len(seq):]
		cursor += idx + len(seq)
		if i == seqIdx {
			recordOff = cursor
			if begin {
				recordOff -= len(seq)
			}
		}
	}
	return recordOff, cursor
}

func CheckSequence(output string, sequence []string) error {
	idx, _ := indexSequence(output, sequence, -1, false)
	if idx < 0 {
		return fmt.Errorf("sequence %q missing from %q", sequence, output)
	}
	return nil
}
