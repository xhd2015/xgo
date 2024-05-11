package main

import "fmt"

const VERSION = "1.0.31"
const REVISION = "771bf0c6ea6525f6c02faada38c18c5d4870fbd9+1"
const NUMBER = 211

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
