package main

import "fmt"

const VERSION = "1.0.11"
const REVISION = "bdfb5083555459e8a25ce31cdde1706aaa58bc43+1"
const NUMBER = 137

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
