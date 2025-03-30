package main

import (
	"encoding/json"
	"fmt"

	"github.com/xhd2015/xgo/runtime/trace"
)

func main() {
	trace.Options().Name("main").OnComplete(func(root *trace.Root) {
		trace, err := json.Marshal(root.Export())
		if err != nil {
			fmt.Printf("trace err: %v\n", err)
			return
		}
		fmt.Printf("collected: %s\n", trace)
	}).Collect(func() {
		helloWorld()
	})
}

func helloWorld() {
	fmt.Printf("%s %s\n", hello(), world())
}

func hello() string {
	return "hello"
}
func world() string {
	return "world"
}
