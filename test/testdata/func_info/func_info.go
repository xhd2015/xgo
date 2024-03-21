package main

import (
	"fmt"

	_ "reflect"

	"github.com/xhd2015/xgo/runtime/functab"
)

// go run ./cmd/xgo run --project-dir ./runtime ./test/func_info
func main() {
	funcInfo := functab.InfoFunc(example)
	if funcInfo == nil {
		panic(fmt.Errorf("func example not found"))
	}

	fmt.Printf("example identityName: %s\n", funcInfo.IdentityName)
	fmt.Printf("example args: %v\n", funcInfo.ArgNames)

	runGeneric()
}

func example(a string) {

}
