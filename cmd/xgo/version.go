package main

import "fmt"

const VERSION = "1.0.39"
const REVISION = "b2c3ae1e20700a8710d719b492b6db9e60fc4800+1"
const NUMBER = 263

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
