package main

import "fmt"

const VERSION = "1.0.22"
const REVISION = "f174b18f76bff4bd8acddffd03835b760dad03d8+1"
const NUMBER = 175

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
