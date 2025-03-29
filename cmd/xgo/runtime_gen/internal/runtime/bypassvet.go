package runtime

import (
	"fmt"
	"os"
)

// with runtime_link.go importing fmt,os
// and then it is replaced by runtime_link_template.go
// go test will report the following vet error:

// 	# github.com/xhd2015/xgo/runtime/internal/runtime
// 	vet: vendor/github.com/xhd2015/xgo/runtime/internal/runtime/runtime_link.go:14:2: undeclared name: fmt

// causing the test to fail
// TODO: report the bug to the go team,see https://github.com/xhd2015/xgo/issues/298
func logError(msg string) {
	fmt.Fprintln(os.Stderr, msg)
}
