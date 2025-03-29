package patch

import "fmt"

// for line directive, check https://github.com/golang/go/blob/24b395119b4df7f16915b9f01a6aded647b79bbd/src/cmd/compile/doc.go#L171
// this tells the compiler, next line's line number is `line`
func FmtLineDirective(srcFile string, line int) string {
	return fmt.Sprintf("//line %s:%d", srcFile, line)
}
