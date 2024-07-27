// package unpatch removes all patch segments in a file
// restore a patched file to its original state,then
// apply patch again

package unpatch

import (
	"fmt"
	"strings"

	"github.com/xhd2015/xgo/support/transform/patch/format"
)

// format:
//  /*<begin X>*/ ... /*<end X>*/

func Unpatch(text string) (string, error) {
	var b strings.Builder
	b.Grow(len(text))

	i := 0
	for {
		l, r, err := findRange(text, i)
		if err != nil {
			return "", err
		}
		if l < 0 {
			b.WriteString(text[i:])
			break
		}

		repL, repR, err := findReplaced(text, l, r)
		if err != nil {
			return "", err
		}
		b.WriteString(text[i:l])
		if repL >= 0 {
			b.WriteString(text[repL:repR])
		}

		i = r
	}

	return b.String(), nil
}

func findRange(s string, start int) (int, int, error) {
	i := strings.Index(s[start:], format.BEGIN)
	if i < 0 {
		return -1, -1, nil
	}

	left := start + i

	i = left + len(format.BEGIN)

	const COMMENT_CLOSE = "*/"
	closeIdx := strings.Index(s[i:], COMMENT_CLOSE)
	if closeIdx < 0 {
		return 0, 0, fmt.Errorf("missing close for %s", ellipse(s[left:], 20))
	}
	if s[i+closeIdx-1] != '>' {
		return 0, 0, fmt.Errorf("missing close for %s", ellipse(s[left:], 20))
	}
	closeIdx--
	id := s[i : i+closeIdx]
	if id == "" {
		return 0, 0, fmt.Errorf("missing id for %s", ellipse(s[left:], 20))
	}
	i += closeIdx + len(format.CLOSE)

	end := format.END + id + format.CLOSE
	endIdx := strings.Index(s[i:], end)
	if endIdx < 0 {
		return 0, 0, fmt.Errorf("missing %s", end)
	}

	right := endIdx + i + len(end)
	return left, right, nil
}

func findReplaced(s string, i int, j int) (int, int, error) {
	sub := s[i:j]
	left := strings.Index(sub, format.REPLACED_BEGIN)
	if left < 0 {
		return -1, -1, nil
	}
	p := left + len(format.REPLACED_BEGIN)

	right := strings.Index(sub[p:], format.REPLACED_END)
	if right < 0 {
		return -1, -1, fmt.Errorf("missing %s", format.REPLACED_END)
	}

	return left + i + len(format.REPLACED_BEGIN), right + p + i, nil
}

func ellipse(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
