package main

import "fmt"

const VERSION = "1.0.35"
const REVISION = "c490c2dbaaecd04df3888542f830cb715dd9b285+1"
const NUMBER = 220

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
