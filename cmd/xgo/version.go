package main

import "fmt"

const VERSION = "1.0.6"
const REVISION = "f33f74e62755baf725f5f05e87bbf25c21af2951+1"
const NUMBER = 103

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
