package patch

import (
	"fmt"
	"strings"

	"github.com/xhd2015/xgo/support/fileutil"
	"github.com/xhd2015/xgo/support/strutil"
)

func EditFile(file string, callback func(content string) (string, error)) error {
	bytes, err := fileutil.ReadFile(file)
	if err != nil {
		return err
	}
	content := CleanPatch(string(bytes))
	newContent, err := callback(content)
	if err != nil {
		return err
	}
	if newContent == content {
		return nil
	}
	return fileutil.WriteFile(file, []byte(newContent))
}

// Deprecated: use AddContentAtIndex instead
func AddCodeAfterImports(code string, beginMark string, endMark string, contents []string) string {
	idx := indexSeq(code, []string{"import", "(", "\n"}, false)
	if idx < 0 {
		panic(fmt.Errorf("import not found"))
	}
	return insertContentLinesNoDuplicate(code, beginMark, endMark, idx, strings.Join(contents, "\n")+"\n")
}

// Deprecated: use AddContentAtIndex instead
func AddContentBefore(content string, beginMark string, endMark string, seq []string, addContent string) string {
	return UpdateContentLines(content, beginMark, endMark, seq, 0, true, addContent)
}

// Deprecated: use AddContentAtIndex instead
func AddContentAfter(content string, beginMark string, endMark string, seq []string, addContent string) string {
	return addContentAt(content, beginMark, endMark, seq, addContent, false)
}

func addContentAt(content string, beginMark string, endMark string, seq []string, addContent string, begin bool) string {
	idx := indexSeq(content, seq, begin)
	if idx < 0 {
		panic(fmt.Errorf("sequence not found: %v", seq))
	}
	return insertContentLinesNoDuplicate(content, beginMark, endMark, idx, addContent)
}

type UpdatePosition bool

const (
	UpdatePosition_After  UpdatePosition = false
	UpdatePosition_Before UpdatePosition = true
)

// UpdateContentLines add content before or after the `i`'s anchor in `seq`
// two lines will be automatically added after `beginMark`, and before `endMark`, so you don't need to include a line in `addContent`
func UpdateContentLines(content string, beginMark string, endMark string, seq []string, i int, position UpdatePosition, addContent string) string {
	return updateContent(content, beginMark, endMark, seq, i, position, addContent, "\n")
}

func UpdateContent(content string, beginMark string, endMark string, seq []string, i int, position UpdatePosition, addContent string) string {
	return updateContent(content, beginMark, endMark, seq, i, position, addContent, "")
}

func updateContent(content string, beginMark string, endMark string, seq []string, i int, position UpdatePosition, addContent string, separator string) string {
	offset, endOffset := strutil.SequenceOffset(content, seq, i, bool(position))
	if offset < 0 {
		panic(fmt.Errorf("sequence missing: %v", seq))
	}
	// ensure sequence is unique
	anotherOff, _ := strutil.SequenceOffset(content[endOffset:], seq, i, false)
	if anotherOff >= 0 {
		panic(fmt.Errorf("sequence duplicate: %v", seq))
	}
	return insertContentNoDuplicate(content, beginMark, endMark, offset, addContent, separator)
}

func ReplaceContentAfter(content string, beginMark string, endMark string, seq []string, target string, replaceContent string) string {
	if replaceContent == "" {
		return content
	}
	closuerContent := beginMark + "\n" + replaceContent + "\n" + endMark + "\n"
	idx := indexSeq(content, seq, false)
	if idx < 0 {
		panic(fmt.Errorf("sequence not found: %v", seq))
	}
	if strings.Contains(content, closuerContent) {
		return content
	}
	content, ok := tryReplaceWithMark(content, beginMark, endMark, closuerContent)
	if ok {
		return content
	}
	targetIdx := strings.Index(content[idx:], target)
	if targetIdx < 0 {
		panic(fmt.Errorf("not found: %s", target))
	}
	return content[:idx+targetIdx] + closuerContent + content[idx+targetIdx+len(target):]
}

// signature example: /*<begin ident>*/ {content} /*<end ident>*/
// insert content at index
func insertContentLinesNoDuplicate(content string, beginMark string, endMark string, idx int, insertContent string) string {
	return insertContentNoDuplicate(content, beginMark, endMark, idx, insertContent, "\n")
}

func insertContentNoDuplicate(content string, beginMark string, endMark string, idx int, insertContent string, separator string) string {
	if insertContent == "" {
		return content
	}
	closuerContent := beginMark + separator + insertContent + separator + endMark + separator
	content, ok := tryReplaceWithMark(content, beginMark, endMark, closuerContent)
	if ok {
		return content
	}
	if strings.Contains(content, closuerContent) {
		return content
	}
	return content[:idx] + closuerContent + content[idx:]
}

func tryReplaceWithMark(content string, beginMark string, endMark string, closureContent string) (string, bool) {
	beginIdx := strings.Index(content, beginMark)
	if beginIdx < 0 {
		return content, false
	}
	endIdx := strings.Index(content, endMark)
	if endIdx < 0 {
		return content, false
	}
	lastIdx := endIdx + len(endMark)
	return content[:beginIdx] + closureContent + content[lastIdx:], true
}

func indexSeq(s string, sequence []string, begin bool) int {
	return strutil.IndexSequenceAt(s, sequence, begin)
}
