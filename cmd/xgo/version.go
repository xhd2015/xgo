package main

import "fmt"

const VERSION = "1.0.30"
const REVISION = "a6ed1f5d4adc882e06ad3da0530ed5e3733c6169+1"
const NUMBER = 208

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
