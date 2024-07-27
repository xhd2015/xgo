package main

import "fmt"

func main() {
	fmt.Println(greet("world"))
}

// greet
func greet(s string) string {
	     /*<begin hello>*/
         return "patched " +s
        /*<end hello>*/
        /*<begin hello>*/
        /*<replaced>
                return "hello " + s
        </replaced>*/
        /*<end hello>*/
}
