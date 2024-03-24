package main

import "fmt"

const VERSION = "1.0.6"
const REVISION = "f07938bdff12fa4bd066d0e9cfde64e99f17122a+1"
const NUMBER = 99

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
