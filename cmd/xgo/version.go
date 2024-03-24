package main

import "fmt"

const VERSION = "1.0.6"
const REVISION = "d9ad37e4e3332b051fcb718eb7eb9f06203a7b7d+1"
const NUMBER = 98

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
