package main

import "fmt"

const VERSION = "1.0.36"
const REVISION = "e44373cb3c83b85599797e1f0cb302f81a95d598+1"
const NUMBER = 225

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
