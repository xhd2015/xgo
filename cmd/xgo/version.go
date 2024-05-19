package main

import "fmt"

const VERSION = "1.0.35"
const REVISION = "efe86afb9931162ad0926648c13c4109383beef3+1"
const NUMBER = 223

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
