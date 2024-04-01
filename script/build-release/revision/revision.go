package revision

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/xhd2015/xgo/support/cmd"
	"github.com/xhd2015/xgo/support/fileutil"
	"github.com/xhd2015/xgo/support/strutil"
)

func GetCommitHash(dir string, ref string) (string, error) {
	var args []string
	if dir != "" {
		args = append(args, "-C", dir)
	}
	args = append(args, "log", "--format=%H", "-1")
	if ref != "" {
		args = append(args, ref)
	}
	return cmd.Output("git", args...)
}

func ReplaceRevision(s string, revision string) (string, error) {
	if strings.Contains(revision, `"`) {
		return "", fmt.Errorf("revision cannot have \": %s", revision)
	}
	replaceLine := func(line string, index int) (string, error) {
		qIdx := strings.Index(line[index+1:], `"`)
		if qIdx < 0 {
			return "", fmt.Errorf("invalid REVISION variable, missing \"")
		}
		qIdx += index + 1
		endIdx := strings.Index(line[qIdx+1:], `"`)
		if endIdx < 0 {
			return "", fmt.Errorf("invalid REVISION variable, missing ending \"")
		}
		endIdx += qIdx + 1
		return line[:qIdx+1] + revision + line[endIdx:], nil
	}
	return replaceSequence(s, []string{"const", "REVISION", "="}, replaceLine)
}

func IncrementNumber(s string) (string, error) {
	replaceLine := func(line string, index int) (string, error) {
		isDigit := func(r rune) bool {
			return '0' <= r && r <= '9'
		}
		start := strings.IndexFunc(line[index+1:], isDigit)
		if start < 0 {
			return "", fmt.Errorf("no number found: %s", line)
		}
		start += index + 1

		end := strings.LastIndexFunc(line[start:], isDigit)
		if end < 0 {
			return "", fmt.Errorf("no number found: %s", line)
		}
		end += start + 1

		numStr := line[start:end]

		num, err := strconv.ParseInt(numStr, 10, 64)
		if err != nil {
			return "", fmt.Errorf("invalid number %s: %w", line, err)
		}
		newNum := num + 1
		return line[:start] + strconv.FormatInt(newNum, 10) + line[end:], nil
	}
	return replaceSequence(s, []string{"const", "NUMBER", "="}, replaceLine)
}

func replaceSequence(s string, seq []string, replaceLine func(line string, index int) (string, error)) (string, error) {
	lines := strings.Split(s, "\n")
	n := len(lines)
	lineIdx := -1
	byteIdx := -1
	for i := 0; i < n; i++ {
		line := lines[i]
		idx := strutil.IndexSequence(line, seq)
		if idx >= 0 {
			lineIdx = i
			byteIdx = idx
			break
		}
	}
	if lineIdx < 0 {
		return "", fmt.Errorf("sequence %v not found", seq)
	}
	var err error
	lines[lineIdx], err = replaceLine(lines[lineIdx], byteIdx)
	if err != nil {
		return "", err
	}

	return strings.Join(lines, "\n"), nil
}

func PatchVersionFile(file string, rev string, incrementNumber bool) error {
	err := fileutil.Patch(file, func(data []byte) ([]byte, error) {
		content := string(data)
		newContent, err := ReplaceRevision(content, rev)
		if err != nil {
			return nil, err
		}
		if incrementNumber {
			newContent, err = IncrementNumber(newContent)
			if err != nil {
				return nil, err
			}
		}
		return []byte(newContent), nil
	})
	if err != nil {
		return fmt.Errorf("%s: %w", file, err)
	}
	return nil
}

func GetVersionFiles(rootDir string) []string {
	return []string{
		filepath.Join(rootDir, "cmd", "xgo", "version.go"),
		filepath.Join(rootDir, "runtime", "core", "version.go"),
	}
}
