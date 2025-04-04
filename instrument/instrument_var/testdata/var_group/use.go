package simple

import (
	"fmt"

	"github.com/xhd2015/xgo/instrument/instrument_var/testdata/var_group/pkg"
)

func UseVar() {
	fmt.Printf("SomeVar: %d\n", pkg.SomeVar)
	fmt.Printf("SomeVar2: %d\n", &pkg.SomeVar2)
}
