package main

import "fmt"

const VERSION = "1.0.38"
const REVISION = "04f19fc3545957ade879144d189553b48be40419+1"
const NUMBER = 255

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
