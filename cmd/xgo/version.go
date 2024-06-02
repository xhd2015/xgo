package main

import "fmt"

const VERSION = "1.0.38"
const REVISION = "7712843f2486408bcc20970ce0a7c4f2e9f4f7dd+1"
const NUMBER = 254

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
