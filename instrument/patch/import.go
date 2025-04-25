package patch

import "fmt"

// package something -> package something;import __my_fmt "fmt"
func AddImportAfterName(code string, beginMark string, endMark string, name string, pkgPath string) string {
	var insertContent string
	if name == "" {
		insertContent = fmt.Sprintf(";import %q", pkgPath)
	} else {
		insertContent = fmt.Sprintf(";import %s %q", name, pkgPath)
	}
	_, end := mustFindPackageName(code)
	content, ok := tryReplaceWithMark(code, beginMark, endMark, "", insertContent)
	if ok {
		return content
	}
	return code[:end] + beginMark + insertContent + endMark + code[end:]
}
