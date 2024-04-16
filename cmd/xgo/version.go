package main

import "fmt"

const VERSION = "1.0.24"
const REVISION = "ccc82ab8c3514f4c1a71039dad9ee7c30109935c+1"
const NUMBER = 184

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
