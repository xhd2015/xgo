package main

import "fmt"

const VERSION = "1.0.29"
const REVISION = "81a667d545fbd097e4403650a69af2c8477814fc+1"
const NUMBER = 206

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
