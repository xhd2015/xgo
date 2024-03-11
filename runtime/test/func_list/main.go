package main

import (
	"fmt"

	"github.com/xhd2015/xgo/runtime/core/functab"
)

func main() {
	fmt.Printf("func list begin\n")

	funcs := functab.GetFuncs()

	total := 0
	countByPkg := make(map[string]int)

	for _, fn := range funcs {
		// example:
		//   main main
		//   main A.String
		pkgPath := fn.PkgPath
		funcName := fn.Name
		fmt.Printf("func:%s %s\n", pkgPath, funcName)
		total++
		countByPkg[pkgPath]++
	}

	for pkgName, count := range countByPkg {
		fmt.Printf(" >> %s: %v\n", pkgName, count)
	}

	fmt.Printf("total: %d\n", total)
	// example output:
	// 	>> main: 2
	// 	>> runtime/internal/atomic: 94
	// 	>> runtime/internal/sys: 12
	// 	>> sort: 71
	// 	>> math: 159
	// 	>> sync/atomic: 67
	// 	>> sync: 87
	// 	>> syscall: 453
	// 	>> runtime/internal/math: 2
	// 	>> strconv: 108
	// 	>> io: 51
	// 	>> path: 15
	// 	>> io/fs: 42
	// 	>> errors: 8
	// 	>> unicode/utf8: 16
	// 	>> unicode: 28
	// 	>> reflect: 330
	// 	>> os: 199
	// 	>> math/bits: 49
	// 	>> time: 176
	// 	>> fmt: 129
	//    total: 2098
}

type A int

func (a A) String() string {
	return ""
}
