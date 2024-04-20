package main

import "fmt"

const VERSION = "1.0.25"
const REVISION = "5becdd180bf57fd65a7579d3760f700a777c35c0+1"
const NUMBER = 193

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
