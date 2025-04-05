package patch

import (
	"fmt"
	"strings"

	"github.com/xhd2015/xgo/support/fileutil"
	"github.com/xhd2015/xgo/support/strutil"
)

func EditFile(file string, callback func(content string) (string, error)) (err error) {
	bytes, err := fileutil.ReadFile(file)
	if err != nil {
		return err
	}
	content := CleanPatch(string(bytes))
	defer func() {
		if e := recover(); e != nil {
			if pe, ok := e.(error); ok {
				err = pe
			} else {
				err = fmt.Errorf("panic: %v", e)
			}
		}
		if err != nil {
			err = fmt.Errorf("%s: %w", file, err)
		}
	}()
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
	return UpdateContentLines(content, beginMark, endMark, seq, 0, UpdatePosition_Before, addContent)
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

type UpdatePosition int

const (
	UpdatePosition_After   UpdatePosition = 0
	UpdatePosition_Before  UpdatePosition = 1
	UpdatePosition_Replace UpdatePosition = 2
)

// UpdateContentLines add content before or after the `i`'s anchor in `seq`
// two lines will be automatically added after `beginMark`, and before `endMark`, so you don't need to include a line in `addContent`
func UpdateContentLines(content string, beginMark string, endMark string, seq []string, i int, position UpdatePosition, addContent string) string {
	return updateContent(content, beginMark, endMark, seq, i, position, addContent, "\n")
}

func UpdateContent(content string, beginMark string, endMark string, seq []string, i int, position UpdatePosition, addContent string) string {
	return updateContent(content, beginMark, endMark, seq, i, position, addContent, "")
}

func updateContent(content string, beginMark string, endMark string, seq []string, i int, position UpdatePosition, patchContent string, separator string) string {
	isReplace := position == UpdatePosition_Replace
	if isReplace {
		replacedContent, ok := tryReplaceWithMark(content, beginMark, endMark, separator, patchContent)
		if ok {
			return replacedContent
		}
	}
	var begin bool
	if position == UpdatePosition_Before || isReplace {
		begin = true
	}
	offset, anchorLen, endOffset := strutil.SequenceOffset(content, seq, i, begin)
	if offset < 0 {
		// logS := content
		// if len(logS) > 1000 {
		// 	logS = logS[:1000] + "..."
		// }
		// fmt.Fprintf(os.Stderr, "DEBUG sequence %d %q missing from\n:%s\n", i, seq, logS)
		qseq := make([]string, len(seq))
		for i, s := range seq {
			qseq[i] = "  " + s
		}
		panic(fmt.Errorf("sequence missing:\n%v", strings.Join(qseq, "\n")))
	}
	// ensure sequence is unique
	anotherOff, _, _ := strutil.SequenceOffset(content[endOffset:], seq, i, false)
	if anotherOff >= 0 {
		panic(fmt.Errorf("sequence duplicate: %v", seq))
	}
	return insertOrReplaceContentNoDuplicate(content, isReplace, anchorLen, beginMark, endMark, offset, patchContent, separator)
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
	content, ok := tryReplaceWithMark(content, beginMark, endMark, "\n", replaceContent)
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
	return insertOrReplaceContentNoDuplicate(content, false, 0, beginMark, endMark, idx, insertContent, "\n")
}

func insertOrReplaceContentNoDuplicate(content string, replace bool, replaceLen int, beginMark string, endMark string, idx int, insertContent string, separator string) string {
	if insertContent == "" {
		return content
	}
	content, ok := tryReplaceWithMark(content, beginMark, endMark, separator, insertContent)
	if ok {
		return content
	}
	if replace {
		return content[:idx] + beginMark + fmt.Sprintf("/*old:%s*/", content[idx:idx+replaceLen]) + separator + insertContent + separator + endMark + separator + content[idx+replaceLen:]
	}
	closureContent := beginMark + separator + insertContent + separator + endMark + separator
	if strings.Contains(content, closureContent) {
		return content
	}
	return content[:idx] + closureContent + content[idx:]
}

func tryReplaceWithMark(content string, beginMark string, endMark string, separator string, patchContent string) (string, bool) {
	beginIdx := strings.Index(content, beginMark)
	if beginIdx < 0 {
		return content, false
	}
	endIdx := strings.Index(content, endMark)
	if endIdx < 0 {
		return content, false
	}
	oldContent := extractOldContent(content[beginIdx+len(beginMark) : endIdx])
	if oldContent != "" {
		patchContent = oldContent + patchContent
	}
	lastIdx := endIdx + len(endMark)
	return content[:beginIdx] + beginMark + separator + patchContent + separator + endMark + separator + content[lastIdx:], true
}

func indexSeq(s string, sequence []string, begin bool) int {
	return strutil.IndexSequenceAt(s, sequence, begin)
}
