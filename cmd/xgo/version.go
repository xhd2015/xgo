package main

import "fmt"

const VERSION = "1.0.1"
const REVISION = "6eca2e984823b4e5d9621e9bd82e4d8db2b18895+1"
const NUMBER = 81

func getRevision() string {
	return fmt.Sprintf("%s %s BUILD_%d", VERSION, REVISION, NUMBER)
}
