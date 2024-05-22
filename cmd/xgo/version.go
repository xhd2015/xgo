package main

import "fmt"

const VERSION = "1.0.36"
const REVISION = "1e1b9419279db2f75b8619256fced2e8a84e9ee7+1"
const NUMBER = 229

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
