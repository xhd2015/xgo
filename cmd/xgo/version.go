package main

import "fmt"

const VERSION = "1.0.38"
const REVISION = "f197fc6ef0ef40ce20374c7c1469fe36e36a4894+1"
const NUMBER = 247

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
