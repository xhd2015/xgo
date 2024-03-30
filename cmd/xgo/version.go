package main

import "fmt"

const VERSION = "1.0.11"
const REVISION = "b74c3b3830780e4793716b2b48505f22502528f2+1"
const NUMBER = 125

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
