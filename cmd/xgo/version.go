package main

import "fmt"

const VERSION = "1.0.36"
const REVISION = "b1fa6d6f3a19df8888bf2c0eb103ddff88257582+1"
const NUMBER = 226

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
