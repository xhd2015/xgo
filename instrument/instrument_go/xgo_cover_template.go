// This file will be added to instrumented $GOROOT/src/cmd/cover
// the 'go:build ignore' line will be removed when copied

//go:build ignore

package main

import (
	"bytes"
	"encoding/json"
	"strconv"
)

type XgoEdit struct {
	Start int    `json:"start,omitempty"`
	End   int    `json:"end,omitempty"`
	New   string `json:"new,omitempty"`
}

func xgoParseApply(nameInOut *string, contentInOut *[]byte) []XgoEdit {
	editFile, edits, original, ok := xgoParseEdit(*contentInOut)
	if ok {
		*nameInOut = editFile
		*contentInOut = []byte(original)
	}
	return edits
}

// we need to find the last 2 lines starting with
// // __xgo_edit_file: "..."
// // __xgo_edits: ...
// // __xgo_original: ...
// see https://github.com/xhd2015/xgo/issues/301 for more details
func xgoParseEdit(content []byte) (string, []XgoEdit, string, bool) {
	lastLines := xgoFindLinesFromEnd(content, 3)
	if len(lastLines) < 3 {
		return "", nil, "", false
	}

	var editFileRaw []byte
	var editsJSON []byte
	var originalRaw []byte
	// expect all last 3 lines have xgo prefix
	for _, line := range lastLines {
		prefix, content := xgoCutPrefixAndSpace(line)
		if len(prefix) == 0 {
			// fmt.Fprintf(os.Stderr, "DEBUG prefix is empty shit: %s\n", string(line))
			return "", nil, "", false
		}
		switch string(prefix) {
		case "// __xgo_file:":
			editFileRaw = content
		case "// __xgo_edits:":
			editsJSON = content
		case "// __xgo_original:":
			originalRaw = content
		default:
			// unrecognized prefix
			return "", nil, "", false
		}
	}

	if len(editFileRaw) == 0 || len(editsJSON) == 0 || len(originalRaw) == 0 {
		return "", nil, "", false
	}

	editFile, _ := strconv.Unquote(string(editFileRaw))
	if editFile == "" {
		return "", nil, "", false
	}

	var edits []XgoEdit
	_ = json.Unmarshal(editsJSON, &edits)
	if len(edits) == 0 {
		return "", nil, "", false
	}
	originalStr, _ := strconv.Unquote(string(originalRaw))
	if len(originalStr) == 0 {
		return "", nil, "", false
	}
	return editFile, edits, originalStr, true
}

func xgoFindLinesFromEnd(content []byte, max int) [][]byte {
	if max == 0 {
		return nil
	}
	var res [][]byte
	count := len(content)

	if count == 0 {
		return nil
	}
	// skip empty line
	if content[count-1] == '\n' {
		count--
	}

	last := count
	for i := count - 1; i >= 0; i-- {
		if content[i] == '\n' {
			// exclude the prefix and suffix '\n'
			res = append(res, content[i+1:last])
			if len(res) >= max {
				break
			}
			last = i
		}
	}
	// reverse
	for i, j := 0, len(res)-1; i < j; i, j = i+1, j-1 {
		res[i], res[j] = res[j], res[i]
	}
	return res
}

const xgoPrefix = "// __xgo_"

func xgoCutPrefixAndSpace(content []byte) (prefix []byte, actualContent []byte) {
	xgoPrefixBytes := []byte(xgoPrefix)
	if !bytes.HasPrefix(content, xgoPrefixBytes) {
		return nil, nil
	}
	n := len(content)
	i := len(xgoPrefixBytes)
	for ; i < n; i++ {
		if content[i] == ':' {
			break
		}
	}
	if i >= n {
		return nil, nil
	}

	return content[:i+1], bytes.TrimSpace(content[i+1:])
}
