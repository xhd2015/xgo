package fileutil

func CleanSpecial(path string) string {
	chars := []rune(path)
	n := len(chars)
	j := 0
	for i := 0; i < n; i++ {
		ch := chars[i]
		if ch == ' ' {
			chars[j] = '_'
			j++
			continue
		}
		if ch < 128 && !(ch == '.' || ch == '_' || ch == '-' || (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')) {
			continue
		}
		chars[j] = chars[i]
		j++
	}
	return string(chars[:j])
}
