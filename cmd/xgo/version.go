package main

import "fmt"

const VERSION = "1.0.16"
const REVISION = "6f59a908da45484b2104d51e2fa56d83d3bf65a5+1"
const NUMBER = 153

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
