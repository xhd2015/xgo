package simple

import (
	"fmt"

	"github.com/xhd2015/xgo/instrument/instrument_var/testdata/simple/pkg"
)

var localVar int

func UseLocalVar() {
	fmt.Printf("localVar: %d\n", localVar)
}

func UseVar() {
	fmt.Printf("SomeVar: %d\n", pkg.SomeVar)
}
