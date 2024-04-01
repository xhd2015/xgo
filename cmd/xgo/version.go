package main

import "fmt"

const VERSION = "1.0.11"
const REVISION = "43756010e13cabfae008c1de9d72f98b946b0a09+1"
const NUMBER = 144

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
