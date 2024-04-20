package main

import "fmt"

const VERSION = "1.0.25"
const REVISION = "d19b96030c13286e9410580ddec787e7d37cfec2+1"
const NUMBER = 194

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
