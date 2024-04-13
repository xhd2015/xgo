package main

import "fmt"

const VERSION = "1.0.22"
const REVISION = "b95076921b299cab29e79a1642abf005462fb732+1"
const NUMBER = 177

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
