package main

import "fmt"

const VERSION = "1.0.25"
const REVISION = "ae8695c56e8c7f1976d409c7fba953041983c299+1"
const NUMBER = 187

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
