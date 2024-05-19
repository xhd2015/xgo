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
	args = append(args, "log", "--format=%H", "-1")
	if ref != "" {
		args = append(args, ref)
	}
	return cmd.Dir(dir).Output("git", args...)
}

// git cat-file -p REF:FILE
func GetFileContent(dir string, ref string, file string) (string, error) {
	if ref == "" {
		return "", fmt.Errorf("requires ref")
	}
	if file == "" {
		return "", fmt.Errorf("requires file")
	}
	return cmd.Dir(dir).Output("git", "cat-file", "-p", fmt.Sprintf("%s:%s", ref, file))
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

var constNumberSeq = []string{"const", "NUMBER", "="}

func IncrementNumber(s string) (string, error) {
	return replaceOrIncrementNumber(s, -1)
}
func replaceOrIncrementNumber(s string, version int) (string, error) {
	replaceLine := func(line string, index int) (string, error) {
		start, end, num, err := parseNum(line[index:])
		if err != nil {
			return "", fmt.Errorf("line %q: %w", line, err)
		}
		start += index
		end += index
		if version < 0 {
			version = num + 1
		}

		return line[:start] + strconv.Itoa(version) + line[end:], nil
	}
	return replaceSequence(s, constNumberSeq, replaceLine)
}

func isDigit(r rune) bool {
	return '0' <= r && r <= '9'
}

func parseNum(str string) (start int, end int, num int, err error) {
	start = strings.IndexFunc(str, isDigit)
	if start < 0 {
		return 0, 0, 0, fmt.Errorf("no number found")
	}

	end = strings.LastIndexFunc(str[start+1:], isDigit)
	if end < 0 {
		return 0, 0, 0, fmt.Errorf("no number found")
	}
	end += start + 2

	numStr := str[start:end]

	num, err = strconv.Atoi(numStr)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid number: %w", err)
	}
	return start, end, int(num), nil
}

func GetVersionNumber(s string) (int, error) {
	// const NUMBER =
	lines := strings.Split(s, "\n")
	lineIdx, byteIdx := IndexLinesSequence(lines, constNumberSeq)
	if lineIdx < 0 {
		return 0, fmt.Errorf("sequence %v not found", constNumberSeq)
	}
	_, _, num, err := parseNum(lines[lineIdx][byteIdx:])
	return num, err
}

func IndexLinesSequence(lines []string, seq []string) (lineIndex int, byteIndex int) {
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
		return -1, -1
	}
	return lineIdx, byteIdx
}
func replaceSequence(s string, seq []string, replaceLine func(line string, index int) (string, error)) (string, error) {
	lines := strings.Split(s, "\n")
	lineIdx, byteIdx := IndexLinesSequence(lines, seq)
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

func PatchVersionFile(file string, rev string, autIncrementNumber bool, version int) error {
	err := fileutil.Patch(file, func(data []byte) ([]byte, error) {
		content := string(data)
		newContent, err := ReplaceRevision(content, rev)
		if err != nil {
			return nil, err
		}
		if version > 0 {
			newContent, err = replaceOrIncrementNumber(newContent, version)
			if err != nil {
				return nil, err
			}
		} else if autIncrementNumber {
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
