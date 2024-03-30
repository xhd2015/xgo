package main

import "fmt"

const VERSION = "1.0.11"
const REVISION = "b1b21571259df7c1632d86cef35ba681727a0cde+1"
const NUMBER = 143

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
