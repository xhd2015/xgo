package main

import "fmt"

const VERSION = "1.0.15"
const REVISION = "2861a46387df90bcadae7651dc6e0d2db8ab0148+1"
const NUMBER = 152

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
