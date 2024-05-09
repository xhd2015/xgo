package main

import "fmt"

const VERSION = "1.0.31"
const REVISION = "12ae038b4e052f42cd97af0721488c1b5f77eb6e+1"
const NUMBER = 210

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
