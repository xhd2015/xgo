package main

import "fmt"

const VERSION = "1.0.37"
const REVISION = "5009615738fb02321f124776178716796d32c0b7+1"
const NUMBER = 244

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
