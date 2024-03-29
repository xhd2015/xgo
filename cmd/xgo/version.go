package main

import "fmt"

const VERSION = "1.0.10"
const REVISION = "90cf0c0b5fe2b0b0bfe8078ed7ee4774f18c279d+1"
const NUMBER = 123

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
