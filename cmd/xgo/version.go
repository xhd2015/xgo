package main

import "fmt"

const VERSION = "1.0.37"
const REVISION = "c60db693248f3308d29cd43a95b71de835346d50+1"
const NUMBER = 237

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
