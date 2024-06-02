package main

import "fmt"

const VERSION = "1.0.38"
const REVISION = "1bb3d68c0400e9594cd3e95a755cbe52f2283706+1"
const NUMBER = 251

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
