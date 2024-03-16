package main

import (
	"fmt"

	"github.com/xhd2015/xgo/runtime/functab"
)

// go run ./cmd/xgo run --project-dir ./runtime ./test/func_info
func main() {
	funcInfo := functab.Info(example)
	if funcInfo == nil {
		panic(fmt.Errorf("func not found"))
	}

	fmt.Printf("fullName: %s\n", funcInfo.FullName)
	fmt.Printf("args: %v\n", funcInfo.ArgNames)
}

func example(a string) {

}
