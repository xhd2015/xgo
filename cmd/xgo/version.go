package main

import "fmt"

const VERSION = "1.0.7"
const REVISION = "c16fef5bf0691ecb53ce936332a854debd48d8f8+1"
const NUMBER = 111

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
