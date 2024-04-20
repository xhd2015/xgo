package main

import "fmt"

const VERSION = "1.0.25"
const REVISION = "9337223b46e108aefcf60d80512764558965983b+1"
const NUMBER = 191

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
