package main

import "fmt"

const VERSION = "1.0.8"
const REVISION = "5ff39051a8c369774b3c6a17a6cccb3778153d8a+1"
const NUMBER = 118

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
