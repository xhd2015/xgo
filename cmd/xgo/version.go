package main

import "fmt"

const VERSION = "1.0.25"
const REVISION = "965c202f092256db5171d680d92030aa720cca4d+1"
const NUMBER = 188

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
