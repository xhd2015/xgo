package main

import "fmt"

const VERSION = "1.0.32"
const REVISION = "43f87cc805c73ee7a651255963b18979e84d8429+1"
const NUMBER = 213

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
