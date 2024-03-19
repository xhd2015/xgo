package main

import "fmt"

const VERSION = "1.0.1"
const REVISION = "cbdc296ff629fdba79e44589f2ce7fa1891c85f1+1"
const NUMBER = 82

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
