package main

import "fmt"

func main() {
	fmt.Println(greet("world"))
}

// greet
func greet(s string) string {
	return "hello " + s
}
