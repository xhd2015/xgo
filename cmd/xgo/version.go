package main

import "fmt"

const VERSION = "1.0.6"
const REVISION = "c5d26155557db3f039bbf0c92c4da64e34143738+1"
const NUMBER = 105

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
