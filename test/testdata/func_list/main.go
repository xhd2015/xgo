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
		pkg := fn.Pkg
		displayName := fn.DisplayName()
		fmt.Printf("func:%s %s\n", pkg, displayName)
		total++
		countByPkg[pkg]++
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
