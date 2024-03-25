package main

import "fmt"

const VERSION = "1.0.6"
const REVISION = "a90c6ef098ea021488a507ab255b738b8712c3be+1"
const NUMBER = 104

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
