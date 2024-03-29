package main

import "fmt"

const VERSION = "1.0.8"
const REVISION = "f911dc86a2f7dbb62b2361f3649c823e0b2ea4b8+1"
const NUMBER = 119

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
