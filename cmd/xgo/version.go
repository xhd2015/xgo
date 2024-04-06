package main

import "fmt"

const VERSION = "1.0.18"
const REVISION = "2c003e4f895ef51b9c22e7c5bb5018dd6007ff96+1"
const NUMBER = 164

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
