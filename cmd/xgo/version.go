package main

import "fmt"

const VERSION = "1.0.22"
const REVISION = "b41281e9fca79eb3bbde5e4347ab8f37763bc545+1"
const NUMBER = 176

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
