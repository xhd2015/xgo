package main

import (
	"fmt"

	"github.com/xhd2015/xgo/runtime/functab"
)

// go run ./cmd/xgo run --project-dir ./runtime ./test/func_info
func main() {
	funcInfo := functab.Info(example)
	if funcInfo == nil {
		panic(fmt.Errorf("func example not found"))
	}

	fmt.Printf("example fullName: %s\n", funcInfo.FullName)
	fmt.Printf("example args: %v\n", funcInfo.ArgNames)

	runGeneric()
}

func example(a string) {

}
