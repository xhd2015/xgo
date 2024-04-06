package main

import "fmt"

const VERSION = "1.0.18"
const REVISION = "86b60236af7fe7147b90073421f57187fbf6990a+1"
const NUMBER = 165

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
