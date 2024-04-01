package pathsum

import (
	"crypto/md5"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
)

func PathSum(prefix string, path string) (string, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	return prefix + shortPath(path, 8) + "_" + getIdentSum(path), nil
}

func getIdentSum(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	sum := hex.EncodeToString(h.Sum(nil))
	if len(sum) < 8 {
		return strings.Repeat("0", 8-len(sum)) + sum
	}
	return sum[:8]
}

func shortPath(path string, maxSeg int) string {
	segs := strings.Split(path, string(os.PathSeparator))
	shortSeg := make([]string, 0, len(segs))
	for _, seg := range segs {
		seg = processSpecial(seg)
		// seg = strings.ReplaceAll(seg, "/", "")
		// seg = strings.ReplaceAll(seg, "\\", "")
		// seg = strings.ReplaceAll(seg, "?", "")
		// seg = strings.ReplaceAll(seg, "$", "")
		// seg = strings.ReplaceAll(seg, "&", "")
		// seg = strings.ReplaceAll(seg, ":", "") // Windows
		// seg = strings.ReplaceAll(seg, ";", "")
		// seg = strings.ReplaceAll(seg, "%", "")
		// seg = strings.ReplaceAll(seg, "#", "")
		// seg = strings.ReplaceAll(seg, "!", "")
		// seg = strings.ReplaceAll(seg, "@", "")
		// seg = strings.ReplaceAll(seg, "*", "")
		// seg = strings.ReplaceAll(seg, "~", "")
		// seg = strings.ReplaceAll(seg, "+", "")
		// seg = strings.ReplaceAll(seg, "+", "")
		if len(seg) > 2 {
			seg = seg[0:2]
		}
		if seg == "" {
			continue
		}
		shortSeg = append(shortSeg, seg)
		if maxSeg > 0 && len(shortSeg) >= maxSeg {
			break
		}
	}
	return strings.Join(shortSeg, "_")
}

func processSpecial(path string) string {
	chars := []rune(path)
	n := len(chars)
	j := 0
	for i := 0; i < n; i++ {
		ch := chars[i]
		if ch < 128 && !(ch == '_' || ch == '-' || (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')) {
			continue
		}
		chars[j] = chars[i]
		j++
	}
	return string(chars[:j])
}
