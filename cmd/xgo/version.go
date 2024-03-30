package main

import "fmt"

const VERSION = "1.0.11"
const REVISION = "1bc959ae70c34e189c36b8cb3f7b2e1a8665756c+1"
const NUMBER = 136

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
