package strutil

import "strings"

func TrimBefore(s string, p string) (string, bool) {
	return trimBefore(s, p)
}

func trimBefore(s string, p string) (string, bool) {
	i := strings.LastIndex(s, p)
	if i < 0 {
		return s, false
	}
	return s[i+len(p):], true
}
