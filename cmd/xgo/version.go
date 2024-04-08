package main

import "fmt"

const VERSION = "1.0.19"
const REVISION = "fe2e2f38793f24bda0a8690eedba01828ef79916+1"
const NUMBER = 169

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
