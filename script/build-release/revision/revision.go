package revision

import (
	"fmt"
	"go/ast"
	"go/token"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/xhd2015/xgo/support/cmd"
	"github.com/xhd2015/xgo/support/fileutil"
	"github.com/xhd2015/xgo/support/goparse"
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

func ReplaceVersion(s string, version string) (string, error) {
	if strings.Contains(version, `"`) {
		return "", fmt.Errorf("version cannot have \": %s", version)
	}
	replaceLine := func(line string, index int) (string, error) {
		qIdx := strings.Index(line[index+1:], `"`)
		if qIdx < 0 {
			return "", fmt.Errorf("invalid VERSION variable, missing \"")
		}
		qIdx += index + 1
		endIdx := strings.Index(line[qIdx+1:], `"`)
		if endIdx < 0 {
			return "", fmt.Errorf("invalid VERSION variable, missing ending \"")
		}
		endIdx += qIdx + 1
		return line[:qIdx+1] + version + line[endIdx:], nil
	}
	return replaceSequence(s, []string{"const", "VERSION", "="}, replaceLine)
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
func replaceOrIncrementNumber(s string, number int) (string, error) {
	replaceLine := func(line string, index int) (string, error) {
		start, end, num, err := parseNum(line[index:])
		if err != nil {
			return "", fmt.Errorf("line %q: %w", line, err)
		}
		start += index
		end += index
		if number < 0 {
			number = num + 1
		}

		return line[:start] + strconv.Itoa(number) + line[end:], nil
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

	if start < len(str)-1 {
		end = strings.LastIndexFunc(str[start+1:], isDigit)
		if end < 0 {
			return 0, 0, 0, fmt.Errorf("no number found")
		}
		end += start + 2
	} else {
		end = start + 1
	}

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

func PatchVersionFile(file string, version string, rev string, autIncrementNumber bool, number int) error {
	err := fileutil.Patch(file, func(data []byte) ([]byte, error) {
		content := string(data)
		newContent := content
		var err error
		if version != "" {
			newContent, err = ReplaceVersion(newContent, version)
			if err != nil {
				return nil, err
			}
		}
		newContent, err = ReplaceRevision(newContent, rev)
		if err != nil {
			return nil, err
		}
		if number > 0 {
			newContent, err = replaceOrIncrementNumber(newContent, number)
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

func ParseVersionConstants(content string) (version string, revision string, number int, err error) {
	return parseVersionConstants(content, "VERSION", "REVISION", "NUMBER")
}

func ParseCoreVersionConstants(content string) (version string, revision string, number int, err error) {
	return parseVersionConstants(content, "CORE_VERSION", "CORE_REVISION", "CORE_NUMBER")
}

func parseVersionConstants(content string, versionName string, revisionName string, numberName string) (version string, revision string, number int, err error) {
	constants, err := ParseConstants(content)
	if err != nil {
		return
	}
	version, err = strconv.Unquote(constants[versionName])
	if err != nil {
		return
	}
	revision, err = strconv.Unquote(constants[revisionName])
	if err != nil {
		return
	}
	num := constants[numberName]
	if num != "" {
		var i int64
		i, err = strconv.ParseInt(num, 10, 64)
		if err != nil {
			return
		}
		number = int(i)
	}
	return
}

// parse
func ParseConstants(content string) (map[string]string, error) {
	file, _, err := goparse.ParseFileCode("src.go", []byte(content))
	if err != nil {
		return nil, err
	}

	constants := make(map[string]string)
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		if genDecl.Tok != token.CONST {
			continue
		}
		for _, spec := range genDecl.Specs {
			constSpec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			var names []string
			for _, name := range constSpec.Names {
				names = append(names, name.Name)
			}
			var values []string
			for _, value := range constSpec.Values {
				var valueLit string
				if lit, ok := value.(*ast.BasicLit); ok {
					valueLit = lit.Value
				} else {
					return nil, fmt.Errorf("unknown const spec value: %T %v", value, value)
				}
				values = append(values, valueLit)
			}
			if len(names) != len(values) {
				return nil, fmt.Errorf("mismatch decl: names=%v,values=%v", names, values)
			}

			for i, name := range names {
				constants[name] = values[i]
			}
		}
	}
	return constants, nil
}

func GetXgoVersionFile(rootDir string) string {
	return filepath.Join(rootDir, "cmd", "xgo", "version.go")
}
func GetRuntimeVersionFile(rootDir string) string {
	return filepath.Join(rootDir, "runtime", "core", "version.go")
}

func GetVersionFiles(rootDir string) []string {
	return []string{
		GetXgoVersionFile(rootDir),

		// NOTE: runtime version not automatically updated,
		// see https://github.com/xhd2015/xgo/issues/216
		// GetRuntimeVersionFile(rootDir),
	}
}

func CopyCoreVersion(xgoVersionFile string, runtimeVersionFile string) error {
	// copy xgo's core version to runtime version
	code, err := ioutil.ReadFile(xgoVersionFile)
	if err != nil {
		return err
	}
	coreVersion, coreRevision, coreNumber, err := ParseCoreVersionConstants(string(code))
	if err != nil {
		return err
	}
	if coreVersion == "" || coreRevision == "" || coreNumber <= 0 {
		return fmt.Errorf("invalid core version: %s", xgoVersionFile)
	}
	return PatchVersionFile(runtimeVersionFile, coreVersion, coreRevision, false, coreNumber)
}
