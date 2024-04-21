package main

import "fmt"

const VERSION = "1.0.27"
const REVISION = "4f58227acb703c893cc78dd39c09f0fe5d15565c+1"
const NUMBER = 200

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
