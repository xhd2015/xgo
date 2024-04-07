package main

import "fmt"

const VERSION = "1.0.19"
const REVISION = "7fd3f2b160c52c890d248efceb23a5e070156e0c+1"
const NUMBER = 167

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
