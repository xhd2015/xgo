package main

import "fmt"

const VERSION = "1.0.31"
const REVISION = "13df8328901b948d243243383aaf8d5b898ec6ca+1"
const NUMBER = 212

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
