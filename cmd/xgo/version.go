package main

import "fmt"

const VERSION = "1.0.4"
const REVISION = "33211d65fc8548c108bdc03947d0b17244441b8f+1"
const NUMBER = 91

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
