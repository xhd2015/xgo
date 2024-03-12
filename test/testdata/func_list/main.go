package main

import (
	"fmt"

	"github.com/xhd2015/xgo/runtime/core/functab"
)

func main() {
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
}

func example() {

}

type someInt int

func (c someInt) value() int {
	return int(c)
}
func (c *someInt) inc() int {
	*c++
	return int(*c)
}
