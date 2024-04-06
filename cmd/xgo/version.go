package main

import "fmt"

const VERSION = "1.0.19"
const REVISION = "7b94c97a1b438510f46aa86fca2b626396b9d2e1+1"
const NUMBER = 166

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
