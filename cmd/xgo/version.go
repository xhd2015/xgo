package main

import "fmt"

const VERSION = "1.0.22"
const REVISION = "f34bf4ca38af8b5adcf36b67de5ea6fb853e4823+1"
const NUMBER = 174

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
