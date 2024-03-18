package revision

import (
	"fmt"
	"path/filepath"
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
		return "", fmt.Errorf("revision connot have \": %s", revision)
	}
	lines := strings.Split(s, "\n")
	n := len(lines)
	lineIdx := -1
	byteIdx := -1
	for i := 0; i < n; i++ {
		line := lines[i]
		idx := strutil.IndexSequence(line, []string{"const", "REVISION", "="})
		if idx >= 0 {
			lineIdx = i
			byteIdx = idx
			break
		}
	}
	if lineIdx < 0 {
		return "", fmt.Errorf("variable REVISION not found")
	}
	line := lines[lineIdx]
	qIdx := strings.Index(line[byteIdx+1:], `"`)
	if qIdx < 0 {
		return "", fmt.Errorf("invalid REVISION variable, missing \"")
	}
	qIdx += byteIdx + 1
	endIdx := strings.Index(line[qIdx+1:], `"`)
	if endIdx < 0 {
		return "", fmt.Errorf("invalid REVISION variable, missing ending \"")
	}
	endIdx += qIdx + 1
	line = line[:qIdx+1] + revision + line[endIdx:]

	lines[lineIdx] = line

	return strings.Join(lines, "\n"), nil
}

func PatchVersionFile(file string, rev string) error {
	err := fileutil.Patch(file, func(data []byte) ([]byte, error) {
		content := string(data)
		newContent, err := ReplaceRevision(content, rev)
		if err != nil {
			return nil, err
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
