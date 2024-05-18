package main

import "fmt"

const VERSION = "1.0.35"
const REVISION = "3431198dea60ee1a41ed69b93034c5919aba32b5+1"
const NUMBER = 218

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
