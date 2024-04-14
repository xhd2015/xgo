package main

import "fmt"

const VERSION = "1.0.24"
const REVISION = "70a2950e60bbb075af333afa7a702fc02b9d55ed+1"
const NUMBER = 182

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
